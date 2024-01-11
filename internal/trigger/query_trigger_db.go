package trigger

import (
	"database/sql"
	"log/slog"
	"sync"

	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/zclconf/go-cty/cty"
)

var queryTriggerLock sync.Mutex

func rowsToCtyList(newRows []map[string]interface{}) ([]cty.Value, error) {
	var newRowsCty []cty.Value
	for _, r := range newRows {
		rowCty, err := rowToCty(r)
		if err != nil {
			return nil, err
		}
		newRowsCty = append(newRowsCty, rowCty)
	}
	return newRowsCty, nil
}

func rowToCty(row map[string]interface{}) (cty.Value, error) {
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
	queryTriggerLock.Lock()
	defer queryTriggerLock.Unlock()

	db, err := initializeDB(dbPath)
	if err != nil {
		return nil, err
	}

	createTableSQL := `create table if not exists query_trigger_captured_row (trigger_name text, primary_key text, row_hash text, created_at text, updated_at text, primary key (trigger_name, primary_key));`

	slog.Info("Creating table", "sql", createTableSQL)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	crateIndexSQL := `create index if not exists idx_data on query_trigger_captured_row (trigger_name, primary_key);`
	_, err = db.Exec(crateIndexSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}
