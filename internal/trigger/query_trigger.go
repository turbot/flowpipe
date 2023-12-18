package trigger

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type TriggerRunnerQuery struct {
	TriggerRunnerBase
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

	primaryKeys := []string{}
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
		primaryKeys = append(primaryKeys, pkString)
	}

	slog.Info("Output", "primaryKeys", primaryKeys)

	safeTriggerName := strings.ReplaceAll(tr.Trigger.FullName, ".", "_")

	db, err := InitializeDB("./test.db", safeTriggerName)
	if err != nil {
		slog.Error("Error initializing db", "error", err)
		return
	}
	newItemPrimaryKeys, err := StoreSlice(db, safeTriggerName, primaryKeys)
	if err != nil {
		slog.Error("Error storing slice", "error", err)
		return
	}

	for _, k := range newItemPrimaryKeys {
		slog.Info("New item key", "key", k)
		row := primaryKeyRowMap[k]
		slog.Info("New item rows", "row", row)
	}
}

func InitializeDB(dbPath, tableName string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (data TEXT);`

	slog.Info("Creating table", "sql", createTableSQL)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	crateIndexSQL := `CREATE INDEX IF NOT EXISTS idx_data ON ` + tableName + ` (data);`
	_, err = db.Exec(crateIndexSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func StoreSlice(db *sql.DB, tableName string, slice []string) ([]string, error) {

	if len(slice) == 0 {
		return nil, nil
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// Create a temporary table
	_, err = tx.Exec(`CREATE TEMPORARY TABLE ` + tableName + `_temp_items (data TEXT)`)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
			return nil, err
		}
		return nil, err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(`INSERT INTO ` + tableName + `_temp_items (data) VALUES (?)`)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
			return nil, err
		}
		return nil, err
	}
	defer tempStmt.Close()

	// Insert all items into the temporary table
	for _, item := range slice {
		_, err = tempStmt.Exec(item)
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
				return nil, err
			}
			return nil, err
		}
	}

	// Find new items by comparing with the main table
	newItemsSQL := `
        INSERT INTO ` + tableName + ` (data)
        SELECT data FROM ` + tableName + `_temp_items
        WHERE data NOT IN (SELECT data FROM ` + tableName + `)
        RETURNING data
    `
	rows, err := tx.Query(newItemsSQL)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	// Collect new items
	var newItems []string
	for rows.Next() {
		var newData string
		if err := rows.Scan(&newData); err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
				return nil, err
			}
			return nil, err
		}
		newItems = append(newItems, newData)
	}

	_, err = tx.Exec(`DROP TABLE ` + tableName + `_temp_items`)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
			return nil, err
		}
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return nil, err
	}

	return newItems, nil
}