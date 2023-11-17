package inprocess

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"time"
)

func RunPipeline(ctx context.Context, esService *es.ESService, pipelineName string, initialWaitTime time.Duration, args modconfig.Input) (*execution.Execution, *event.PipelineQueue, error) {
	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
		Args:                args,
	}

	if err := esService.Send(pipelineCmd); err != nil {
		return nil, nil, fmt.Errorf("error sending pipeline command: %w", err)

	}

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	time.Sleep(initialWaitTime)

	// check if the execution id has been completed, check 3 times
	ex, err := execution.NewExecution(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = ex.LoadProcess(pipelineCmd.Event)
	if err != nil {
		return nil, nil, err
	}

	return ex, pipelineCmd, nil
}

func GetPipelineExAndWait(ctx context.Context, event *event.Event, pipelineExecutionID string, waitTime time.Duration, waitRetry int, expectedState string) (*execution.Execution, *execution.PipelineExecution, error) {
	// check if the execution id has been completed, check 3 times
	ex, err := execution.NewExecution(ctx)
	if err != nil {
		return nil, nil, err
	}

	err = ex.LoadProcess(event)
	if err != nil {
		return nil, nil, err
	}

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		return nil, nil, fmt.Errorf("Pipeline execution " + pipelineExecutionID + " not found")
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < waitRetry; i++ {
		time.Sleep(waitTime)

		err = ex.LoadProcess(event)
		if err != nil {
			return nil, nil, fmt.Errorf("error loading process: %w", err)
		}
		if pex == nil {
			return nil, nil, fmt.Errorf("Pipeline execution " + pipelineExecutionID + " not found")
		}
		pex = ex.PipelineExecutions[pipelineExecutionID]

		if pex.Status == expectedState || pex.Status == "failed" || pex.Status == "finished" {
			break
		}
	}

	if !pex.IsComplete() {
		return ex, pex, fmt.Errorf("not completed")
	}

	return ex, pex, nil
}
