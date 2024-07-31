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

func UpgradeFlowpipeDB2() error {
	dbPath := filepaths.FlowpipeDBFileName()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	sql := `select value from internal where name = 'db_version'`

	rows, err := db.Query(sql)
	if err != nil {
		slog.Error("error getting current db_version", "error", err)
		return perr.InternalWithMessage("error getting current db_version")
	}

	// close DB here .. it will be reopened by the caller
	defer db.Close()

	var currentDbVersion string
	for rows.Next() {
		err = rows.Scan(&currentDbVersion)
		if err != nil {
			slog.Error("error getting db_version", "error", err)
			return perr.InternalWithMessage("error getting db_version")
		}
	}
	defer rows.Close()

	if currentDbVersion == "2.0" {
		return nil
	}

	createTableSQL := `
	create table if not exists process_log (
		id text primary key,
		struct_version text,
		process_id text,
		message text,
		level stext,
		created_at datetime,
		detail text,
		constraint fk_process_log_execution_id foreign key (process_id) references pipeline_run(execution_id) on delete cascade
	)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating process_log table", "error", err)
		return perr.InternalWithMessage("error creating process_log table")
	}

	createTableSQL = `drop table if exists event`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error dropping event table", "error", err)
		return perr.InternalWithMessage("error dropping event table")
	}

	processIdIndexSql := `create index if not exists idx_process_log_execution_id on process_log (process_id);`
	_, err = db.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating process_id index", "error", err)
		return perr.InternalWithMessage("error creating process_id index in process_log table")
	}

	processIdIndexSql = `create index if not exists idx_process_log_created_at on process_log (created_at);`
	_, err = db.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating created_at index", "error", err)
		return perr.InternalWithMessage("error creating created_at index in process_log_table")
	}

	if currentDbVersion == "" {
		updateMetadata := `insert into internal (name, value, created_at, updated_at) values ('db_version', '2.0', datetime('now'), datetime('now'))`
		_, err = db.Exec(updateMetadata)
		if err != nil {
			slog.Error("error updating metadata", "error", err)
			return perr.InternalWithMessage("error updating metadata")
		}
	} else {
		updateMetadata := `update internal set value = '2.0', updated_at = datetime('now') where name = 'db_version'`
		_, err = db.Exec(updateMetadata)
		if err != nil {
			slog.Error("error updating metadata", "error", err)
			return perr.InternalWithMessage("error updating metadata")
		}
	}

	return nil
}

func InitializeFlowpipeDB() error {

	err := moveFlowpipeDbFromModDirToFlowpipeModDir()
	if err != nil {
		return err
	}

	dbPath := filepaths.FlowpipeDBFileName()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		slog.Error("error enabling foreign key constraints (init)", "error", err, "dbPath", dbPath)
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

	processIdIndexSql := `create unique index if not exists idx_pipeline_run_execution_id on pipeline_run(execution_id)`
	_, err = db.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating pipeline_run index", "error", err)
		return perr.InternalWithMessage("error creating pipeline_run index")
	}

	createTableSQL = `
	create table if not exists process_log (
		id text primary key,
		struct_version text,
		process_id text,
		message text,
		level stext,
		created_at datetime,
		detail text,
		constraint fk_process_log_execution_id foreign key (process_id) references pipeline_run(execution_id) on delete cascade
	)`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating process_log table", "error", err)
		return perr.InternalWithMessage("error creating process_log table")
	}

	createTableSQL = `drop table if exists event`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		slog.Error("error dropping event table", "error", err)
		return perr.InternalWithMessage("error dropping event table")
	}

	processIdIndexSql = `create index if not exists idx_process_log_execution_id on event (process_id);`
	_, err = db.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating event index", "error", err)
		return perr.InternalWithMessage("error creating process_id index")
	}

	processIdIndexSql = `create index if not exists idx_process_log_created_at on event (created_at);`
	_, err = db.Exec(processIdIndexSql)
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

	processIdIndexSql = `create index if not exists idx_query_trigger_captured_row_trigger_name_primary_key on query_trigger_captured_row (trigger_name, primary_key);`
	_, err = db.Exec(processIdIndexSql)
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

	processIdIndexSql = `create unique index if not exists idx_internal_name on internal (name);`
	_, err = db.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating internal index", "error", err)
		return perr.InternalWithMessage("error creating internal index")
	}

	return nil
}

func OpenFlowpipeDB() (*sql.DB, error) {

	dbPath := filepaths.FlowpipeDBFileName()

	_, err := os.Stat(dbPath)

	if os.IsNotExist(err) {
		slog.Debug("flowpipe.db does not exist, creating it")
		err := InitializeFlowpipeDB()
		if err != nil {
			slog.Error("error initializing flowpipe.db", "error", err)
			return nil, err
		}
	} else {
		err := UpgradeFlowpipeDB2()
		if err != nil {
			slog.Error("error upgrading flowpipe.db", "error", err)
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, perr.InternalWithMessage("Error opening SQLite database " + err.Error())
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		slog.Error("error enabling foreign key constraints", "error", err, "dbPath", dbPath)
		return nil, perr.InternalWithMessage("error enabling foreign key constraints " + err.Error())
	}

	// Note: do not close the db connection here. The caller is responsible for closing it.
	return db, nil
}
