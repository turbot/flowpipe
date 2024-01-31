package store

import (
	"database/sql"
	"log/slog"

	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/pipe-fittings/perr"

	_ "github.com/mattn/go-sqlite3"
)

func InitializeFlowpipeDB() error {
	db, err := OpenFlowpipeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	createTableSQL := `
	create table if not exists event (
		id integer primary key autoincrement,
		execution_id string,
		created_at datetime,
		type text,
		data text
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating event table", "error", err)
		return perr.InternalWithMessage("error creating event table")
	}

	crateIndexSQL := `create index if not exists idx_execution_id on event (execution_id);`
	_, err = db.Exec(crateIndexSQL)
	if err != nil {
		slog.Error("error creating event index", "error", err)
		return perr.InternalWithMessage("error creating event index")
	}

	createTableSQL = `create table if not exists query_trigger_captured_row (trigger_name text, primary_key text, row_hash text, created_at text, updated_at text, primary key (trigger_name, primary_key));`

	slog.Debug("Creating table", "sql", createTableSQL)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating query_trigger_captured_row table", "error", err)
		return perr.InternalWithMessage("error creating query_trigger_captured_row table")
	}

	crateIndexSQL = `create index if not exists idx_data on query_trigger_captured_row (trigger_name, primary_key);`
	_, err = db.Exec(crateIndexSQL)
	if err != nil {
		slog.Error("error creating query_trigger_captured_row index", "error", err)
		return perr.ExecutionErrorWithMessage("error creating query_trigger_captured_row index")
	}

	return nil
}

func OpenFlowpipeDB() (*sql.DB, error) {
	dbPath := filepaths.FlowpipeDBFileName()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, perr.InternalWithMessage("Error opening SQLite database " + err.Error())
	}

	// Note: do not close the db connection here. The caller is responsible for closing it.
	return db, nil
}
