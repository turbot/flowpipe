package store

import (
	"log/slog"

	"github.com/turbot/pipe-fittings/perr"
)

func ListExecutionIDs() ([]string, error) {
	db, err := OpenFlowpipeDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("select distinct process_id from event order by id desc")
	if err != nil {
		slog.Error("error querying process", "error", err)
		return nil, perr.InternalWithMessage("error querying process")
	}
	defer rows.Close()

	var executionIDs []string
	for rows.Next() {
		var executionID string
		err = rows.Scan(&executionID)
		if err != nil {
			slog.Error("error scanning process", "error", err)
			return nil, perr.InternalWithMessage("error scanning process")
		}
		executionIDs = append(executionIDs, executionID)
	}

	return executionIDs, nil
}
