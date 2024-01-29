package trigger

import (
	"database/sql"
	"log/slog"
	"sync"
)

var queryTriggerLock sync.Mutex

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

	slog.Debug("Creating table", "sql", createTableSQL)
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
