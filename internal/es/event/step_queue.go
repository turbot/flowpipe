package event

import (
	"fmt"
	"github.com/turbot/pipe-fittings/schema"
	"strings"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type StepQueue struct {
	// Event metadata
	Event *Event `json:"event"`
	// Step execution details
	PipelineExecutionID string          `json:"pipeline_execution_id"`
	StepExecutionID     string          `json:"step_execution_id"`
	StepName            string          `json:"step_name"`
	StepInput           modconfig.Input `json:"input"`

	// for_each controls
	StepForEach *modconfig.StepForEach `json:"step_for_each,omitempty"`
	StepLoop    *modconfig.StepLoop    `json:"step_loop,omitempty"`
	StepRetry   *modconfig.StepRetry   `json:"step_retry,omitempty"`

	NextStepAction modconfig.NextStepAction `json:"action,omitempty"`
}

func (e *StepQueue) GetEvent() *Event {
	return e.Event
}

func (e *StepQueue) HandlerName() string {
	return CommandStepQueue
}

type StepQueueOption func(*StepQueue) error

// NewStepQueue creates a new StepQueue event.
func NewStepQueue(opts ...StepQueueOption) (*StepQueue, error) {
	// Defaults
	e := &StepQueue{
		StepExecutionID: util.NewStepExecutionID(),
	}
	// Set options
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return e, err
		}
	}
	return e, nil
}

func NewStepQueueFromPipelineStepFinishedForLoop(e *StepFinished, stepName string) *StepQueue {

	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionID(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	}

	cmd.StepName = stepName
	cmd.StepInput = *e.StepLoop.Input
	cmd.StepForEach = e.StepForEach
	cmd.StepLoop = e.StepLoop
	cmd.StepRetry = e.StepRetry
	cmd.NextStepAction = modconfig.NextStepActionStart

	return cmd
}

func NewStepQueueFromPipelineStepFinishedForRetry(e *StepFinished, stepName string) *StepQueue {

	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionID(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	}

	cmd.StepName = stepName
	cmd.StepInput = *e.StepRetry.Input
	cmd.StepForEach = e.StepForEach
	cmd.StepLoop = e.StepLoop
	cmd.StepRetry = e.StepRetry
	cmd.NextStepAction = modconfig.NextStepActionStart

	return cmd
}

func NewStepQueueFromStepForEachPlanned(e *StepForEachPlanned, nextStep *modconfig.NextStep) (*StepQueue, error) {
	cmd := &StepQueue{
		Event:           NewChildEvent(e.Event),
		StepExecutionID: util.NewStepExecutionID(),
	}
	if e.PipelineExecutionID != "" {
		cmd.PipelineExecutionID = e.PipelineExecutionID
	} else {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline planned event: %v", e))
	}

	extendedInput := extendInputs(cmd, e.StepName, nextStep.Input)
	cmd.StepName = e.StepName
	cmd.StepInput = extendedInput
	cmd.StepForEach = nextStep.StepForEach
	cmd.StepLoop = nil
	cmd.NextStepAction = nextStep.Action

	return cmd, nil
}

func StepQueueForPipelinePlanned(e *PipelinePlanned) StepQueueOption {
	return func(cmd *StepQueue) error {
		cmd.Event = NewChildEvent(e.Event)
		if e.PipelineExecutionID != "" {
			cmd.PipelineExecutionID = e.PipelineExecutionID
		} else {
			return perr.BadRequestWithMessage(fmt.Sprintf("missing pipeline execution ID in pipeline planned event: %v", e))
		}
		return nil
	}
}

func StepQueueWithStep(name string, input modconfig.Input, stepForEach *modconfig.StepForEach, stepLoop *modconfig.StepLoop, nextStepAction modconfig.NextStepAction) StepQueueOption {
	return func(cmd *StepQueue) error {
		extendedInput := extendInputs(cmd, name, input)
		cmd.StepName = name
		cmd.StepInput = extendedInput
		cmd.StepForEach = stepForEach
		cmd.StepLoop = stepLoop
		cmd.NextStepAction = nextStepAction
		return nil
	}
}

// TODO: refactor/tidy
func extendInputs(cmd *StepQueue, stepName string, input modconfig.Input) modconfig.Input {
	stepType := strings.Split(stepName, ".")[0]
	switch stepType {
	case "input":
		if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
			if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
				for _, n := range notifies {
					if notify, ok := n.(map[string]any); ok {
						integration := notify["integration"].(map[string]any)
						integrationType := integration["type"].(string)
						switch integrationType {
						case schema.IntegrationTypeEmail, schema.IntegrationTypeWebform:
							webformUrl, _ := util.GetWebformUrl(cmd.Event.ExecutionID, cmd.PipelineExecutionID, cmd.StepExecutionID)
							input["webform_url"] = webformUrl
							return input
						default:
							// slack, teams, etc - do nothing
						}
					}
				}
			}
		}
		return input
	default:
		return input
	}
}
