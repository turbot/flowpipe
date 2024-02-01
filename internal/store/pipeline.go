package store

import (
	"log/slog"
	"time"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
)

func StartPipeline(executionID, pipelineName string) error {
	db, err := OpenFlowpipeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare the insert statement
	stmt, err := db.Prepare("insert into pipeline_run(execution_id, pipeline, state, started_at, updated_at) values(?, ?, ?, ?, ?)")
	if err != nil {
		slog.Error("error preparing statement", "error", err)
		return perr.InternalWithMessage("error preparing statement " + err.Error())
	}
	defer stmt.Close()

	// Execute the statement
	currentTime := time.Now().UTC()
	currentTimeString := currentTime.Format(util.RFC3389WithMS)
	_, err = stmt.Exec(executionID, pipelineName, "queued", currentTimeString, currentTimeString)
	if err != nil {
		slog.Error("error executing statement", "error", err)
		return perr.InternalWithMessage("error executing statement " + err.Error())
	}

	return nil
}

func UpdatePipelineState(executionID, newState string) error {
	db, err := OpenFlowpipeDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Prepare the update statement
	stmt, err := db.Prepare("update pipeline_run set state = ?, updated_at = ? where execution_id = ?")
	if err != nil {
		slog.Error("error preparing update statement", "error", err)
		return perr.InternalWithMessage("error preparing update statement " + err.Error())
	}
	defer stmt.Close()

	// Execute the update statement
	currentTime := time.Now().UTC()
	currentTimeString := currentTime.Format(util.RFC3389WithMS)
	_, err = stmt.Exec(newState, currentTimeString, executionID)
	if err != nil {
		slog.Error("error executing update statement", "error", err)
		return perr.InternalWithMessage("error executing update statement " + err.Error())
	}

	return nil
}
