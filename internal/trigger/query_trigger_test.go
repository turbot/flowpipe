package trigger

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	db, err := InitializeDB(outputPath)
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
			"registration_date": "2020-01-01",
			"is_active":         true,
		},
		{
			"id":                "3",
			"name":              "Joe",
			"age":               40,
			"registration_date": "2020-01-01",
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

	trigger := &modconfig.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "foo.bar.baz",
		},
		Pipeline: cty.ObjectVal(pipeline),
	}

	trigger.Config = &modconfig.TriggerQuery{
		ConnectionString: "sqlite:./test_trigger_query.db",
		Sql:              "select * from test_one",
		PrimaryKey:       "id",
	}

	triggerRunner := NewTriggerRunner(ctx, nil, nil, trigger)

	assert.NotNil(triggerRunner, "trigger runner should not be nil")

	triggerRunner.Run()
}
