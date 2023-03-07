package event

// Queue a mod for running in a given workspace context.
type Queue struct {
	// Base event structure
	Event *Event `json:"event"`
	// Host of the workspace. If empty, then assume localhost.
	CloudHost string `json:"host"`
	// The workspace context to use. May be a local workspace (e.g. default) or
	// a cloud workspace (e.g. e-gineer/scratch).
	Workspace string `json:"workspace"`
	// File system location where the mod is located, including pipeline
	// defintions.
	ModLocation string `json:"mod_location"`
}

// Load a mod for running in a given workspace context.
type Load struct {
	Event *Event `json:"event"`
}

type Plan Load

type Start Load

type Stop Load

type PipelineQueue struct {
	Event *Event `json:"event"`
	// Pipeline details
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type PipelineLoad struct {
	Event *Event `json:"event"`
}

type PipelineStart struct {
	Event        *Event                 `json:"event"`
	PipelineName string                 `json:"pipeline_name"`
	StepIndex    int                    `json:"step_index"`
	Input        map[string]interface{} `json:"input"`
}

type PipelinePlan PipelineStart

type PipelineFinish struct {
	Event *Event `json:"event"`
}

type PipelineStepStart PipelineStart
