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
