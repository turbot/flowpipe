package event

import (
	"github.com/turbot/flowpipe/internal/resources"
)

type StepForEachPlanned struct {
	Event               *Event              `json:"event"`
	StepName            string              `json:"step_name"`
	PipelineExecutionID string               `json:"pipeline_execution_id"`
	NextSteps           []resources.NextStep `json:"next_steps"`
}

func (e *StepForEachPlanned) GetEvent() *Event {
	return e.Event
}

func (e *StepForEachPlanned) HandlerName() string {
	return HandlerStepForEachPlanned
}

func NewStepForEachPlannedFromStepForEachPlan(cmd *StepForEachPlan, nextSteps []resources.NextStep) *StepForEachPlanned {
	return &StepForEachPlanned{
		Event:               cmd.Event,
		PipelineExecutionID: cmd.PipelineExecutionID,
		StepName:            cmd.StepName,
		NextSteps:           nextSteps,
	}
}
