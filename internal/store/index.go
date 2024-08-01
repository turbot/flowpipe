package store

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"

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

func createEventTable(tx *sql.Tx) error {
	createTableSQL := `
	create table if not exists event (
		id text primary key,
		struct_version text,
		process_id text,
		message text,
		level stext,
		created_at datetime,
		detail text,
		constraint fk_event_execution_id foreign key (process_id) references pipeline_run(execution_id) on delete cascade
	)`

	_, err := tx.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating event table", "error", err)
		return perr.InternalWithMessage("error creating event table")
	}

	processIdIndexSql := `create index if not exists idx_event_process_id on event (process_id);`
	_, err = tx.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating event index", "error", err)
		return perr.InternalWithMessage("error creating process_id index")
	}

	processIdIndexSql = `create index if not exists idx_event_created_at on event (created_at);`
	_, err = tx.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating event index", "error", err)
		return perr.InternalWithMessage("error creating event index")
	}

	return nil
}

func migrateEventTable(tx *sql.Tx) error {
	// Select data from the event_old table
	rows, err := tx.Query(`SELECT id, execution_id, created_at, type, data FROM event_old`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Prepare insert statement for the event table
	stmt, err := tx.Prepare(`
			INSERT INTO event (id, struct_version, process_id, message, level, created_at, detail)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id int
		var executionID, createdAt, eventType, data string
		if err := rows.Scan(&id, &executionID, &createdAt, &eventType, &data); err != nil {
			log.Fatal(err)
		}

		// Generate new ID using the utility function
		newID := util.NewProcessLogId()

		// Set level to "event"
		level := "event"

		// Use current timestamp if created_at is not a valid datetime
		parsedCreatedAt, err := time.Parse(utils.RFC3339WithMS, createdAt)
		if err != nil {
			parsedCreatedAt = time.Now()
		}

		// Insert data into the event table
		_, err = stmt.Exec(newID, "2.0", executionID, eventType, level, parsedCreatedAt, data)
		if err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	slog.Debug("Migrated event_old table to event table")
	return nil
}

func UpgradeFlowpipeDB2() error {
	dbPath := filepaths.FlowpipeDBFileName()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	dbVersionSql := `select value from internal where name = 'db_version'`

	rows, err := db.Query(dbVersionSql)
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
		}
	}
	defer rows.Close()

	if currentDbVersion == "2.0" {
		return nil
	}
	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		slog.Error("error starting transaction", "error", err)
		return perr.InternalWithMessage("error starting transaction")
	}

	commited := false
	defer func() {
		if !commited {
			err := tx.Rollback()
			if err != nil {
				slog.Error("error rolling back transaction", "error", err)
			}
		}
	}()

	// rename event table first
	renameTableEvent := `alter table event rename to event_old`
	_, err = tx.Exec(renameTableEvent)
	if err != nil {
		slog.Error("error renaming event table", "error", err)
		return perr.InternalWithMessage("error renaming event table")
	}

	dropIndexSql := `drop index if exists idx_event_execution_id`
	_, err = tx.Exec(dropIndexSql)
	if err != nil {
		slog.Error("error dropping index", "error", err)
		return perr.InternalWithMessage("error dropping index")
	}

	dropIndexSql = `drop index if exists idx_event_created_at`
	_, err = tx.Exec(dropIndexSql)
	if err != nil {
		slog.Error("error dropping index", "error", err)
		return perr.InternalWithMessage("error dropping index")
	}

	err = createEventTable(tx)
	if err != nil {
		slog.Error("error creating event table", "error", err)
		return perr.InternalWithMessage("error creating event table")
	}

	err = migrateEventTable(tx)
	if err != nil {
		slog.Error("error migrating event table", "error", err)
		return perr.InternalWithMessage("error migrating event table")
	}

	dropTableSql := `drop table if exists event_old`
	_, err = tx.Exec(dropTableSql)
	if err != nil {
		slog.Error("error dropping table", "error", err)
		return perr.InternalWithMessage("error dropping table")
	}

	if currentDbVersion == "" {
		updateMetadata := `insert into internal (name, value, created_at, updated_at) values ('db_version', '2.0', datetime('now'), datetime('now'))`
		_, err = tx.Exec(updateMetadata)
		if err != nil {
			slog.Error("error updating metadata", "error", err)
			return perr.InternalWithMessage("error updating metadata")
		}
	} else {
		updateMetadata := `update internal set value = '2.0', updated_at = datetime('now') where name = 'db_version'`
		_, err = tx.Exec(updateMetadata)
		if err != nil {
			slog.Error("error updating metadata", "error", err)
			return perr.InternalWithMessage("error updating metadata")
		}
	}

	err = tx.Commit()
	if err != nil {
		slog.Error("error committing transaction", "error", err)
		return perr.InternalWithMessage("error committing transaction")
	}
	commited = true

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

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		slog.Error("error starting transaction", "error", err)
		return perr.InternalWithMessage("error starting transaction")
	}

	commited := false
	defer func() {
		if !commited {
			err := tx.Rollback()
			if err != nil {
				slog.Error("error rolling back transaction", "error", err)
			}
		}
	}()

	createTableSQL := `create table if not exists pipeline_run (
		id integer primary key autoincrement,
		execution_id text,
		pipeline text,
		state text,
		started_at datetime,
		updated_at datetime
	)`

	_, err = tx.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating pipeline_run table", "error", err)
		return perr.InternalWithMessage("error creating pipeline_run table")
	}

	processIdIndexSql := `create unique index if not exists idx_pipeline_run_execution_id on pipeline_run(execution_id)`
	_, err = tx.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating pipeline_run index", "error", err)
		return perr.InternalWithMessage("error creating pipeline_run index")
	}

	err = createEventTable(tx)
	if err != nil {
		slog.Error("error creating event table", "error", err)
		return perr.InternalWithMessage("error creating event table")
	}

	createTableSQL = `create table if not exists query_trigger_captured_row (
		trigger_name text,
		primary_key text,
		row_hash text,
		created_at text,
		updated_at text,
		primary key (trigger_name, primary_key));`

	slog.Debug("Creating table", "sql", createTableSQL)
	_, err = tx.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating query_trigger_captured_row table", "error", err)
		return perr.InternalWithMessage("error creating query_trigger_captured_row table")
	}

	processIdIndexSql = `create index if not exists idx_query_trigger_captured_row_trigger_name_primary_key on query_trigger_captured_row (trigger_name, primary_key);`
	_, err = tx.Exec(processIdIndexSql)
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

	_, err = tx.Exec(createTableSQL)
	if err != nil {
		slog.Error("error creating internal table", "error", err)
		return perr.InternalWithMessage("error creating internal table")
	}

	processIdIndexSql = `create unique index if not exists idx_internal_name on internal (name);`
	_, err = tx.Exec(processIdIndexSql)
	if err != nil {
		slog.Error("error creating internal index", "error", err)
		return perr.InternalWithMessage("error creating internal index")
	}

	updateMetadata := `insert into internal (name, value, created_at, updated_at) values ('db_version', '2.0', datetime('now'), datetime('now'))`
	_, err = tx.Exec(updateMetadata)
	if err != nil {
		slog.Error("error updating metadata", "error", err)
		return perr.InternalWithMessage("error updating metadata")
	}

	err = tx.Commit()
	if err != nil {
		slog.Error("error committing transaction", "error", err)
		return perr.InternalWithMessage("error committing transaction")
	}
	commited = true

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
