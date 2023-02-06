package state

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/utils"
)

type EventLogEntry struct {
	EventType string          `json:"event_type"`
	Timestamp *time.Time      `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type StackEntry struct {
	PipelineName string `json:"pipeline_name"`
	StepIndex    int    `json:"step_index"`
}

type Stack map[string]StackEntry

// Queue a mod for running in a given workspace context.
type State struct {
	// Host of the workspace. If empty, then assume localhost.
	CloudHost string `json:"host"`
	// The workspace context to use. May be a local workspace (e.g. default) or
	// a cloud workspace (e.g. e-gineer/scratch).
	Workspace string `json:"workspace"`
	// File system location where the mod is located, including pipeline
	// defintions.
	ModLocation string `json:"mod_location"`
	// Pipeline information
	PipelineName           string                 `json:"pipeline_name"`
	PipelineInput          map[string]interface{} `json:"pipeline_input"`
	PipelineCompletedSteps []int                  `json:"pipeline_completed_steps"`
	// Current execution stack
	RunID string `json:"run_id"`
	Stack Stack  `json:"stack"`
}

func NewState(ctx context.Context, runID string) (*State, error) {
	s := &State{}
	s.Stack = Stack{}
	err := s.LoadProcess(ctx, runID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func Dump(ctx context.Context) error {
	s := &State{}
	return s.Load(ctx)
}

func (s *State) LoadProcess(ctx context.Context, runID string) error {

	logFile := fmt.Sprintf("logs/%s.jsonl", utils.Session(ctx))

	// Open the event log
	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var e EventLogEntry
		err := json.Unmarshal(ba, &e)
		if err != nil {
			return err
		}

		switch e.EventType {

		case "event.PipelineQueued":
			// Get the run ID from the payload
			var queue event.PipelineQueued
			err := json.Unmarshal(e.Payload, &queue)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			if queue.RunID != runID {
				continue
			}
			s.RunID = queue.RunID
			s.PipelineName = queue.Name
			s.PipelineInput = queue.Input

		case "event.Queue":
			// Get the run ID from the payload
			var queue event.Queue
			err := json.Unmarshal(e.Payload, &queue)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			s.CloudHost = queue.CloudHost
			s.Workspace = queue.Workspace
			if queue.SpanID != runID {
				continue
			}

		/*
			case "event.PipelineStepExecute":
				var execute event.PipelineStepExecute
				err := json.Unmarshal(e.Payload, &execute)
				if err != nil {
					// TODO - log and continue?
					return err
				}
				s.Stack[execute.StackID] = StackEntry{PipelineName: execute.PipelineName, StepIndex: execute.StepIndex}
		*/

		case "event.PipelineStepExecuted":
			var executed event.PipelineStepExecuted
			err := json.Unmarshal(e.Payload, &executed)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			if executed.RunID != runID {
				continue
			}
			s.PipelineCompletedSteps = append(s.PipelineCompletedSteps, executed.StepIndex)

		default:
			// Ignore unknown types while loading
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	/*
		if s.SpanID == "" {
			return errors.New(fmt.Sprintf("load_failed: %s", logFile))
		}
	*/

	return nil

}

// This is an attempt at Loading the state of a session, including the status of
// any running pipelines. Next steps:
// * How do we know the flow of states, e.g. if load goes to failed (instead of loaded)?
// * How do we inject the right commands / events back into the bus to resume?
func (s *State) Load(ctx context.Context) error {

	logFile := fmt.Sprintf("logs/%s.jsonl", utils.Session(ctx))

	// Open the event log
	f, err := os.Open(logFile)
	if err != nil {
		return err
	}
	defer f.Close()

	pipelineStatus := map[string]string{}

	scanner := bufio.NewScanner(f)
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	for scanner.Scan() {

		ba := scanner.Bytes()

		// Get the run ID from the payload
		var entryMetadata EventLogEntry
		err := json.Unmarshal(ba, &entryMetadata)
		if err != nil {
			return err
		}

		switch entryMetadata.EventType {

		case "event.PipelineQueue":
			// Get the run ID from the payload
			var e event.PipelineQueue
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "queue"

		case "event.PipelineQueued":
			// Get the run ID from the payload
			var e event.PipelineQueued
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "queued"

		case "event.PipelineLoad":
			// Get the run ID from the payload
			var e event.PipelineLoaded
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "load"

		case "event.PipelineLoaded":
			// Get the run ID from the payload
			var e event.PipelineLoaded
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "loaded"

		case "event.PipelineStart":
			// Get the run ID from the payload
			var e event.PipelineStarted
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "start"

		case "event.PipelineStarted":
			// Get the run ID from the payload
			var e event.PipelineStarted
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "started"

		case "event.PipelinePlan":
			// Get the run ID from the payload
			var e event.PipelinePlanned
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "plan"

		case "event.PipelinePlanned":
			// Get the run ID from the payload
			var e event.PipelinePlanned
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "planned"

		case "event.PipelineStepExecute":
			var e event.PipelineStepExecute
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "step_execute"

		case "event.PipelineStepExecuted":
			var e event.PipelineStepExecuted
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "step_executed"

		case "event.PipelineFinish":
			// Get the run ID from the payload
			var e event.PipelineFinished
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "finish"

		case "event.PipelineFinished":
			// Get the run ID from the payload
			var e event.PipelineFinished
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.RunID] = "finished"

		default:
			// Ignore unknown types while loading
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Println(pipelineStatus)

	return nil

}
