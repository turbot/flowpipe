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
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/es/event"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunnerQuery struct {
	TriggerRunnerBase
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

		if helpers.IsNil(value) {
			continue
		}

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

func (tr *TriggerRunnerQuery) Run() {
	tr.Fqueue.Enqueue(tr.RunOne)
	tr.Fqueue.Execute()
}

func (tr *TriggerRunnerQuery) RunOne() error {

	slog.Info("Running trigger", "trigger", tr.Trigger.Name())

	config := tr.Trigger.Config.(*modconfig.TriggerQuery)

	queryPrimitive := primitive.Query{}

	input := modconfig.Input{
		schema.AttributeTypeSql:      config.Sql,
		schema.AttributeTypeDatabase: config.Database,
	}

	output, _, err := queryPrimitive.RunWithMetadata(context.Background(), input)
	if err != nil {
		slog.Error("Error running trigger query", "error", err)
		if o.IsServerMode {
			o.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error running query trigger "+tr.Trigger.Name(), err))
		}
		return err
	}

	if output.Data["rows"] == nil {
		slog.Info("No rows returned from trigger query", "trigger", tr.Trigger.Name())
		if o.IsServerMode {
			o.RenderServerOutput(context.TODO(), types.NewServerOutputQueryTriggerRun(tr.Trigger.Name(), 0, 0, 0))
		}
		return nil
	}

	rows, ok := output.Data["rows"].([]map[string]interface{})
	if !ok {
		slog.Error("Error converting rows to []interface{}", "trigger", tr.Trigger.Name())
		if o.IsServerMode {
			o.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error converting rows to []interface{} "+tr.Trigger.Name(), err))
		}
		return nil
	}

	controlItems := []queryTriggerMetadata{}

	primaryKeyRowMap := map[string]interface{}{}

	if config.PrimaryKey != "" {
		for _, r := range rows {
			// get the primary key
			primaryKey := r[config.PrimaryKey]
			if primaryKey == nil {
				slog.Error("Primary key not found in row", "trigger", tr.Trigger.Name())
				if o.IsServerMode {
					o.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), fmt.Sprintf("primary key %s not found in query row from query trigger %s", config.PrimaryKey, tr.Trigger.Name()), err))
				}
				return perr.InternalWithMessage("Primary key not found in row")
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
	} else {
		for _, r := range rows {
			rowHash := hashRow(r)
			// use the rowHash as the primary key
			primaryKeyRowMap[rowHash] = r

			controlItem := queryTriggerMetadata{
				PrimaryKey: rowHash,
				RowHash:    rowHash,
			}
			controlItems = append(controlItems, controlItem)
		}
	}

	safeTriggerName := strings.ReplaceAll(tr.Trigger.FullName, ".", "_")

	db, err := store.OpenFlowpipeDB()
	if err != nil {
		slog.Error("Error opening Flowpipe db", "error", err)
		return err
	}
	defer db.Close()

	newItemPrimaryKeys, updatedItemPrimaryKeys, deletedPrimaryKeys, err := calculatedNewUpdatedDeletedData(db, safeTriggerName, controlItems)
	if err != nil {
		slog.Error("Error storing slice", "error", err)
		return err
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

	newRowCtyVals, err := hclhelpers.ConvertInterfaceToCtyValue(newRows)
	if err != nil {
		slog.Error("Error building new rows cty", "error", err)
		return err
	}

	updatedRows := []map[string]interface{}{}
	for _, k := range updatedItemPrimaryKeys {
		slog.Debug("New item key", "key", k)
		row := primaryKeyRowMap[k]

		if row == nil {
			slog.Warn("New item not found in row map", "key", k)
			continue
		}

		slog.Debug("Updated item row", "row", row)
		updatedRows = append(updatedRows, row.(map[string]interface{}))
	}

	updatedRowCtyVals, err := hclhelpers.ConvertInterfaceToCtyValue(updatedRows)
	if err != nil {
		slog.Error("Error building updated rows cty", "error", err)
		return err
	}

	deletedKeysCty, err := hclhelpers.ConvertInterfaceToCtyValue(deletedPrimaryKeys)
	if err != nil {
		slog.Error("Error building deleted rows cty", "error", err)
		return err
	}

	evalContext, err := buildEvalContext(tr.rootMod)
	if err != nil {
		slog.Error("Error building eval context", "error", err)
		return err
	}

	// Add the new rows to the pipeline args
	selfVars := map[string]cty.Value{}

	if len(newRows) > 0 {
		selfVars["inserted_rows"] = newRowCtyVals
	} else {
		selfVars["inserted_rows"] = cty.ListValEmpty(cty.DynamicPseudoType)
	}
	if len(updatedRows) > 0 {
		selfVars["updated_rows"] = updatedRowCtyVals
	} else {
		selfVars["updated_rows"] = cty.ListValEmpty(cty.DynamicPseudoType)
	}

	if len(deletedPrimaryKeys) > 0 {
		selfVars["deleted_rows"] = deletedKeysCty
	} else {
		selfVars["deleted_rows"] = cty.ListValEmpty(cty.String)
	}

	varsEvalContext := evalContext.Variables
	varsEvalContext["self"] = cty.ObjectVal(selfVars)

	queryStat := map[string]int{
		"insert": len(newRows),
		"update": len(updatedRows),
		"delete": len(deletedPrimaryKeys),
	}
	if o.IsServerMode {
		o.RenderServerOutput(context.TODO(), types.NewServerOutputQueryTriggerRun(tr.Trigger.Name(), len(newRows), len(updatedRows), len(deletedPrimaryKeys)))
	}
	for _, capture := range config.Captures {
		err := runPipeline(capture, tr, evalContext, queryStat)
		if err != nil {
			slog.Error("Error running pipeline", "error", err)
			return err
		}
	}

	return nil
}

func runPipeline(capture *modconfig.TriggerQueryCapture, tr *TriggerRunnerQuery, evalContext *hcl.EvalContext, queryStat map[string]int) error {

	if queryStat[capture.Type] <= 0 {
		return nil
	}

	pipelineArgs, diags := capture.GetArgs(evalContext)
	if diags.HasErrors() {
		slog.Error("Error getting trigger args", "trigger", tr.Trigger.Name(), "errors", diags)
		return perr.InternalWithMessage("Error getting trigger args")
	}

	pipeline := capture.Pipeline

	if pipeline == cty.NilVal {
		slog.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return perr.BadRequestWithMessage("Pipeline is nil, cannot run trigger")
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
		Args:                pipelineArgs,
	}

	slog.Info("Trigger fired", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID, "args", pipelineArgs, "capture_type", capture.Type, "capture_count", queryStat[capture.Type])
	if o.IsServerMode {
		o.RenderServerOutput(context.TODO(), types.NewServerOutputTriggerExecution(time.Now(), pipelineCmd.Event.ExecutionID, tr.Trigger.Name(), pipelineName))
	}

	if err := tr.commandBus.Send(context.TODO(), pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)
		if o.IsServerMode {
			o.RenderServerOutput(context.TODO(), types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "error sending pipeline command", err))
		}
		return err
	}

	return nil
}

func calculatedNewUpdatedDeletedData(db *sql.DB, triggerName string, controlItems []queryTriggerMetadata) ([]string, []string, []string, error) {
	if len(controlItems) == 0 {
		return nil, nil, nil, nil
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, nil, err
	}

	// Create a temporary table
	_, err = tx.Exec(`create temporary table ` + triggerName + `_temp_items (primary_key text primary key, row_hash text, created_at text)`)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	// Prepare statement for inserting into the temporary table
	timeNow := time.Now()
	tempStmt, err := tx.Prepare(`insert into ` + triggerName + `_temp_items (primary_key, row_hash, created_at) values (?, ?, ?)`) //nolint:gosec // should be safe to use
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
		_, err = tempStmt.Exec(item.PrimaryKey, item.RowHash, timeNow.UTC().Format(util.RFC3389WithMS))
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
			}
			return nil, nil, nil, err
		}
	}

	// Insert the "new items" into our tracking table
	newItems, err := insertNewItems(tx, triggerName)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	updatedItems, err := updatedItems(tx, triggerName)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return nil, nil, nil, err
	}

	slog.Debug("updatedItems", "updatedItems", updatedItems)

	// Find deleted items by comparing with the main table
	//nolint:gosec // TODO: investigate string concat
	deletedItemsSQL := `select primary_key from query_trigger_captured_row
						where primary_key not in (select primary_key from ` + triggerName + `_temp_items) and trigger_name = ?`

	deletedRows, err := tx.Query(deletedItemsSQL, triggerName)
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

	deleteItemsFromTrackingDb := `delete from query_trigger_captured_row where primary_key = ? and trigger_name = ?`
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
		slog.Debug("Deleting item", "item", deletedItem, "table", triggerName)
		_, err = deleteStmt.Exec(deletedItem, triggerName)
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
			}
			return nil, nil, nil, err
		}
	}

	_, err = tx.Exec(`drop table ` + triggerName + `_temp_items`)
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

func insertNewItems(tx *sql.Tx, triggerName string) ([]string, error) {
	// Find new items by comparing with the main table
	//nolint:gosec // TODO: investigate string concat
	newItemsSQL := `
        insert into query_trigger_captured_row (trigger_name, primary_key, row_hash, created_at)
        select '` + triggerName + `', primary_key, row_hash, created_at from ` + triggerName + `_temp_items
        where primary_key not in (select primary_key from query_trigger_captured_row where trigger_name = ?)
        returning primary_key
    `
	rows, err := tx.Query(newItemsSQL, triggerName)
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

func updatedItems(tx *sql.Tx, triggerName string) ([]string, error) {

	sourceTable := triggerName + "_temp_items"

	timeNow := time.Now()
	//nolint:gosec // TODO: investigate string concat
	updateItemsSQL := `WITH Updated AS (
		SELECT primary_key
		FROM query_trigger_captured_row
		WHERE EXISTS (
			SELECT 1
			FROM ` + sourceTable + `
			WHERE ` + sourceTable + `.primary_key = query_trigger_captured_row.primary_key
			  AND ` + sourceTable + `.row_hash <>  query_trigger_captured_row.row_hash
		) AND trigger_name = '` + triggerName + `'
	)
	UPDATE query_trigger_captured_row
	SET row_hash = (
		SELECT row_hash
		FROM ` + sourceTable + `
		WHERE ` + sourceTable + `.primary_key = query_trigger_captured_row.primary_key
	), updated_at  = ?
	WHERE primary_key IN (SELECT primary_key FROM Updated) AND trigger_name = '` + triggerName + `'
	RETURNING primary_key;
	`

	rows, err := tx.Query(updateItemsSQL, timeNow.UTC().Format(util.RFC3389WithMS))
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
