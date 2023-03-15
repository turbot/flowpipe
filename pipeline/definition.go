package pipeline

// PipelineInput is the input to a pipeline.
type PipelineInput map[string]interface{}

// Output is the output from a pipeline.
type Output map[string]interface{}

func (o *Output) Get(key string) interface{} {
	return (*o)[key]
}

// StepInput is the input to a step.
type StepInput map[string]interface{}

type Pipeline struct {
	Type     string                   `json:"type"`
	Name     string                   `json:"name"`
	Steps    map[string]*PipelineStep `json:"steps"`
	Parallel bool                     `json:"parallel"`
	Input    PipelineInput            `json:"input"`
	Output   string                   `json:"output,omitempty"`
}

type PipelineStep struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Input     string   `json:"input_template"`
	DependsOn []string `json:"depends_on"`
	For       string   `json:"for,omitempty"`
}
