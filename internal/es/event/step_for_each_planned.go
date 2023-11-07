package event

import "github.com/turbot/pipe-fittings/modconfig"

type StepForEachPlanned struct {
	Event               *Event               `json:"event"`
	StepName            string               `json:"step_name"`
	PipelineExecutionID string               `json:"pipeline_execution_id"`
	NextSteps           []modconfig.NextStep `json:"next_steps"`
}

func (e *StepForEachPlanned) GetEvent() *Event {
	return e.Event
}

func (e *StepForEachPlanned) HandlerName() string {
	return "handler.step_for_each_planned"
}

func NewStepForEachPlannedFromStepForEachPlan(cmd *StepForEachPlan, nextSteps []modconfig.NextStep) *StepForEachPlanned {
	return &StepForEachPlanned{
		Event:               cmd.Event,
		PipelineExecutionID: cmd.PipelineExecutionID,
		StepName:            cmd.StepName,
		NextSteps:           nextSteps,
	}
}
