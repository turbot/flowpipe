package event

type StepForEachPlan struct {
	Event               *Event `json:"event"`
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepName            string `json:"step_name"`
}

func (e *StepForEachPlan) GetEvent() *Event {
	return e.Event
}

func (e *StepForEachPlan) HandlerName() string {
	return CommandStepForEachPlan
}

func NewStepForEachPlanFromPipelinePlanned(e *PipelinePlanned, stepName string) *StepForEachPlan {
	return &StepForEachPlan{
		Event:               e.Event,
		PipelineExecutionID: e.PipelineExecutionID,
		StepName:            stepName,
	}
}

func NewStepForEachPlanFromPipelineStepFinished(e *StepFinished, stepName string) *StepForEachPlan {
	return &StepForEachPlan{
		Event:               e.Event,
		PipelineExecutionID: e.PipelineExecutionID,
		StepName:            stepName,
	}
}
