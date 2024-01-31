package store

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/turbot/pipe-fittings/perr"
)

func CleanupFlowpipeDB(currentTime time.Time, offset time.Duration) (int, error) {
	slog.Debug("Cleaning up flowpipe db")
	db, err := OpenFlowpipeDB()
	if err != nil {
		slog.Error("error opening flowpipe db", "error", err)
		return -1, perr.InternalWithMessage("error opening flowpipe db")
	}
	defer db.Close()

	timeLimit := currentTime.Add(offset)

	fmt.Print("timeLimit: ", timeLimit, "\n")

	cleanupQuery := `delete from event
	where execution_id in (
		select execution_id
		from event
		group by execution_id
		having min(created_at) < ? and max(created_at) < ?)`

	timeAsString := timeLimit.Format("2006-01-02T15:04:05Z")

	result, err := db.Exec(cleanupQuery, timeAsString, timeAsString)
	if err != nil {
		slog.Error("error cleaning up flowpipe db", "error", err)
		return -1, perr.InternalWithMessage("error cleaning up flowpipe db")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("error cleaning up flowpipe db", "error", err)
		return -1, perr.InternalWithMessage("error cleaning up flowpipe db")
	}

	slog.Debug("Cleaned up flowpipe db", "rowsAffected", rowsAffected)
	return int(rowsAffected), nil
}

func CleanupRunner() {

	currentTime := time.Now().UTC()

	// TODO: configure this
	offset := -1 * time.Hour

	rowsAffected, err := CleanupFlowpipeDB(currentTime, offset)
	if err != nil {
		slog.Error("error cleaning up flowpipe db", "error", err)
		return
	}

	slog.Info("Cleaned up flowpipe db", "rowsAffected", rowsAffected)
}

// Force cleanup run if we haven't run it more than 1 day
func ForceCleanup() {
	slog.Debug("Checking if cleanup must be run")

	sql := `select value from metadata where name = 'last_cleanup'`
	db, err := OpenFlowpipeDB()
	if err != nil {
		slog.Error("error opening flowpipe db", "error", err)
		return
	}
	defer db.Close()

	rows, err := db.Query(sql)
	if err != nil {
		slog.Error("error getting last cleanup time", "error", err)
		return
	}

	var lastCleanupTime string
	for rows.Next() {
		err = rows.Scan(&lastCleanupTime)
		if err != nil {
			slog.Error("error getting last cleanup time", "error", err)
			return
		}
	}
	defer rows.Close()

	runCleanup := false
	if lastCleanupTime == "" {
		runCleanup = true
	} else {
		lastCleanupTime, err := time.Parse("2006-01-02T15:04:05Z", lastCleanupTime)
		if err != nil {
			slog.Error("error parsing last cleanup time", "error", err)
			runCleanup = true
		}

		// force run cleanup if we haven't run cleanup for 1 day
		if time.Now().UTC().Sub(lastCleanupTime) > 24*time.Hour {
			runCleanup = true
		}
	}

	if !runCleanup {
		slog.Debug("Skipping force cleanup")
		return
	}

	slog.Debug("Running force cleanup")

	CleanupRunner()
	currentTime := time.Now().UTC()
	currentTimeStringFormat := currentTime.Format("2006-01-02T15:04:05Z")

	if lastCleanupTime != "" {
		sql = `update metadata set value = ?, updated_at = ? where name = 'last_cleanup'`
		_, err = db.Exec(sql, currentTimeStringFormat, currentTimeStringFormat)
		if err != nil {
			slog.Error("error updating last cleanup time", "error", err)
			return
		}
	} else {
		sql = `insert into metadata (name, value, created_at) values ('last_cleanup', ?, ?)`
		_, err = db.Exec(sql, currentTimeStringFormat, currentTimeStringFormat)
		if err != nil {
			slog.Error("error inserting last cleanup time", "error", err)
			return
		}
	}
}
