package store

import (
	"log/slog"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
	putils "github.com/turbot/pipe-fittings/utils"
)

func StartPipeline(executionId, pipelineName string) error {
	retentionInSecond := viper.GetInt(constants.ArgProcessRetention)
	if retentionInSecond == 0 {
		return nil
	}

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
	currentTimeString := currentTime.Format(putils.RFC3339WithMS)
	_, err = stmt.Exec(executionId, pipelineName, "queued", currentTimeString, currentTimeString)
	if err != nil {
		sqlIteErr, ok := err.(sqlite3.Error)
		if ok && sqlIteErr.Code == sqlite3.ErrConstraint && sqlIteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			slog.Error("pipeline execution already exists", "executionID", executionId)
			return perr.BadRequestWithMessage("pipeline execution '" + executionId + "' already exists")
		}

		slog.Error("error executing statement", "error", err)
		return perr.InternalWithMessage("error executing statement " + err.Error())
	}

	return nil
}

func UpdatePipelineState(executionID, newState string) error {
	retentionInSecond := viper.GetInt(constants.ArgProcessRetention)
	if retentionInSecond == 0 {
		return nil
	}

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
	currentTimeString := currentTime.Format(putils.RFC3339WithMS)
	_, err = stmt.Exec(newState, currentTimeString, executionID)
	if err != nil {
		slog.Error("error executing update statement", "error", err)
		return perr.InternalWithMessage("error executing update statement " + err.Error())
	}

	return nil
}
