package trigger

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func createTestTable(db *sql.DB, tableName string) error {

	createTableSQL := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (id text primary key, name text, age integer, registration_date date, is_active boolean);`

	slog.Info("Creating table", "sql", createTableSQL)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	return nil
}

func populateTestTable(db *sql.DB, tableName string, data []map[string]interface{}) error {
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

func updateTestTable(db *sql.DB, tableName string, data map[string]interface{}) error {
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

// func deleteFromTestTable(db *sql.DB, tableName string, idsToDelete []any) error {
// 	// Start a transaction
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}

// 	// Prepare statement for inserting into the temporary table
// 	placeholders := strings.Join(strings.Split(strings.Repeat("?", len(idsToDelete)), ""), ",")

// 	tempStmt, err := tx.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id in (%s)", tableName, placeholders))
// 	if err != nil {
// 		return err
// 	}
// 	defer tempStmt.Close()

// 	_, err = tempStmt.Exec(idsToDelete...)
// 	if err != nil {
// 		return err
// 	}

// 	// Commit the transaction
// 	if err := tx.Commit(); err != nil {
// 		slog.Error("Error committing transaction", "error", err)
// 		return err
// 	}

// 	return nil
// }

func TestTriggerQuery(t *testing.T) {

	// TODO: Test with integer as primary key
	// TODO: Test with jsonb column (does SQLIte support jsonb?)
	// TODO: Test with blob ... how do we detect changes?

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

	err = createTestTable(db, "test_one")
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

	err = populateTestTable(db, "test_one", data)
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

	triggerRunner.Run()

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
	triggerRunner.Run()

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
	assert.Equal(cty.NilVal, insertedRows, "inserted rows should be nil, there's no new addition detected by the query trigger")

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

	err = populateTestTable(db, "test_one", data)
	if err != nil {
		assert.Fail("Error populating test table", err)
		return
	}

	triggerRunner.Run()

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

	updateTestTable(db, "test_one", map[string]interface{}{
		"id":                "1",
		"name":              "John",
		"age":               35,
		"registration_date": "2020-01-01",
		"is_active":         false,
	})

	triggerRunner.Run()

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
	assert.Equal(cty.NilVal, insertedRows, "inserted rows should be nil, there's no new addition detected by the query trigger")

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
	// Delete some data

	// idsToDelete := []any{"1", "4"}
	// err = deleteFromTestTable(db, "test_one", idsToDelete)
	// if err != nil {
	// 	assert.Fail("Error deleting from test table", err)
	// 	return
	// }
}
