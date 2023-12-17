package trigger

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"

	_ "github.com/mattn/go-sqlite3"
)

type TriggerRunnerBase struct {
	Trigger    *modconfig.Trigger
	commandBus handler.FpCommandBus
	rootMod    *modconfig.Mod
}

type TriggerRunner interface {
	Run()
}

func NewTriggerRunner(ctx context.Context, commandBus handler.FpCommandBus, rootMod *modconfig.Mod, trigger *modconfig.Trigger) TriggerRunner {
	switch trigger.Config.(type) {
	case *modconfig.TriggerSchedule, *modconfig.TriggerInterval:
		return &TriggerRunnerBase{
			Trigger:    trigger,
			commandBus: commandBus,
			rootMod:    rootMod,
		}
	case *modconfig.TriggerQuery:
		return &TriggerRunnerQuery{
			TriggerRunnerBase: TriggerRunnerBase{
				Trigger:    trigger,
				commandBus: commandBus,
				rootMod:    rootMod},
		}
	default:
		return nil
	}
}

func (tr *TriggerRunnerBase) Run() {
	pipeline := tr.Trigger.GetPipeline()

	if pipeline == cty.NilVal {
		slog.Error("Pipeline is nil, cannot run trigger", "trigger", tr.Trigger.Name())
		return
	}

	pipelineDefn := pipeline.AsValueMap()
	pipelineName := pipelineDefn["name"].AsString()

	modFullName := tr.Trigger.GetMetadata().ModFullName
	slog.Info("Running trigger", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "mod", modFullName)

	// We can only run trigger from root mod

	if modFullName != tr.rootMod.FullName {
		slog.Error("Trigger can only be run from root mod", "trigger", tr.Trigger.Name(), "mod", modFullName, "root_mod", tr.rootMod.FullName)
		return
	}

	vars := map[string]cty.Value{}
	for _, v := range tr.rootMod.ResourceMaps.Variables {
		vars[v.GetMetadata().ResourceName] = v.Value
	}

	executionVariables := map[string]cty.Value{}
	executionVariables[schema.AttributeVar] = cty.ObjectVal(vars)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(constants.ArgModLocation)),
	}

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

	slog.Info("Trigger fired", "trigger", tr.Trigger.Name(), "pipeline", pipelineName, "pipeline_execution_id", pipelineCmd.PipelineExecutionID)

	if err := tr.commandBus.Send(context.TODO(), pipelineCmd); err != nil {
		slog.Error("Error sending pipeline command", "error", err)
		return
	}
}

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

	primaryKeys := []interface{}{}
	for _, r := range rows {
		// get the primary key
		primaryKey := r[config.PrimaryKey]
		if primaryKey == nil {
			slog.Error("Primary key not found in row", "trigger", tr.Trigger.Name())
			return
		}
		primaryKeys = append(primaryKeys, primaryKey)
	}

	slog.Info("Output", "primaryKeys", primaryKeys)

	safeTriggerName := strings.ReplaceAll(tr.Trigger.FullName, ".", "_")

	db, err := InitializeDB("./test.db", safeTriggerName)
	if err != nil {
		slog.Error("Error initializing db", "error", err)
		return
	}
	_, err = StoreSlice(db, safeTriggerName, primaryKeys)
	if err != nil {
		slog.Error("Error storing slice", "error", err)
		return
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

func StoreSlice(db *sql.DB, tableName string, slice []interface{}) ([]interface{}, error) {
	var newItems []interface{}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// Create a temporary table
	_, err = tx.Exec(`CREATE TEMPORARY TABLE ` + tableName + `_temp_items (data TEXT)`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(`INSERT INTO ` + tableName + `_temp_items (data) VALUES (?)`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer tempStmt.Close()

	// Insert all items into the temporary table
	for _, item := range slice {
		jsonData, err := json.Marshal(item)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		_, err = tempStmt.Exec(string(jsonData))
		if err != nil {
			tx.Rollback()
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
	for rows.Next() {
		var jsonData string
		if err := rows.Scan(&jsonData); err != nil {
			tx.Rollback()
			return nil, err
		}
		var item interface{}
		if err := json.Unmarshal([]byte(jsonData), &item); err != nil {
			tx.Rollback()
			return nil, err
		}
		newItems = append(newItems, item)
	}

	_, err = tx.Exec(`DROP TABLE ` + tableName + `_temp_items`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return newItems, nil
}
