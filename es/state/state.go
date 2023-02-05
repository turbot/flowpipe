package state

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
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

func NewState(runID string) (*State, error) {
	s := &State{}
	s.Stack = Stack{}
	err := s.Load(runID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *State) Load(runID string) error {

	logFile := fmt.Sprintf("logs/%s.jsonl", runID)

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
