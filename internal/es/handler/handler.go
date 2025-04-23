package handler

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/metrics"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/pipe-fittings/cache"
	"github.com/turbot/pipe-fittings/perr"
)

type EventHandler struct {
	// Event handlers can only send commands, they are not even permitted access
	// to the EventBus.
	CommandBus FpCommandBus
}

type FpCommandBus interface {
	Send(ctx context.Context, command interface{}) error
}

type FpCommandBusImpl struct {
	Cb *cqrs.CommandBus
}

// Send sends command to the command bus.
func (c FpCommandBusImpl) Send(ctx context.Context, cmd interface{}) error {

	err := LogEventMessage(ctx, cmd, nil)
	if err != nil {
		return err
	}
	return c.Cb.Send(ctx, cmd)
}

func LogEventMessage(ctx context.Context, cmd interface{}, lock *sync.Mutex) error {

	commandEvent, ok := cmd.(event.CommandEvent)

	if !ok {
		return perr.BadRequestWithMessage("event is not a CommandEvent")
	}

	logMessage := event.NewEventLogFromCommand(commandEvent)

	if strings.ToLower(os.Getenv("FLOWPIPE_EVENT_FORMAT")) == "jsonl" {
		err := execution.LogEventMessageToFile(ctx, logMessage)
		if err != nil {
			return err
		}
	}

	db, err := store.OpenFlowpipeDB()
	if err != nil {
		return perr.InternalWithMessage("Error opening SQLite database " + err.Error())
	}
	defer db.Close()

	newExecution := false

	var name string
	if executionQueueCmd, ok := commandEvent.(*event.ExecutionQueue); ok {
		newExecution = true
		if executionQueueCmd.TriggerQueue != nil {
			name = executionQueueCmd.TriggerQueue.Name
		} else if executionQueueCmd.PipelineQueue != nil {
			name = executionQueueCmd.PipelineQueue.Name
		} else {
			return perr.BadRequestWithMessage("Invalid ExecutionQueue command, no TriggerQueue or PipelineQueue")
		}
	}

	var ex *execution.ExecutionInMemory
	executionID := commandEvent.GetEvent().ExecutionID
	if newExecution {
		ex = &execution.ExecutionInMemory{
			Execution: execution.Execution{
				ID:                 executionID,
				PipelineExecutions: map[string]*execution.PipelineExecution{},
				Lock:               event.GetEventStoreMutex(executionID),
			},
		}

		// Effectively forever
		ok := cache.GetCache().SetWithTTL(executionID, ex, 10*365*24*time.Hour)
		if !ok {
			slog.Error("Error setting execution in cache", "execution_id", executionID)
			return perr.InternalWithMessage("Error setting execution in cache")
		}

		metrics.RunMetricInstance.StartExecution(executionID, name)

		err = store.StartPipeline(executionID, name)
		if err != nil {
			slog.Error("Unable to save pipeline in the database", "error", err)
			return err
		}

	} else {
		var err error
		ex, err = execution.GetExecution(executionID)
		if err != nil {
			slog.Error("Error getting execution from cache to log event.", "execution_id", executionID, "error", err)
			return perr.InternalWithMessage("Error getting execution from cache to log event")
		}
	}

	err = ex.AddEvent(logMessage)
	if err != nil {
		slog.Error("Error adding event to execution", "error", err)
		return err
	}

	err = execution.SaveEventToSQLite(db, executionID, logMessage)
	if err != nil {
		slog.Error("Error saving event to SQLite", "error", err)
		return err
	}

	return nil
}

func pipelineCompletionHandler(executionID, pipelineExecutionID string, pipelineDefn *resources.Pipeline, stepExecutions map[string]*execution.StepExecution) {
	event.ReleaseEventLogMutex(executionID)
	execution.CompletePipelineExecutionStepSemaphore(pipelineExecutionID)
	err := execution.ReleasePipelineSemaphore(pipelineDefn)
	if err != nil {
		slog.Error("Releasing pipeline semaphore", "error", err)
	}

	for _, se := range stepExecutions {
		db.RemoveStepExecutionIDMap(se.ID)
	}
}
