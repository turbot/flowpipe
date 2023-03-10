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

type PipelineFailed struct {
	Event        *Event `json:"event"`
	ErrorMessage string `json:"error_message"`
}
