package trigger

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/turbot/flowpipe/internal/resources"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/cache"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func TestPostgresSqlTriggerQuery(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	connectionString := "postgres://flowpipe:password@localhost:5432/flowpipe-test?sslmode=disable"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	// delete flowpipe.db
	flowpipeDbFilename := filepaths.FlowpipeDBFileName()

	_, err = os.Stat(flowpipeDbFilename)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove(flowpipeDbFilename)
		if err != nil {
			panic(err)
		}
	}

	err = store.InitializeFlowpipeDB()
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}

	err = dropPostgresSqlTestTable(db, "test_one")
	if err != nil {
		assert.Fail("Error deleting test table", err)
		return
	}

	err = createPostgresSqlTestTableA(db, "test_one")
	if err != nil {
		assert.Fail("Error creating test table", err)
		return
	}

	data := []map[string]interface{}{
		{
			"id":                "1",
			"name":              "John",
			"age":               30,
			"registration_date": "2020-01-01",
			"is_active":         true,
		},
		{
			"id":                "2",
			"name":              "Jane",
			"age":               25,
			"registration_date": "2020-02-20",
			"is_active":         false,
		},
		{
			"id":                "3",
			"name":              "Joe",
			"age":               40,
			"registration_date": "2020-03-05",
			"is_active":         true,
		},
	}

	err = populatePostgresSqlTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// We just need a name for the pipeline
	insertPipelineMap := map[string]cty.Value{
		"name": cty.StringVal("insert_pipe"),
	}
	updatePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("update_pipe"),
	}
	deletePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("delete_pipe"),
	}

	var generatedEvalContext *hcl.EvalContext
	hclExpressionMock := &util.HclExpressionMock{
		ValueFunc: func(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
			generatedEvalContext = evalCtx
			res := map[string]cty.Value{
				"from": cty.StringVal("test"),
			}
			return cty.ObjectVal(res), nil
		},
	}

	trigger := &resources.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger",
		},
		ArgsRaw: hclExpressionMock,
	}

	// build the captures
	insertPipelineCty := cty.ObjectVal(insertPipelineMap)
	updatePipelineCty := cty.ObjectVal(updatePipelineMap)
	deletePipelineCty := cty.ObjectVal(deletePipelineMap)

	// TODO: args?
	insertCapture := &resources.TriggerQueryCapture{
		Type:     "insert",
		Pipeline: insertPipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	updateCapture := &resources.TriggerQueryCapture{
		Type:     "update",
		Pipeline: updatePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	deleteCapture := &resources.TriggerQueryCapture{
		Type:     "delete",
		Pipeline: deletePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &resources.TriggerQuery{
		Database:   connectionString,
		Sql:        "select * from test_one",
		PrimaryKey: "id",
		Captures: map[string]*resources.TriggerQueryCapture{
			"insert": insertCapture,
			"update": updateCapture,
			"delete": deleteCapture,
		},
	}

	// Add trigger to db
	cache.GetCache().SetWithTTL(trigger.Name(), trigger, 10*time.Minute)

	executionId := util.NewExecutionId()
	triggerExecutionId := util.NewTriggerExecutionId()

	triggerRunner := NewTriggerRunner(trigger, executionId, triggerExecutionId)

	pipelineQueues, err := triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar := generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap := selfVar.AsValueMap()
	insertedRows := selfVarMap["inserted_rows"]
	assert.NotEqual(cty.NilVal, insertedRows, "inserted rows should not be nil")

	insertedRowsList := insertedRows.AsValueSlice()
	assert.Equal(3, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "2" {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "3" {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	// check the triggerCommands .. we check how many pipeline is executed and which pipeline is executed
	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the insert pipeline should be executed")
	assert.Equal("insert_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// SECOND RUN
	//
	// Without changing anything, the second run should not have any new "inserted_rows"

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(0, len(pipelineQueues), "trigger command should be nil, since there's no change the pipeline should NOT be called")

	//
	// THIRD RUN
	//

	// Add a new rows to our test table
	data = []map[string]interface{}{
		{
			"id":                "4",
			"name":              "Jack",
			"age":               35,
			"registration_date": "2020-04-01",
			"is_active":         true,
		},
		{
			"id":                "5",
			"name":              "Jill",
			"age":               30,
			"registration_date": "2020-05-20",
			"is_active":         false,
		},
	}

	err = populatePostgresSqlTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar = generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap = selfVar.AsValueMap()
	insertedRows = selfVarMap["inserted_rows"]
	insertedRowsList = insertedRows.AsValueSlice()
	assert.Equal(2, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "4" {
			assert.Equal("Jack", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(35), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-04-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "5" {
			assert.Equal("Jill", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-05-20T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the insert pipeline should be executed")
	assert.Equal("insert_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// FOURTH RUN
	//
	// Test for update

	err = updatePostgresSqlTestTableA(db, "test_one", map[string]interface{}{
		"id":                "1",
		"name":              "John",
		"age":               35,
		"registration_date": "2020-01-01",
		"is_active":         false,
	})
	if err != nil {
		assert.Fail("Error updating test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar = generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap = selfVar.AsValueMap()
	insertedRows = selfVarMap["inserted_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), insertedRows, "inserted rows should be nil, there's no new addition detected by the query trigger")

	updatedRows := selfVarMap["updated_rows"]
	assert.NotNil(updatedRows, "updated rows should not be nil")
	updatedRowsList := updatedRows.AsValueSlice()

	assert.Equal(1, len(updatedRowsList), "wrong number of updated rows")
	for _, row := range updatedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(35), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
			return
		}
	}

	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the update pipeline should be executed")
	assert.Equal("update_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// FIFTH RUN
	//
	// run it again, shouldn't have any new updates

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(0, len(pipelineQueues), "no update")

	//
	// SIXTH RUN
	//
	// Delete some rows
	idsToDelete := []any{"1", "4"}
	err = deleteFromPostgresSqlTestTable(db, "test_one", idsToDelete)
	if err != nil {
		assert.Fail("Error deleting from test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar = generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap = selfVar.AsValueMap()
	insertedRows = selfVarMap["inserted_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), insertedRows, "inserted rows should be nil, there's no new addition detected by the query trigger")

	updatedRows = selfVarMap["updated_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys := selfVarMap["deleted_rows"]

	deletedKeyValueSlice := deletedKeys.AsValueSlice()
	assert.Equal(2, len(deletedKeyValueSlice), "wrong number of deleted keys")

	for _, deletedKey := range deletedKeyValueSlice {
		deletedKeyString := deletedKey.AsString()
		if deletedKeyString != "1" && deletedKeyString != "4" {
			assert.Fail("wrong deleted key")
			return
		}
	}

	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the delete pipeline should be executed")
	assert.Equal("delete_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// SEVENTH RUN
	//
	// Multiple pipeline get executed
	data = []map[string]interface{}{
		{
			"id":                "6",
			"name":              "Jack",
			"age":               35,
			"registration_date": "2020-04-01",
			"is_active":         true,
		},
		{
			"id":                "7",
			"name":              "Jill",
			"age":               30,
			"registration_date": "2020-05-20",
			"is_active":         false,
		},
	}

	err = populatePostgresSqlTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	idsToDelete = []any{"2"}
	err = deleteFromPostgresSqlTestTable(db, "test_one", idsToDelete)
	if err != nil {
		assert.Fail("Error deleting from test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(2, len(pipelineQueues), "wrong number of trigger commands only the delete pipeline should be executed")
	assert.Contains([]string{"delete_pipe", "insert_pipe"}, pipelineQueues[0].Name, "wrong pipeline name")
	assert.Contains([]string{"delete_pipe", "insert_pipe"}, pipelineQueues[1].Name, "wrong pipeline name")
	assert.False(pipelineQueues[0].Name == pipelineQueues[1].Name, "ensure that we don't call insert_pipe twice or delete_pipe twice")
}

func TestPostgresSqlTriggerQueryNoPrimaryKey(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	connectionString := "postgres://flowpipe:password@localhost:5432/flowpipe-test?sslmode=disable"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	// delete flowpipe.db
	flowpipeDbFilename := filepaths.FlowpipeDBFileName()

	_, err = os.Stat(flowpipeDbFilename)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove(flowpipeDbFilename)
		if err != nil {
			panic(err)
		}
	}

	err = store.InitializeFlowpipeDB()
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}

	err = dropPostgresSqlTestTable(db, "test_one")
	if err != nil {
		assert.Fail("Error deleting test table", err)
		return
	}

	err = createPostgresSqlTestTableA(db, "test_one")
	if err != nil {
		assert.Fail("Error creating test table", err)
		return
	}

	data := []map[string]interface{}{
		{
			"id":                "1",
			"name":              "John",
			"age":               30,
			"registration_date": "2020-01-01",
			"is_active":         true,
		},
		{
			"id":                "2",
			"name":              "Jane",
			"age":               25,
			"registration_date": "2020-02-20",
			"is_active":         false,
		},
		{
			"id":                "3",
			"name":              "Joe",
			"age":               40,
			"registration_date": "2020-03-05",
			"is_active":         true,
		},
	}

	err = populatePostgresSqlTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// We just need a name for the pipeline
	insertPipelineMap := map[string]cty.Value{
		"name": cty.StringVal("insert_pipe"),
	}
	updatePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("update_pipe"),
	}
	deletePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("delete_pipe"),
	}

	var generatedEvalContext *hcl.EvalContext
	hclExpressionMock := &util.HclExpressionMock{
		ValueFunc: func(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
			generatedEvalContext = evalCtx
			res := map[string]cty.Value{
				"from": cty.StringVal("test"),
			}
			return cty.ObjectVal(res), nil
		},
	}

	trigger := &resources.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger",
		},
		ArgsRaw: hclExpressionMock,
	}

	// build the captures
	insertPipelineCty := cty.ObjectVal(insertPipelineMap)
	updatePipelineCty := cty.ObjectVal(updatePipelineMap)
	deletePipelineCty := cty.ObjectVal(deletePipelineMap)

	// TODO: args?
	insertCapture := &resources.TriggerQueryCapture{
		Type:     "insert",
		Pipeline: insertPipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	updateCapture := &resources.TriggerQueryCapture{
		Type:     "update",
		Pipeline: updatePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	deleteCapture := &resources.TriggerQueryCapture{
		Type:     "delete",
		Pipeline: deletePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &resources.TriggerQuery{
		Database: connectionString,
		Sql:      "select * from test_one",
		Captures: map[string]*resources.TriggerQueryCapture{
			"insert": insertCapture,
			"update": updateCapture,
			"delete": deleteCapture,
		},
	}

	// Add trigger to db
	cache.GetCache().SetWithTTL(trigger.Name(), trigger, 10*time.Minute)

	executionId := util.NewExecutionId()
	triggerExecutionId := util.NewTriggerExecutionId()

	triggerRunner := NewTriggerRunner(trigger, executionId, triggerExecutionId)

	pipelineQueues, err := triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar := generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap := selfVar.AsValueMap()
	insertedRows := selfVarMap["inserted_rows"]
	assert.NotEqual(cty.NilVal, insertedRows, "inserted rows should not be nil")

	insertedRowsList := insertedRows.AsValueSlice()
	assert.Equal(3, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "2" {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "3" {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the insert pipeline should be executed")
	assert.Equal("insert_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// SECOND RUN
	//
	// Without changing anything, the second run should not have any new "inserted_rows"

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(0, len(pipelineQueues), "trigger command should be nil, since there's no change the pipeline should NOT be called")

	//
	// THIRD RUN
	//

	// Add a new rows to our test table
	data = []map[string]interface{}{
		{
			"id":                "4",
			"name":              "Jack",
			"age":               35,
			"registration_date": "2020-04-01",
			"is_active":         true,
		},
		{
			"id":                "5",
			"name":              "Jill",
			"age":               30,
			"registration_date": "2020-05-20",
			"is_active":         false,
		},
	}

	err = populatePostgresSqlTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar = generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap = selfVar.AsValueMap()

	updatedRows := selfVarMap["updated_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys := selfVarMap["deleted_rows"]
	assert.Equal(cty.ListValEmpty(cty.String), deletedKeys, "deleted keys should be nil, there's no new deletion detected by the query trigger")

	insertedRows = selfVarMap["inserted_rows"]
	insertedRowsList = insertedRows.AsValueSlice()
	assert.Equal(2, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "4" {
			assert.Equal("Jack", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(35), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-04-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "5" {
			assert.Equal("Jill", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-05-20T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the insert pipeline should be executed")
	assert.Equal("insert_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// FOURTH RUN
	//
	// Update doesn't work with no primary key

	err = updatePostgresSqlTestTableA(db, "test_one", map[string]interface{}{
		"id":                "1",
		"name":              "John",
		"age":               35,
		"registration_date": "2020-01-01",
		"is_active":         false,
	})
	if err != nil {
		assert.Fail("Error updating test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	selfVar = generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap = selfVar.AsValueMap()
	insertedRows = selfVarMap["inserted_rows"]
	insertedRowsList = insertedRows.AsValueSlice()
	assert.Equal(1, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(35), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
			return
		}
	}

	updatedRows = selfVarMap["updated_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows does not work with no primary key query trigger")

	// The "updated row" is now a deleted entry
	deletedKeys = selfVarMap["deleted_rows"]
	deletedKeysSlice := deletedKeys.AsValueSlice()
	assert.Equal(1, len(deletedKeysSlice), "wrong number of deleted keys")
	// This Hash is generated in the function RunOne
	assert.Equal("0432ffd1257d9c9e9903528dc8367bed39b5aff0afe68662fd0afa4e85b20b56", deletedKeysSlice[0].AsString(), "wrong deleted key")

	// because update doesn't work without primary key, we have insert & delete instead
	assert.Equal(2, len(pipelineQueues), "wrong number of trigger commands only the delete pipeline should be executed")
	assert.Contains([]string{"delete_pipe", "insert_pipe"}, pipelineQueues[0].Name, "wrong pipeline name")
	assert.Contains([]string{"delete_pipe", "insert_pipe"}, pipelineQueues[1].Name, "wrong pipeline name")
	assert.False(pipelineQueues[0].Name == pipelineQueues[1].Name, "ensure that we don't call insert_pipe twice or delete_pipe twice")
}

func TestPostgresSqlTriggerQueryB(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	connectionString := "postgres://flowpipe:password@localhost:5432/flowpipe-test?sslmode=disable"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	// delete flowpipe.db
	flowpipeDbFilename := filepaths.FlowpipeDBFileName()

	_, err = os.Stat(flowpipeDbFilename)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove(flowpipeDbFilename)
		if err != nil {
			panic(err)
		}
	}

	err = store.InitializeFlowpipeDB()
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}

	err = dropPostgresSqlTestTable(db, "test_one")
	if err != nil {
		assert.Fail("Error deleting test table", err)
		return
	}

	err = createPostgresSqlTestTableB(db, "test_one")
	if err != nil {
		assert.Fail("Error creating test table", err)
		return
	}

	data := []map[string]interface{}{
		{
			"id":                1,
			"name":              "John",
			"age":               30,
			"registration_date": "2020-01-01",
			"is_active":         true,
		},
		{
			"id":                2,
			"name":              "Jane",
			"age":               25,
			"registration_date": "2020-02-20",
			"is_active":         false,
		},
		{
			"id":                3,
			"name":              "Joe",
			"age":               40,
			"registration_date": "2020-03-05",
			"is_active":         true,
		},
	}

	blobSizeMultiplier := 20
	for _, item := range data {
		blobData := make([]byte, 10*(blobSizeMultiplier+1))
		for i := range blobData {
			blobData[i] = byte(rand.Intn(256)) //nolint:gosec // just a test case
		}

		item["blob_data"] = blobData
	}

	err = populatePostgresSqlTestTableB(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// We just need a name for the pipeline
	insertPipelineMap := map[string]cty.Value{
		"name": cty.StringVal("insert_pipe"),
	}
	updatePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("update_pipe"),
	}
	deletePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("delete_pipe"),
	}

	var generatedEvalContext *hcl.EvalContext
	hclExpressionMock := &util.HclExpressionMock{
		ValueFunc: func(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
			generatedEvalContext = evalCtx
			res := map[string]cty.Value{
				"from": cty.StringVal("test"),
			}
			return cty.ObjectVal(res), nil
		},
	}

	trigger := &resources.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger_b",
		},
		ArgsRaw: hclExpressionMock,
	}

	// build the captures
	insertPipelineCty := cty.ObjectVal(insertPipelineMap)
	updatePipelineCty := cty.ObjectVal(updatePipelineMap)
	deletePipelineCty := cty.ObjectVal(deletePipelineMap)

	// TODO: args?
	insertCapture := &resources.TriggerQueryCapture{
		Type:     "insert",
		Pipeline: insertPipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	updateCapture := &resources.TriggerQueryCapture{
		Type:     "update",
		Pipeline: updatePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	deleteCapture := &resources.TriggerQueryCapture{
		Type:     "delete",
		Pipeline: deletePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &resources.TriggerQuery{
		Database:   connectionString,
		Sql:        "select * from test_one",
		PrimaryKey: "id",
		Captures: map[string]*resources.TriggerQueryCapture{
			"insert": insertCapture,
			"update": updateCapture,
			"delete": deleteCapture,
		},
	}

	// Add trigger to db
	cache.GetCache().SetWithTTL(trigger.Name(), trigger, 10*time.Minute)

	executionId := util.NewExecutionId()
	triggerExecutionId := util.NewTriggerExecutionId()

	triggerRunner := NewTriggerRunner(trigger, executionId, triggerExecutionId)

	pipelineQueues, err := triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}
	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar := generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap := selfVar.AsValueMap()
	insertedRows := selfVarMap["inserted_rows"]
	assert.NotEqual(cty.NilVal, insertedRows, "inserted rows should not be nil")

	insertedRowsList := insertedRows.AsValueSlice()
	assert.Equal(3, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()

		idGo, err := hclhelpers.CtyToGo(rowMap["id"])
		if err != nil {
			assert.Fail("Error converting id to go", err)
			return
		}

		idFloat, ok := idGo.(int)
		if !ok {
			assert.Fail("id should be float64")
			return
		}
		if idFloat == 1 {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if idFloat == 2 {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if idFloat == 3 {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	// check the triggerCommands .. we check how many pipeline is executed and which pipeline is executed
	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the insert pipeline should be executed")
	assert.Equal("insert_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	//
	// SECOND RUN
	//
	// No update

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(0, len(pipelineQueues), "trigger command should be nil, since there's no change the pipeline should NOT be called")

	//
	// THIRD RUN
	//
	// Update the blob data

	data[1]["blob_data"] = make([]byte, 10*(blobSizeMultiplier+2))
	for i := range data[1]["blob_data"].([]byte) {
		data[1]["blob_data"].([]byte)[i] = byte(rand.Intn(256)) //nolint:gosec // just a test case
	}

	err = updatePostgresSqlTestTableB(db, "test_one", data[1])
	if err != nil {
		assert.Fail("Error updating test table", err)
		return
	}

	// Reset
	generatedEvalContext = nil

	pipelineQueues, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(1, len(pipelineQueues), "wrong number of trigger commands only the update pipeline should be executed")
	assert.Equal("update_pipe", pipelineQueues[0].Name, "wrong pipeline name")

	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar = generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap = selfVar.AsValueMap()
	insertedRows = selfVarMap["inserted_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), insertedRows, "inserted rows should be nil, there's no new addition detected by the query trigger")

	updatedRows := selfVarMap["updated_rows"]
	assert.NotNil(updatedRows, "updated rows should not be nil")
	updatedRowsList := updatedRows.AsValueSlice()
	assert.Equal(1, len(updatedRowsList), "wrong number of updated rows")
	for _, row := range updatedRowsList {
		rowMap := row.AsValueMap()

		idGo, err := hclhelpers.CtyToGo(rowMap["id"])
		if err != nil {
			assert.Fail("Error converting id to go", err)
			return
		}

		idFloat, ok := idGo.(int)
		if !ok {
			assert.Fail("id should be float64")
			return
		}
		if idFloat == 2 {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")

			// Comparing blobData with otherSlice
			// TODO: byte array is converted to a number array in CTY. The conversion back to Go results in a float array. Perhaps we can improve this in the future, but for now I think we'll leave it at that.
		} else {
			assert.Fail("wrong id")
		}
	}
}

func TestPostgresSqlTriggerQueryBCustomCapture(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	connectionString := "postgres://flowpipe:password@localhost:5432/flowpipe-test?sslmode=disable"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	// delete flowpipe.db
	flowpipeDbFilename := filepaths.FlowpipeDBFileName()

	_, err = os.Stat(flowpipeDbFilename)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove(flowpipeDbFilename)
		if err != nil {
			panic(err)
		}
	}

	err = store.InitializeFlowpipeDB()
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}

	err = dropPostgresSqlTestTable(db, "test_one")
	if err != nil {
		assert.Fail("Error deleting test table", err)
		return
	}

	err = createPostgresSqlTestTableB(db, "test_one")
	if err != nil {
		assert.Fail("Error creating test table", err)
		return
	}

	data := []map[string]interface{}{
		{
			"id":                1,
			"name":              "John",
			"age":               30,
			"registration_date": "2020-01-01",
			"is_active":         true,
		},
		{
			"id":                2,
			"name":              "Jane",
			"age":               25,
			"registration_date": "2020-02-20",
			"is_active":         false,
		},
		{
			"id":                3,
			"name":              "Joe",
			"age":               40,
			"registration_date": "2020-03-05",
			"is_active":         true,
		},
	}

	blobSizeMultiplier := 20
	for _, item := range data {
		blobData := make([]byte, 10*(blobSizeMultiplier+1))
		for i := range blobData {
			blobData[i] = byte(rand.Intn(256)) //nolint:gosec // just a test case
		}

		item["blob_data"] = blobData
	}

	err = populatePostgresSqlTestTableB(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// We just need a name for the pipeline
	updatePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("update_pipe"),
	}
	deletePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("delete_pipe"),
	}

	hclExpressionMock := &util.HclExpressionMock{
		ValueFunc: func(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
			res := map[string]cty.Value{
				"from": cty.StringVal("test"),
			}
			return cty.ObjectVal(res), nil
		},
	}

	trigger := &resources.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger_b",
		},
		ArgsRaw: hclExpressionMock,
	}

	// build the captures
	updatePipelineCty := cty.ObjectVal(updatePipelineMap)
	deletePipelineCty := cty.ObjectVal(deletePipelineMap)

	// TODO: args?
	updateCapture := &resources.TriggerQueryCapture{
		Type:     "update",
		Pipeline: updatePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	deleteCapture := &resources.TriggerQueryCapture{
		Type:     "delete",
		Pipeline: deletePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &resources.TriggerQuery{
		Database:   connectionString,
		Sql:        "select * from test_one",
		PrimaryKey: "id",
		Captures: map[string]*resources.TriggerQueryCapture{
			"update": updateCapture,
			"delete": deleteCapture,
		},
	}

	// Add trigger to db
	cache.GetCache().SetWithTTL(trigger.Name(), trigger, 10*time.Minute)

	executionId := util.NewExecutionId()
	triggerExecutionId := util.NewTriggerExecutionId()

	triggerRunner := NewTriggerRunner(trigger, executionId, triggerExecutionId)

	pipelineQueues, err := triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}

	assert.Equal(0, len(pipelineQueues), "insert capture not defined no pipeline should be executed")
}

func TestPostgresSqlTriggerQueryWithNull(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	connectionString := "postgres://flowpipe:password@localhost:5432/flowpipe-test?sslmode=disable"

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	// delete flowpipe.db
	flowpipeDbFilename := filepaths.FlowpipeDBFileName()

	_, err = os.Stat(flowpipeDbFilename)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove(flowpipeDbFilename)
		if err != nil {
			panic(err)
		}
	}

	err = store.InitializeFlowpipeDB()
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}

	err = dropPostgresSqlTestTable(db, "test_one")
	if err != nil {
		assert.Fail("Error deleting test table", err)
		return
	}

	err = createPostgresSqlTestTableA(db, "test_one")
	if err != nil {
		assert.Fail("Error creating test table", err)
		return
	}

	data := []map[string]interface{}{
		{
			"id":                "1",
			"name":              "John",
			"age":               30,
			"registration_date": "2020-01-01",
			"is_active":         true,
		},
		{
			"id":                "2",
			"name":              "Jane",
			"age":               25,
			"registration_date": nil,
			"is_active":         false,
		},
		{
			"id":                "3",
			"name":              "Joe",
			"age":               40,
			"registration_date": "2020-03-05",
			"is_active":         true,
		},
	}

	err = populatePostgresSqlTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	// We just need a name for the pipeline
	insertPipelineMap := map[string]cty.Value{
		"name": cty.StringVal("insert_pipe"),
	}
	updatePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("update_pipe"),
	}
	deletePipelineMap := map[string]cty.Value{
		"name": cty.StringVal("delete_pipe"),
	}

	var generatedEvalContext *hcl.EvalContext
	hclExpressionMock := &util.HclExpressionMock{
		ValueFunc: func(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
			generatedEvalContext = evalCtx
			res := map[string]cty.Value{
				"from": cty.StringVal("test"),
			}
			return cty.ObjectVal(res), nil
		},
	}

	trigger := &resources.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger",
		},
		ArgsRaw: hclExpressionMock,
	}

	// build the captures
	insertPipelineCty := cty.ObjectVal(insertPipelineMap)
	updatePipelineCty := cty.ObjectVal(updatePipelineMap)
	deletePipelineCty := cty.ObjectVal(deletePipelineMap)

	insertCapture := &resources.TriggerQueryCapture{
		Type:     "insert",
		Pipeline: insertPipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	updateCapture := &resources.TriggerQueryCapture{
		Type:     "update",
		Pipeline: updatePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}
	deleteCapture := &resources.TriggerQueryCapture{
		Type:     "delete",
		Pipeline: deletePipelineCty,
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &resources.TriggerQuery{
		Database:   connectionString,
		Sql:        "select * from test_one",
		PrimaryKey: "id",
		Captures: map[string]*resources.TriggerQueryCapture{
			"insert": insertCapture,
			"update": updateCapture,
			"delete": deleteCapture,
		},
	}

	// Add trigger to db
	cache.GetCache().SetWithTTL(trigger.Name(), trigger, 10*time.Minute)

	executionId := util.NewExecutionId()
	triggerExecutionId := util.NewTriggerExecutionId()

	triggerRunner := NewTriggerRunner(trigger, executionId, triggerExecutionId)

	_, err = triggerRunner.GetPipelineQueuesWithArgs(ctx, nil, nil)
	if err != nil {
		assert.Fail("Error executing trigger", err)
		return
	}
	// The callback to the mocks should have been called by now
	if generatedEvalContext == nil {
		assert.Fail("generated eval context should not be nil")
		return
	}

	selfVar := generatedEvalContext.Variables["self"]
	if selfVar == cty.NilVal {
		assert.Fail("self variable should not be nil")
		return
	}

	selfVarMap := selfVar.AsValueMap()
	insertedRows := selfVarMap["inserted_rows"]
	assert.NotEqual(cty.NilVal, insertedRows, "inserted rows should not be nil")

	insertedRowsList := insertedRows.AsValueSlice()
	assert.Equal(3, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "2" {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal(cty.NilVal, rowMap["registration_date"], "wrong registration date, registration for id 2 should be a nil value")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "3" {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00.000Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}
}

func createPostgresSqlTestTableA(db *sql.DB, tableName string) error {
	createTableSQL := `create table if not exists ` + tableName + ` (id text primary key, name text, age integer, registration_date date, is_active boolean);`

	slog.Info("Creating table", "sql", createTableSQL)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func dropPostgresSqlTestTable(db *sql.DB, tableName string) error {
	dropTableSql := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)

	slog.Info("Deleting table", "sql", dropTableSql)
	_, err := db.Exec(dropTableSql)
	if err != nil {
		return err
	}

	return nil
}

func populatePostgresSqlTestTableA(db *sql.DB, tableName string, data []map[string]interface{}) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(`INSERT INTO ` + tableName + ` (id, name, age, registration_date, is_active) VALUES ($1, $2, $3, $4, $5)`) //nolint:gosec // should be safe to use
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
			return err
		}
		return err
	}
	defer tempStmt.Close()

	// Insert all items into the temporary table
	for _, item := range data {
		_, err = tempStmt.Exec(item["id"], item["name"], item["age"], item["registration_date"], item["is_active"])
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
				return err
			}
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}

	return nil
}

func updatePostgresSqlTestTableA(db *sql.DB, tableName string, data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	tempStmt, err := tx.Prepare(`UPDATE ` + tableName + ` SET name = $1, age = $2, registration_date = $3, is_active = $4 WHERE id = $5`) //nolint:gosec // should be safe to use
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return err
	}
	defer tempStmt.Close()

	_, err = tempStmt.Exec(data["name"], data["age"], data["registration_date"], data["is_active"], data["id"])
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}

	return nil
}

func deleteFromPostgresSqlTestTable(db *sql.DB, tableName string, idsToDelete []any) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Generate placeholders for each ID in the format PostgreSQL expects
	placeholders := make([]string, len(idsToDelete))
	for i := range idsToDelete {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholdersStr := strings.Join(placeholders, ", ")

	query := fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)", tableName, placeholdersStr) //nolint:gosec // should be safe to use
	tempStmt, err := tx.Prepare(query)
	if err != nil {
		// Roll back the transaction on error
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
			return err
		}
	}
	defer tempStmt.Close()

	_, err = tempStmt.Exec(idsToDelete...)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}

	return nil
}

func createPostgresSqlTestTableB(db *sql.DB, tableName string) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
		id INTEGER PRIMARY KEY,  -- Changed to INTEGER and is the primary key
		name TEXT,
		age INTEGER,
		registration_date DATE,
		is_active BOOLEAN,
		blob_data BYTEA          -- for storing BLOB data
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func populatePostgresSqlTestTableB(db *sql.DB, tableName string, data []map[string]interface{}) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(fmt.Sprintf(`INSERT INTO %s (id, name, age, registration_date, is_active, blob_data) VALUES ($1, $2, $3, $4, $5, $6)`, tableName))
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
			return err
		}
		return err
	}
	defer tempStmt.Close()

	// Insert all items into the temporary table
	for _, item := range data {
		_, err = tempStmt.Exec(item["id"], item["name"], item["age"], item["registration_date"], item["is_active"], item["blob_data"])
		if err != nil {
			err2 := tx.Rollback()
			if err2 != nil {
				slog.Error("Error rolling back transaction", "error", err2)
				return err
			}
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}

	return nil
}

func updatePostgresSqlTestTableB(db *sql.DB, tableName string, data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmtText := fmt.Sprintf(`UPDATE %s SET name = $1, age = $2, registration_date = $3, is_active = $4, blob_data = $5 WHERE id = $6`, tableName) //nolint:gosec // should be safe to use
	tempStmt, err := tx.Prepare(stmtText)
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return err
	}
	defer tempStmt.Close()

	_, err = tempStmt.Exec(data["name"], data["age"], data["registration_date"], data["is_active"], data["blob_data"], data["id"])
	if err != nil {
		err2 := tx.Rollback()
		if err2 != nil {
			slog.Error("Error rolling back transaction", "error", err2)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		slog.Error("Error committing transaction", "error", err)
		return err
	}

	return nil
}
