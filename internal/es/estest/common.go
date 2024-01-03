package estest

import (
	"fmt"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func runPipeline(suite *FlowpipeTestSuite, name string, initialWaitTime time.Duration, args modconfig.Input) (*execution.ExecutionInMemory, *event.PipelineQueue, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		name = "local.pipeline." + name
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                name,
	}

	if args != nil {
		pipelineCmd.Args = args
	}

	if err := suite.esService.Send(pipelineCmd); err != nil {
		return nil, nil, fmt.Errorf("error sending pipeline command: %w", err)

	}

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	time.Sleep(initialWaitTime)

	ex, err := execution.GetExecution(pipelineCmd.Event.ExecutionID)

	if err != nil && perr.IsNotFound(err) {
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			ex, err = execution.GetExecution(pipelineCmd.Event.ExecutionID)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return nil, nil, err
	}

	return ex, pipelineCmd, nil
}

func getPipelineExAndWait(suite *FlowpipeTestSuite, evt *event.Event, pipelineExecutionID string, waitTime time.Duration, waitRetry int, expectedState string) (*execution.ExecutionInMemory, *execution.PipelineExecution, error) {

	plannerMutex := event.GetEventStoreMutex(evt.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	ex, err := execution.GetExecution(evt.ExecutionID)
	if err != nil {
		return nil, nil, err
	}

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		return nil, nil, fmt.Errorf("Pipeline execution " + pipelineExecutionID + " not found")
	}

	if pex.Status == expectedState {
		return ex, pex, nil
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < waitRetry; i++ {
		plannerMutex.Unlock()
		plannerMutex = nil

		time.Sleep(waitTime)

		plannerMutex = event.GetEventStoreMutex(evt.ExecutionID)
		plannerMutex.Lock()

		ex, err = execution.GetExecution(evt.ExecutionID)
		if err != nil {
			return nil, nil, err
		}

		pex = ex.PipelineExecutions[pipelineExecutionID]
		if pex == nil {
			return nil, nil, fmt.Errorf("Pipeline execution " + pipelineExecutionID + " not found")
		}

		if pex.Status == expectedState || pex.Status == "failed" || pex.Status == "finished" {
			break
		}
	}

	// if pex.Status == expectedState {
	// 	return ex, pex, nil
	// }

	if !pex.IsComplete() {
		return ex, pex, fmt.Errorf("not completed")
	}

	return ex, pex, nil
}
