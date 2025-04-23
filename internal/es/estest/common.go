package estest

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
)

func runPipelineWithId(suite *FlowpipeTestSuite, executionId, name string, initialWaitTime time.Duration, args resources.Input) (*execution.ExecutionInMemory, *event.PipelineQueue, error) {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		name = "local.pipeline." + name
	}

	executionCmd := event.NewExecutionQueueForPipeline(executionId, name)

	if args != nil {
		executionCmd.PipelineQueue.Args = args
	}

	if err := suite.esService.Send(executionCmd); err != nil {
		return nil, nil, fmt.Errorf("error sending pipeline command: %w", err)

	}

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	time.Sleep(initialWaitTime)

	ex, err := execution.GetExecution(executionCmd.Event.ExecutionID)

	if err != nil && perr.IsNotFound(err) {
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			ex, err = execution.GetExecution(executionCmd.Event.ExecutionID)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return nil, nil, err
	}

	return ex, executionCmd.PipelineQueue, nil
}

func runPipeline(suite *FlowpipeTestSuite, name string, initialWaitTime time.Duration, args resources.Input) (*execution.ExecutionInMemory, *event.PipelineQueue, error) {
	return runPipelineWithId(suite, "", name, initialWaitTime, args)
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
		return nil, nil, errors.New("Pipeline execution " + pipelineExecutionID + " not found")
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
			return nil, nil, errors.New("Pipeline execution " + pipelineExecutionID + " not found")
		}

		if pex.Status == expectedState || pex.Status == "failed" || pex.Status == "finished" {
			// check the ex.Status as well
			if ex.Status == expectedState || ex.Status == "failed" || ex.Status == "finished" {
				break
			}
		}
	}

	if !pex.IsComplete() {
		return ex, pex, errors.New("not completed")
	}

	// This is a simple 1:1 mapping between execution & pipeline execution. The bulk of the test cases were written before
	// we elevated Execution to have a status and so we're just going to keep the same pattern for now.
	if ex.Status != expectedState {
		return ex, pex, fmt.Errorf("pipeline execution %s expected state '%s' execution status '%s'", pipelineExecutionID, expectedState, ex.Status)
	}

	return ex, pex, nil
}

func getExAndWait(suite *FlowpipeTestSuite, executionId string, waitTime time.Duration, waitRetry int, expectedState string) (*execution.ExecutionInMemory, error) {

	plannerMutex := event.GetEventStoreMutex(executionId)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	ex, err := execution.GetExecution(executionId)
	if err != nil {
		return nil, err
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < waitRetry; i++ {
		plannerMutex.Unlock()
		plannerMutex = nil

		time.Sleep(waitTime)

		plannerMutex = event.GetEventStoreMutex(executionId)
		plannerMutex.Lock()

		ex, err = execution.GetExecution(executionId)
		if err != nil {
			return nil, err
		}

		if ex.Status == expectedState {
			break
		}
	}

	return ex, nil
}

func getPipelineExWaitForStepStarted(suite *FlowpipeTestSuite, evt *event.Event, pipelineExecutionID string, waitTime time.Duration, waitRetry int, stepName string) (*execution.ExecutionInMemory, *execution.PipelineExecution, *execution.StepExecution, error) {
	startedStatuses := []string{"starting", "started", "finished", "failed"}
	plannerMutex := event.GetEventStoreMutex(evt.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	ex, err := execution.GetExecution(evt.ExecutionID)
	if err != nil {
		return nil, nil, nil, err
	}

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		return nil, nil, nil, errors.New("Pipeline execution " + pipelineExecutionID + " not found")
	}

	for _, stEx := range pex.StepExecutions {
		if stEx.Name == stepName && slices.Contains(startedStatuses, stEx.Status) {
			return ex, pex, stEx, nil
		}
	}

	for i := 0; i < waitRetry; i++ {
		plannerMutex.Unlock()
		plannerMutex = nil

		time.Sleep(waitTime)

		plannerMutex = event.GetEventStoreMutex(evt.ExecutionID)
		plannerMutex.Lock()

		ex, err = execution.GetExecution(evt.ExecutionID)
		if err != nil {
			return nil, nil, nil, err
		}

		pex = ex.PipelineExecutions[pipelineExecutionID]
		if pex == nil {
			return nil, nil, nil, errors.New("Pipeline execution " + pipelineExecutionID + " not found")
		}

		for _, stEx := range pex.StepExecutions {
			if stEx.Name == stepName && slices.Contains(startedStatuses, stEx.Status) {
				return ex, pex, stEx, nil
			}
		}

		if pex.Status == "failed" || pex.Status == "finished" {
			return nil, nil, nil, fmt.Errorf("pipeline execution %s completed but expected step %s didn't start", pipelineExecutionID, stepName)
		}
	}

	return nil, nil, nil, fmt.Errorf("pipeline execution %s wait retries completed but expected step %s didn't start", pipelineExecutionID, stepName)
}
