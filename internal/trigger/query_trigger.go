package trigger

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strings"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunnerQuery struct {
	TriggerRunnerBase
	DatabasePath string
}

type queryTriggerMetadata struct {
	PrimaryKey string
	RowHash    string
}

// hashRow generates a hash for the given row, properly handling blob data.
func hashRow(row map[string]interface{}) string {
	// Sort the keys to ensure consistent ordering
	var keys []string
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Initialize a hash writer
	hasher := sha256.New()

	// Process each key-value pair
	for _, k := range keys {
		value := row[k]

		// Check if the value is a slice of bytes (blob data)
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			slice, ok := value.([]byte)
			if ok {
				// Write the raw bytes directly to the hasher
				hasher.Write(slice)
				continue
			}
		}

		// For other data types, use fmt.Sprintf to convert them to strings
		hasher.Write([]byte(fmt.Sprintf("%v=%v;", k, value)))
	}

	// Compute the hash
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string
	return hex.EncodeToString(hashBytes)
}

// TODO: ensure only 1 trigger query is running at any given time
func (tr *TriggerRunnerQuery) Run() {
	pipeline := tr.Trigger.GetPipeline()

	if pipeline == cty.NilVal {
		slog.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	slog.Info("Running trigger", "trigger", tr.Trigger.Name(), "pipeline", pipelineName)

	config := tr.Trigger.Config.(*modconfig.TriggerQuery)

	queryPrimitive := primitive.Query{}

	input := modconfig.Input{
		schema.AttributeTypeSql:              config.Sql,
		schema.AttributeTypeConnectionString: config.ConnectionString,
	}

	output, err := queryPrimitive.Run(context.Background(), input)
	if err != nil {
		slog.Error("Error running trigger query", "error", err)
		return
	}

	if output.Data["rows"] == nil {
		slog.Info("No rows returned from trigger query", "trigger", tr.Trigger.Name())
		return
	}

	rows, ok := output.Data["rows"].([]map[string]interface{})
	if !ok {
		slog.Error("Error converting rows to []interface{}", "trigger", tr.Trigger.Name())
		return
	}

	controlItems := []queryTriggerMetadata{}

	primaryKeyRowMap := map[string]interface{}{}
	for _, r := range rows {
		// get the primary key
		primaryKey := r[config.PrimaryKey]
		if primaryKey == nil {
			slog.Error("Primary key not found in row", "trigger", tr.Trigger.Name())
			return
		}
		pkString, ok := primaryKey.(string)
		if !ok {
			pkString = fmt.Sprintf("%v", primaryKey)
		}

		primaryKeyRowMap[pkString] = r

		rowHash := hashRow(r)

		controlItem := queryTriggerMetadata{
			PrimaryKey: pkString,
			RowHash:    rowHash,
		}
		controlItems = append(controlItems, controlItem)
	}

	safeTriggerName := strings.ReplaceAll(tr.Trigger.FullName, ".", "_")

	db, err := initializeQueryTriggerDB(tr.DatabasePath, safeTriggerName)
	if err != nil {
		slog.Error("Error initializing db", "error", err)
		return
	}
	defer db.Close()

	newItemPrimaryKeys, updatedItemPrimaryKeys, deletedPrimaryKeys, err := calculatedNewUpdatedDeletedData(db, safeTriggerName, controlItems)
	if err != nil {
		slog.Error("Error storing slice", "error", err)
		return
	}

	newRows := []map[string]interface{}{}
	for _, k := range newItemPrimaryKeys {
		slog.Debug("New item key", "key", k)
		row := primaryKeyRowMap[k]

		if row == nil {
			slog.Warn("New item not found in row map", "key", k)
			continue
		}

		slog.Debug("New item rows", "row", row)
		newRows = append(newRows, row.(map[string]interface{}))
	}

	newRowCtyVals, err := newRowsCty(newRows)
	if err != nil {
		slog.Error("Error building new rows cty", "error", err)
		return
	}

	updatedRows := []map[string]interface{}{}
	for _, k := range updatedItemPrimaryKeys {
		slog.Debug("New item key", "key", k)
		row := primaryKeyRowMap[k]

		if row == nil {
			slog.Warn("New item not found in row map", "key", k)
			continue
		}

		slog.Debug("New item rows", "row", row)
		updatedRows = append(updatedRows, row.(map[string]interface{}))
	}

	updatedRowCtyVals, err := newRowsCty(updatedRows)
	if err != nil {
		slog.Error("Error building new rows cty", "error", err)
		return
	}

	deletedKeysCty := []cty.Value{}
	for _, k := range deletedPrimaryKeys {
		deletedKeysCty = append(deletedKeysCty, cty.StringVal(k))
	}

	// Add the new rows to the pipeline args
	selfVars := map[string]cty.Value{}
	if len(newRowCtyVals) > 0 {
		selfVars["inserted_rows"] = cty.ListVal(newRowCtyVals)
	} else {
		selfVars["inserted_rows"] = cty.ListValEmpty(cty.DynamicPseudoType)
	}

	if len(updatedRowCtyVals) > 0 {
		selfVars["updated_rows"] = cty.ListVal(updatedRowCtyVals)
	} else {
		selfVars["updated_rows"] = cty.ListValEmpty(cty.DynamicPseudoType)
	}

	if len(deletedKeysCty) > 0 {
		selfVars["deleted_keys"] = cty.ListVal(deletedKeysCty)
	} else {
		selfVars["deleted_keys"] = cty.ListValEmpty(cty.String)
	}

	evalContext, err := buildEvalContext(tr.rootMod)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return
	}

	varsEvalContext := evalContext.Variables
	varsEvalContext["self"] = cty.ObjectVal(selfVars)

	pipelineArgs, diags := tr.Trigger.GetArgs(evalContext)
	if diags.HasErrors() {
		slog.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
		Args:                pipelineArgs,
	}

	slog.Info("Trigger fired", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID, "args", pipelineArgs)

	if err := tr.commandBus.Send(context.TODO(), pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)
		return
	}
}

func newRowsCty(newRows []map[string]interface{}) ([]cty.Value, error) {
	var newRowsCty []cty.Value
	for _, r := range newRows {
		rowCty, err := newRowCty(r)
		if err != nil {
			return nil, err
		}
		newRowsCty = append(newRowsCty, rowCty)
	}
	return newRowsCty, nil
}

func newRowCty(row map[string]interface{}) (cty.Value, error) {
	rowCty := map[string]cty.Value{}
	for k, v := range row {
		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(v)
		if err != nil {
			return cty.NilVal, err
		}
		rowCty[k] = ctyVal
	}
	return cty.ObjectVal(rowCty), nil
}

func initializeDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initializeQueryTriggerDB(dbPath, tableName string) (*sql.DB, error) {

	db, err := initializeDB(dbPath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `create table if not exists ` + tableName + ` (_fp_data text primary key, _fp_hash text);`

	slog.Info("Creating table", "sql", createTableSQL)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	crateIndexSQL := `create index if not exists idx_data on ` + tableName + ` (_fp_data);`
	_, err = db.Exec(crateIndexSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func calculatedNewUpdatedDeletedData(db *sql.DB, tableName string, controlItems []queryTriggerMetadata) ([]string, []string, []string, error) {
	if len(controlItems) == 0 {
		return nil, nil, nil, nil
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, nil, err
	}

	// Create a temporary table
	_, err = tx.Exec(`create temporary table ` + tableName + `_temp_items (_fp_data text primary key, _fp_hash text)`)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(`insert into ` + tableName + `_temp_items (_fp_data, _fp_hash) values (?, ?)`) //nolint:gosec // should be safe to use
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}
	defer tempStmt.Close()

	// Insert all items into the temporary table
	for _, item := range controlItems {
		_, err = tempStmt.Exec(item.PrimaryKey, item.RowHash)
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
			}
			return nil, nil, nil, err
		}
	}

	newItems, err := insertNewItems(tx, tableName)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	updatedItems, err := updatedItems(tx, tableName)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	slog.Info("updatedItems", "updatedItems", updatedItems)

	// Find deleted items by comparing with the main table
	deletedItemsSQL := `
		select _fp_data from ` + tableName + `
		where _fp_data not in (select _fp_data from ` + tableName + `_temp_items)
	`

	deletedRows, err := tx.Query(deletedItemsSQL)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}
	defer deletedRows.Close()

	// Collect deleted items
	var deletedItems []string
	for deletedRows.Next() {
		var deletedData string
		if err := deletedRows.Scan(&deletedData); err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
			}
		}
		deletedItems = append(deletedItems, deletedData)
	}

	slog.Debug("deleted items found", "deletedItems", deletedItems)

	deleteItemsFromTrackingDb := `delete from ` + tableName + ` where _fp_data = ?`
	deleteStmt, err := tx.Prepare(deleteItemsFromTrackingDb)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}
	defer deleteStmt.Close()
	for _, deletedItem := range deletedItems {
		slog.Debug("Deleting item", "item", deletedItem, "table", tableName)
		_, err = deleteStmt.Exec(deletedItem)
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
			}
			return nil, nil, nil, err
		}
	}

	_, err = tx.Exec(`drop table ` + tableName + `_temp_items`)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return nil, nil, nil, err
	}

	return newItems, updatedItems, deletedItems, nil
}

func insertNewItems(tx *sql.Tx, tableName string) ([]string, error) {
	// Find new items by comparing with the main table
	newItemsSQL := `
        insert into ` + tableName + ` (_fp_data, _fp_hash)
        select _fp_data, _fp_hash from ` + tableName + `_temp_items
        where _fp_data not in (select _fp_data from ` + tableName + `)
        returning _fp_data
    `
	rows, err := tx.Query(newItemsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect new items
	var newItems []string
	for rows.Next() {
		var newData string
		if err := rows.Scan(&newData); err != nil {
			return nil, err
		}
		newItems = append(newItems, newData)
	}
	return newItems, nil
}

func updatedItems(tx *sql.Tx, tableName string) ([]string, error) {

	sourceTable := tableName + "_temp_items"

	updateItemsSQL := `WITH Updated AS (
		SELECT _fp_data
		FROM ` + tableName + `
		WHERE EXISTS (
			SELECT 1
			FROM ` + sourceTable + `
			WHERE ` + sourceTable + `._fp_data = ` + tableName + `._fp_data
			  AND ` + tableName + `._fp_hash <> ` + sourceTable + `._fp_hash
		)
	)
	UPDATE ` + tableName + `
	SET _fp_hash = (
		SELECT _fp_hash
		FROM ` + sourceTable + `
		WHERE ` + sourceTable + `._fp_data = ` + tableName + `._fp_data
	)
	WHERE _fp_data IN (SELECT _fp_data FROM Updated)
	RETURNING _fp_data;
	`

	rows, err := tx.Query(updateItemsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect new items
	var updatedItems []string
	for rows.Next() {
		var newData string
		if err := rows.Scan(&newData); err != nil {
			return nil, err
		}
		updatedItems = append(updatedItems, newData)
	}
	return updatedItems, nil
}
