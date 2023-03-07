package event

type runEvent struct {
	Event *Event `json:"event"`
}

type Loaded runEvent
type Queued runEvent

type Failed struct {
	Event        *Event `json:"event"`
	ErrorMessage string `json:"error_message"`
}

type Started struct {
	Event        *Event                 `json:"event"`
	PipelineName string                 `json:"pipeline_name"`
	Input        map[string]interface{} `json:"input"`
}

type Planned Started

type Stopped struct {
	Event  *Event                 `json:"event"`
	Output map[string]interface{} `json:"output"`
}

type PipelineStarted struct {
	Event *Event                 `json:"event"`
	Input map[string]interface{} `json:"input"`
}

type PipelinePlanned struct {
	Event           *Event                 `json:"event"`
	NextStepIndexes []int                  `json:"next_step_indexes"`
	Input           map[string]interface{} `json:"input"`
}

type PipelineFinished struct {
	Event  *Event                 `json:"event"`
	Output map[string]interface{} `json:"output"`
}

type PipelineStepStarted struct {
	Event     *Event `json:"event"`
	StepIndex int    `json:"step_index"`
}

type PipelineStepFinished struct {
	Event     *Event                 `json:"event"`
	StepIndex int                    `json:"step_index"`
	Output    map[string]interface{} `json:"output"`
}

type PipelineFailed struct {
	Event        *Event `json:"event"`
	ErrorMessage string `json:"error_message"`
}
