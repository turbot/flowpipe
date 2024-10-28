package event

import (
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
)

type StepForEachPlanned struct {
	Event               *Event              `json:"event"`
	StepName            string              `json:"step_name"`
	PipelineExecutionID string              `json:"pipeline_execution_id"`
	NextSteps           []flowpipe.NextStep `json:"next_steps"`
}

func (e *StepForEachPlanned) GetEvent() *Event {
	return e.Event
}

func (e *StepForEachPlanned) HandlerName() string {
	return HandlerStepForEachPlanned
}

func NewStepForEachPlannedFromStepForEachPlan(cmd *StepForEachPlan, nextSteps []flowpipe.NextStep) *StepForEachPlanned {
	return &StepForEachPlanned{
		Event:               cmd.Event,
		PipelineExecutionID: cmd.PipelineExecutionID,
		StepName:            cmd.StepName,
		NextSteps:           nextSteps,
	}
}
