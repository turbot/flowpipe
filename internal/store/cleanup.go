package store

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
	putils "github.com/turbot/pipe-fittings/utils"
)

func cleanupFlowpipeDB(currentTime time.Time, offset time.Duration) (int, error) {
	slog.Debug("Cleaning up flowpipe db")
	db, err := OpenFlowpipeDB()
	if err != nil {
		slog.Error("error opening flowpipe db", "error", err)
		return -1, perr.InternalWithMessage("error opening flowpipe db")
	}
	defer db.Close()

	timeLimit := currentTime.Add(offset)

	// TODO: how do we cleanup orphaned pipeline runs? We can filter this to just 'finished', 'cancelled' and 'failed' states
	// but then the orphan pipeline will never be cleaned. Should we have a hard limit?
	cleanupQuery := `delete from pipeline_run where updated_at < ?;`

	timeAsString := timeLimit.Format(putils.RFC3339WithMS)

	result, err := db.Exec(cleanupQuery, timeAsString)
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

	sql := `select value from internal where name = 'last_cleanup'`

	rows, err := db.Query(sql)
	if err != nil {
		slog.Error("error getting last cleanup time", "error", err)
		return -1, perr.InternalWithMessage("error getting last cleanup time")
	}

	var lastCleanupTime string
	for rows.Next() {
		err = rows.Scan(&lastCleanupTime)
		if err != nil {
			slog.Error("error getting last cleanup time", "error", err)
			return -1, perr.InternalWithMessage("error getting last cleanup time")
		}
	}
	defer rows.Close()

	currentTimeStringFormat := currentTime.Format(putils.RFC3339WithMS)

	if lastCleanupTime != "" {
		sql = `update internal set value = ?, updated_at = ? where name = 'last_cleanup'`
		_, err = db.Exec(sql, currentTimeStringFormat, currentTimeStringFormat)
		if err != nil {
			slog.Error("error updating last cleanup time", "error", err)
			return -1, perr.InternalWithMessage("error updating last cleanup time")
		}
	} else {
		sql = `insert into internal (name, value, created_at) values ('last_cleanup', ?, ?)`
		_, err = db.Exec(sql, currentTimeStringFormat, currentTimeStringFormat)
		if err != nil {
			slog.Error("error inserting last cleanup time", "error", err)
			return -1, perr.InternalWithMessage("error inserting last cleanup time")
		}
	}

	return int(rowsAffected), nil
}

func CleanupRunner() {

	currentTime := time.Now().UTC()

	retentionInSecond := viper.GetInt(constants.ArgProcessRetention)
	if retentionInSecond == -1 {
		slog.Debug("Skipping cleanup as retention is set to -1")
		return
	}

	offset := time.Duration(-1*retentionInSecond) * time.Second

	rowsAffected, err := cleanupFlowpipeDB(currentTime, offset)
	if err != nil {
		slog.Error("error cleaning up flowpipe db", "error", err)
		return
	}

	slog.Info("Cleaned up flowpipe db", "rowsAffected", rowsAffected)

	deleteOldJsonlFiles(filepaths.EventStoreDir(), offset)
}

// Force cleanup run if we haven't run it more than 1 day
func ForceCleanup() {
	// can only clean up if flowpipe.db exist
	dbPath := filepaths.FlowpipeDBFileName()

	_, err := os.Stat(dbPath)

	if os.IsNotExist(err) {
		slog.Debug("Skipping force cleanup as flowpipe.db does not exist")
		return
	}

	slog.Debug("Checking if cleanup must be run")

	sql := `select value from internal where name = 'last_cleanup'`
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
		lastCleanupTime, err := time.Parse(putils.RFC3339WithMS, lastCleanupTime)
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
}

// This function should be removed eventually. SQLite store is out in v0.3.
func deleteOldJsonlFiles(dir string, olderThan time.Duration) {
	// Read files in directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		slog.Error("error reading directory", "error", err, "dir", dir)
		return
	}

	// Current time
	now := time.Now()

	for _, entry := range entries {
		// Ignore directories and non-jsonl files
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		// Get FileInfo for the file
		info, err := entry.Info()
		if err != nil {
			slog.Error("error getting info for file", "error", err, "file", entry.Name())
			continue
		}

		// Calculate the file's age
		fileAge := now.Sub(info.ModTime())

		// Check if the file is older than the specified duration
		if fileAge > olderThan {
			// Construct file path
			filePath := dir + "/" + entry.Name()

			// Delete the file
			err := os.Remove(filePath)
			if err != nil {
				slog.Error("error deleting file", "error", err, "file", filePath)
			} else {
				slog.Debug("Deleted file", "file", filePath)
			}
		}
	}

}
