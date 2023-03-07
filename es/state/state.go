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

/*
type StackEntry struct {
	PipelineName string `json:"pipeline_name"`
	StepIndex    int    `json:"step_index"`
}

type Stack map[string]StackEntry
*/

type Stack struct {
	ID         string            `json:"id"`
	Status     string            `json:"status"`
	StepStatus map[int]string    `json:"pipeline_step_status"`
	Stacks     map[string]*Stack `json:"children"`
}

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
	PipelineName       string                 `json:"pipeline_name"`
	PipelineInput      map[string]interface{} `json:"pipeline_input"`
	PipelineStepStatus map[int]string         `json:"pipeline_step_status"`
	// Current execution stack
	ExecutionID string            `json:"run_id"`
	Stacks      map[string]*Stack `json:"stack"`
}

func NewState(ctx context.Context, e *event.Event) (*State, error) {
	s := &State{}
	s.PipelineStepStatus = map[int]string{}
	s.Stacks = map[string]*Stack{}
	err := s.LoadProcess(ctx, e)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *State) LookupStack(e *event.Event) (*Stack, error) {
	if len(e.StackIDs) == 0 {
		return nil, fmt.Errorf("event has no stack: %s", e.ExecutionID)
	}
	// Lookup the stack
	stacks := s.Stacks
	for _, stackID := range e.StackIDs[:len(e.StackIDs)-1] {
		stacks = stacks[stackID].Stacks
	}
	return stacks[e.StackIDs[len(e.StackIDs)-1]], nil
}

func (s *State) LoadProcess(ctx context.Context, e *event.Event) error {

	s.ExecutionID = e.ExecutionID

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

		case "event.PipelineQueued":
			// Get the run ID from the payload
			var et event.PipelineQueued
			err := json.Unmarshal(e.Payload, &et)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			s.PipelineName = et.Name
			s.PipelineInput = et.Input

			//lastStackID := et.Event.LastStackID()
			lastStackID := et.Event.StackIDs[len(et.Event.StackIDs)-1]
			s.Stacks[lastStackID] = &Stack{
				ID:         lastStackID,
				Status:     "queued",
				StepStatus: map[int]string{},
			}

			/*
				stack, err := s.LookupStack(et.Event)
				if err != nil {
					return err
				}
				stack.Status = "queued"
			*/

		case "event.PipelinePlanned":
			// Get the run ID from the payload
			var et event.PipelinePlanned
			err := json.Unmarshal(e.Payload, &et)
			if err != nil {
				// TODO - log and continue?
				return err
			}

			/*
				stack, err := s.LookupStack(et.Event)
				if err != nil {
					return err
				}
				stack.Status = "planned"
			*/

			//lastStackID := et.Event.LastStackID()
			lastStackID := et.Event.StackIDs[len(et.Event.StackIDs)-1]
			s.Stacks[lastStackID].Status = "planned"

			for _, i := range et.NextStepIndexes {
				s.Stacks[lastStackID].StepStatus[i] = "planned"
			}

		case "event.PipelineStepStart":
			var et event.PipelineStepStart
			err := json.Unmarshal(e.Payload, &et)
			if err != nil {
				// TODO - log and continue?
				return err
			}

			/*
				s.PipelineStepStatus[et.StepIndex] = "started"
			*/

			lastStackID := et.Event.LastStackID()
			s.Stacks[lastStackID].StepStatus[et.StepIndex] = "started"

		case "event.PipelineStepFinished":
			var et event.PipelineStepFinished
			err := json.Unmarshal(e.Payload, &et)
			if err != nil {
				// TODO - log and continue?
				return err
			}

			/*
				s.PipelineStepStatus[et.StepIndex] = "finished"
			*/

			lastStackID := et.Event.LastStackID()
			s.Stacks[lastStackID].StepStatus[et.StepIndex] = "finished"

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

func Dump(ctx context.Context) error {
	s := &State{}
	return s.Load(ctx)
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
			pipelineStatus[e.Event.ExecutionID] = "queue"

		case "event.PipelineQueued":
			// Get the run ID from the payload
			var e event.PipelineQueued
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "queued"

		case "event.PipelineLoad":
			// Get the run ID from the payload
			var e event.PipelineLoaded
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "load"

		case "event.PipelineLoaded":
			// Get the run ID from the payload
			var e event.PipelineLoaded
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "loaded"

		case "event.PipelineStart":
			// Get the run ID from the payload
			var e event.PipelineStarted
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "start"

		case "event.PipelineStarted":
			// Get the run ID from the payload
			var e event.PipelineStarted
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "started"

		case "event.PipelinePlan":
			// Get the run ID from the payload
			var e event.PipelinePlanned
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "plan"

		case "event.PipelinePlanned":
			// Get the run ID from the payload
			var e event.PipelinePlanned
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "planned"

		case "event.PipelineStepStart":
			var e event.PipelineStepStart
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "step_start"

		case "event.PipelineStepFinished":
			var e event.PipelineStepFinished
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "step_finished"

		case "event.PipelineFinish":
			// Get the run ID from the payload
			var e event.PipelineFinished
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "finish"

		case "event.PipelineFinished":
			// Get the run ID from the payload
			var e event.PipelineFinished
			err := json.Unmarshal(entryMetadata.Payload, &e)
			if err != nil {
				// TODO - log and continue?
				return err
			}
			pipelineStatus[e.Event.ExecutionID] = "finished"

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
