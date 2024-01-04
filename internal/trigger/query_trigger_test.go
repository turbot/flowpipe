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

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func createTestTableA(db *sql.DB, tableName string) error {

	createTableSQL := `create table if not exists ` + tableName + ` (id text primary key, name text, age integer, registration_date date, is_active boolean);`

	slog.Info("Creating table", "sql", createTableSQL)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func populateTestTableA(db *sql.DB, tableName string, data []map[string]interface{}) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(`INSERT INTO ` + tableName + ` (id, name, age, registration_date, is_active) VALUES (?, ?, ?, ?, ?)`) //nolint:gosec // should be safe to use
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

func updateTestTableA(db *sql.DB, tableName string, data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	tempStmt, err := tx.Prepare(`UPDATE ` + tableName + ` SET name = ?, age = ?, registration_date = ?, is_active = ? WHERE id = ?`) //nolint:gosec // should be safe to use
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

func deleteFromTestTable(db *sql.DB, tableName string, idsToDelete []any) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare statement for inserting into the temporary table
	placeholders := strings.Join(strings.Split(strings.Repeat("?", len(idsToDelete)), ""), ",")

	tempStmt, err := tx.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id in (%s)", tableName, placeholders))
	if err != nil {
		return err
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

func createTestTableB(db *sql.DB, tableName string) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (
		id INTEGER PRIMARY KEY,  -- Changed to INTEGER and is the primary key
		name TEXT,
		age INTEGER,
		registration_date DATE,
		is_active BOOLEAN,
		blob_data BLOB       -- for storing BLOB data
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func populateTestTableB(db *sql.DB, tableName string, data []map[string]interface{}) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Prepare statement for inserting into the temporary table
	tempStmt, err := tx.Prepare(`INSERT INTO ` + tableName + ` (id, name, age, registration_date, is_active, blob_data) VALUES (?, ?, ?, ?, ?, ?)`) //nolint:gosec // should be safe to use
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

func updateTestTableB(db *sql.DB, tableName string, data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	tempStmt, err := tx.Prepare(`UPDATE ` + tableName + ` SET name = ?, age = ?, registration_date = ?, is_active = ?, blob_data = ? WHERE id = ?`) //nolint:gosec // should be safe to use
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

func TestTriggerQuery(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	outputPath := "./test_trigger_query.db"
	// Check if the directory exists
	_, err := os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	flowpipeDb := "./flowpipe.db"
	// Check if the directory exists
	_, err = os.Stat(flowpipeDb)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(flowpipeDb)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	db, err := initializeDB(outputPath)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	err = createTestTableA(db, "test_one")
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

	err = populateTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	pipeline := map[string]cty.Value{
		"name": cty.StringVal("test"),
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

	trigger := &modconfig.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger",
		},
		Pipeline: cty.ObjectVal(pipeline),
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &modconfig.TriggerQuery{
		ConnectionString: "sqlite:./test_trigger_query.db",
		Sql:              "select * from test_one",
		PrimaryKey:       "id",
	}

	var triggerCommand interface{}
	commandBusMock := &util.CommandBusMock{
		SendFunc: func(ctx context.Context, command interface{}) error {
			triggerCommand = command
			return nil
		},
	}

	triggerRunner := NewTriggerRunner(ctx, commandBusMock, nil, trigger)

	triggerRunnerQuery := triggerRunner.(*TriggerRunnerQuery)
	triggerRunnerQuery.DatabasePath = "./flowpipe.db"

	assert.NotNil(triggerRunner, "trigger runner should not be nil")

	receiveChannel := make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res := <-receiveChannel
	assert.Nil(res)

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
			assert.Equal("2020-01-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "2" {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "3" {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	//
	// SECOND RUN
	//
	// Without changing anything, the second run should not have any new "inserted_rows"
	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys := selfVarMap["deleted_keys"]
	assert.Equal(cty.ListValEmpty(cty.String), deletedKeys, "deleted keys should be nil, there's no new deletion detected by the query trigger")

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

	err = populateTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
			assert.Equal("2020-04-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "5" {
			assert.Equal("Jill", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-05-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	//
	// FOURTH RUN
	//
	// Test for update

	err = updateTestTableA(db, "test_one", map[string]interface{}{
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

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
	assert.NotNil(updatedRows, "updated rows should not be nil")
	updatedRowsList := updatedRows.AsValueSlice()

	assert.Equal(1, len(updatedRowsList), "wrong number of updated rows")
	for _, row := range updatedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(35), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
			return
		}
	}

	//
	// FOURTH RUN
	//
	// run it again, shouldn't have any new updates

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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

	//
	// FIFTH RUN
	//
	// Delete some rows
	idsToDelete := []any{"1", "4"}
	err = deleteFromTestTable(db, "test_one", idsToDelete)
	if err != nil {
		assert.Fail("Error deleting from test table", err)
		return
	}

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

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

	deletedKeys = selfVarMap["deleted_keys"]

	deletedKeyValueSlice := deletedKeys.AsValueSlice()
	assert.Equal(2, len(deletedKeyValueSlice), "wrong number of deleted keys")

	for _, deletedKey := range deletedKeyValueSlice {
		deletedKeyString := deletedKey.AsString()
		if deletedKeyString != "1" && deletedKeyString != "4" {
			assert.Fail("wrong deleted key")
			return
		}
	}
}

func TestTriggerQueryWithEventFilter(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	outputPath := "./test_trigger_query.db"
	// Check if the directory exists
	_, err := os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	flowpipeDb := "./flowpipe.db"
	// Check if the directory exists
	_, err = os.Stat(flowpipeDb)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(flowpipeDb)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	db, err := initializeDB(outputPath)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	err = createTestTableA(db, "test_one")
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

	err = populateTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	pipeline := map[string]cty.Value{
		"name": cty.StringVal("test"),
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

	trigger := &modconfig.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger",
		},
		Pipeline: cty.ObjectVal(pipeline),
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &modconfig.TriggerQuery{
		ConnectionString: "sqlite:./test_trigger_query.db",
		Sql:              "select * from test_one",
		PrimaryKey:       "id",
		Events:           []string{"insert", "update"},
	}

	var triggerCommand interface{}
	commandBusMock := &util.CommandBusMock{
		SendFunc: func(ctx context.Context, command interface{}) error {
			triggerCommand = command
			return nil
		},
	}

	triggerRunner := NewTriggerRunner(ctx, commandBusMock, nil, trigger)

	triggerRunnerQuery := triggerRunner.(*TriggerRunnerQuery)
	triggerRunnerQuery.DatabasePath = "./flowpipe.db"

	assert.NotNil(triggerRunner, "trigger runner should not be nil")

	receiveChannel := make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res := <-receiveChannel
	assert.Nil(res)

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
			assert.Equal("2020-01-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "2" {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "3" {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	//
	// Delete some rows
	idsToDelete := []any{"1", "3"}
	err = deleteFromTestTable(db, "test_one", idsToDelete)
	if err != nil {
		assert.Fail("Error deleting from test table", err)
		return
	}

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)
	assert.NotNil(triggerCommand, "trigger command should not be nil")

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
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys := selfVarMap["deleted_keys"]
	assert.Equal(cty.NilVal, deletedKeys, "deleted rows should be empty since we only listen to insert and update events")
}

func TestTriggerQueryNoPrimaryKey(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	outputPath := "./test_trigger_query.db"
	// Check if the directory exists
	_, err := os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	flowpipeDb := "./flowpipe.db"
	// Check if the directory exists
	_, err = os.Stat(flowpipeDb)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(flowpipeDb)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	db, err := initializeDB(outputPath)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	err = createTestTableA(db, "test_one")
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

	err = populateTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	pipeline := map[string]cty.Value{
		"name": cty.StringVal("test"),
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

	trigger := &modconfig.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger",
		},
		Pipeline: cty.ObjectVal(pipeline),
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &modconfig.TriggerQuery{
		ConnectionString: "sqlite:./test_trigger_query.db",
		Sql:              "select * from test_one",
	}

	var triggerCommand interface{}
	commandBusMock := &util.CommandBusMock{
		SendFunc: func(ctx context.Context, command interface{}) error {
			triggerCommand = command
			return nil
		},
	}

	triggerRunner := NewTriggerRunner(ctx, commandBusMock, nil, trigger)

	triggerRunnerQuery := triggerRunner.(*TriggerRunnerQuery)
	triggerRunnerQuery.DatabasePath = "./flowpipe.db"

	assert.NotNil(triggerRunner, "trigger runner should not be nil")

	receiveChannel := make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res := <-receiveChannel
	assert.Nil(res)

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
			assert.Equal("2020-01-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "2" {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "3" {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	//
	// SECOND RUN
	//
	// Without changing anything, the second run should not have any new "inserted_rows"
	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys := selfVarMap["deleted_keys"]
	assert.Equal(cty.ListValEmpty(cty.String), deletedKeys, "deleted keys should be nil, there's no new deletion detected by the query trigger")

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

	err = populateTestTableA(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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

	updatedRows = selfVarMap["updated_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys = selfVarMap["deleted_keys"]
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
			assert.Equal("2020-04-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if id == "5" {
			assert.Equal("Jill", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(30), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-05-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	//
	// FOURTH RUN
	//
	// Update doesn't work with no primary key

	err = updateTestTableA(db, "test_one", map[string]interface{}{
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

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
	assert.Equal(1, len(insertedRowsList), "wrong number of inserted rows")
	for _, row := range insertedRowsList {
		rowMap := row.AsValueMap()
		id := rowMap["id"].AsString()
		if id == "1" {
			assert.Equal("John", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(35), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-01-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
			return
		}
	}

	updatedRows = selfVarMap["updated_rows"]
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows does not work with no primary key query trigger")

	// The "updated row" is now a deleted entry
	deletedKeys = selfVarMap["deleted_keys"]
	deletedKeysSlice := deletedKeys.AsValueSlice()
	assert.Equal(1, len(deletedKeysSlice), "wrong number of deleted keys")
	assert.Equal("a7f390faa9ddca021f647b042c2c127f70e49ed3bdec2194df4e784367b5416a", deletedKeysSlice[0].AsString(), "wrong deleted key")
}

func TestTriggerQueryB(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	outputPath := "./test_trigger_query_b.db"
	// Check if the directory exists
	_, err := os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	flowpipeDb := "./flowpipe.db"
	// Check if the directory exists
	_, err = os.Stat(flowpipeDb)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(flowpipeDb)
		if err != nil {
			assert.Fail("Error removing test directory", err)
			return
		}
	}

	db, err := initializeDB(outputPath)
	if err != nil {
		assert.Fail("Error initializing db", err)
		return
	}
	defer db.Close()

	err = createTestTableB(db, "test_one")
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

	err = populateTestTableB(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	pipeline := map[string]cty.Value{
		"name": cty.StringVal("test"),
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

	trigger := &modconfig.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "query.test_trigger_b",
		},
		Pipeline: cty.ObjectVal(pipeline),
		ArgsRaw:  hclExpressionMock,
	}

	trigger.Config = &modconfig.TriggerQuery{
		ConnectionString: "sqlite:./test_trigger_query_b.db",
		Sql:              "select * from test_one",
		PrimaryKey:       "id",
	}

	var triggerCommand interface{}
	commandBusMock := &util.CommandBusMock{
		SendFunc: func(ctx context.Context, command interface{}) error {
			triggerCommand = command
			return nil
		},
	}

	triggerRunner := NewTriggerRunner(ctx, commandBusMock, nil, trigger)

	triggerRunnerQuery := triggerRunner.(*TriggerRunnerQuery)
	triggerRunnerQuery.DatabasePath = "./flowpipe.db"

	assert.NotNil(triggerRunner, "trigger runner should not be nil")

	receiveChannel := make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res := <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")

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
			assert.Equal("2020-01-01T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else if idFloat == 2 {
			assert.Equal("Jane", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(25), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-02-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")
		} else if idFloat == 3 {
			assert.Equal("Joe", rowMap["name"].AsString(), "wrong name")
			assert.Equal(int64(40), util.BigFloatToInt64(rowMap["age"].AsBigFloat()), "wrong age")
			assert.Equal("2020-03-05T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(true, rowMap["is_active"].True(), "wrong is_active")
		} else {
			assert.Fail("wrong id")
		}
	}

	//
	// SECOND RUN
	//
	// No update

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
	assert.Equal(cty.ListValEmpty(cty.DynamicPseudoType), updatedRows, "updated rows should be nil, there's no new update detected by the query trigger")

	deletedKeys := selfVarMap["deleted_keys"]
	assert.Equal(cty.ListValEmpty(cty.String), deletedKeys, "deleted keys should be nil, there's no new deletion detected by the query trigger")

	//
	// THIRD RUN
	//
	// Update the blob data

	data[1]["blob_data"] = make([]byte, 10*(blobSizeMultiplier+2))
	for i := range data[1]["blob_data"].([]byte) {
		data[1]["blob_data"].([]byte)[i] = byte(rand.Intn(256)) //nolint:gosec // just a test case
	}

	err = updateTestTableB(db, "test_one", data[1])
	if err != nil {
		assert.Fail("Error updating test table", err)
		return
	}

	receiveChannel = make(chan error)
	triggerRunner.GetFqueue().RegisterCallback(receiveChannel)

	triggerRunner.Run()
	res = <-receiveChannel
	assert.Nil(res)

	assert.NotNil(triggerCommand, "trigger command should not be nil")
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
			assert.Equal("2020-02-20T00:00:00Z", rowMap["registration_date"].AsString(), "wrong registration date, registration date is converted to RFC3339 format during cty conversion")
			assert.Equal(false, rowMap["is_active"].True(), "wrong is_active")

			// Comparing blobData with otherSlice
			// TODO: byte array is converted to a number array in CTY. The conversion back to Go results in a float array. Perhaps we can improve this in the future, but for now I think we'll leave it at that.
		} else {
			assert.Fail("wrong id")
		}
	}
}
