package store

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/pipe-fittings/perr"

	_ "github.com/mattn/go-sqlite3"
)

func moveFlowpipeDbFromModDirToFlowpipeModDir() error {

	sourcePath := filepaths.LegacyFlowpipeDBFileName()
	destPath := filepaths.FlowpipeDBFileName()

	// Check if flowpipe.db exists in Mod Dir
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil
	}

	// Check if flowpipe.db already exists in Mod's .flowpipe dir
	if _, err := os.Stat(destPath); err == nil {
		// flowpipe.db already exists in ModFlowpipeDir, aborting
		return perr.InternalWithMessage("flowpipe.db already exists in the mod's .flowpipe directory, aborting.")
	} else if !os.IsNotExist(err) {
		// An error other than "not exists", propagate it
		return err
	}

	// Move flowpipe.db
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		return err
	}

	return nil
}

func InitializeFlowpipeDB() error {

	err := moveFlowpipeDbFromModDirToFlowpipeModDir()
	if err != nil {
		return err
	}

	db, err := OpenFlowpipeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	createTableSQL := `create table if not exists pipeline_run (
		id integer primary key autoincrement,
		execution_id text,
		pipeline text,
		state text,
		started_at datetime,
		updated_at datetime
	)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating pipeline_run table", "error", err)
		return perr.InternalWithMessage("error creating pipeline_run table")
	}

	createIndexSQL := `create unique index if not exists idx_pipeline_run_execution_id on pipeline_run(execution_id)`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		slog.Error("error creating pipeline_run index", "error", err)
		return perr.InternalWithMessage("error creating pipeline_run index")
	}

	createTableSQL = `
	create table if not exists event (
		id integer primary key autoincrement,
		execution_id text,
		created_at datetime,
		type text,
		data text,
		constraint fk_event_execution_id foreign key (execution_id) references pipeline_run(execution_id) on delete cascade
	)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating event table", "error", err)
		return perr.InternalWithMessage("error creating event table")
	}

	createIndexSQL = `create index if not exists idx_event_execution_id on event (execution_id);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		slog.Error("error creating event index", "error", err)
		return perr.InternalWithMessage("error creating event index")
	}

	createIndexSQL = `create index if not exists idx_event_created_at on event (created_at);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		slog.Error("error creating event index", "error", err)
		return perr.InternalWithMessage("error creating event index")
	}

	createTableSQL = `create table if not exists query_trigger_captured_row (
		trigger_name text,
		primary_key text,
		row_hash text,
		created_at text,
		updated_at text,
		primary key (trigger_name, primary_key));`

	slog.Debug("Creating table", "sql", createTableSQL)
	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating query_trigger_captured_row table", "error", err)
		return perr.InternalWithMessage("error creating query_trigger_captured_row table")
	}

	createIndexSQL = `create index if not exists idx_query_trigger_captured_row_trigger_name_primary_key on query_trigger_captured_row (trigger_name, primary_key);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		slog.Error("error creating query_trigger_captured_row index", "error", err)
		return perr.ExecutionErrorWithMessage("error creating query_trigger_captured_row index")
	}

	createTableSQL = `
	create table if not exists internal (
		id integer primary key autoincrement,
		name string,
		created_at datetime,
		updated_at datetime,
		value text
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating internal table", "error", err)
		return perr.InternalWithMessage("error creating internal table")
	}

	createIndexSQL = `create unique index if not exists idx_internal_name on internal (name);`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		slog.Error("error creating internal index", "error", err)
		return perr.InternalWithMessage("error creating internal index")
	}

	return nil
}

func OpenFlowpipeDB() (*sql.DB, error) {
	dbPath := filepaths.FlowpipeDBFileName()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, perr.InternalWithMessage("Error opening SQLite database " + err.Error())
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		slog.Error("error enabling foreign key constraints", "error", err)
		return nil, perr.InternalWithMessage("error enabling foreign key constraints")
	}

	// Note: do not close the db connection here. The caller is responsible for closing it.
	return db, nil
}
