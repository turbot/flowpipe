package types

type Input map[string]interface{}

// Output is the output from a pipeline.
type Output map[string]interface{}

func (o *Output) Get(key string) interface{} {
	return (*o)[key]
}

type Pipeline struct {
	Type     string                   `json:"type"`
	Name     string                   `json:"name"`
	Steps    map[string]*PipelineStep `json:"steps"`
	Parallel bool                     `json:"parallel"`
	Args     Input                    `json:"args"`
	Output   string                   `json:"output,omitempty"`
}

type PipelineStep struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Input     string   `json:"input_template"`
	DependsOn []string `json:"depends_on"`
	For       string   `json:"for,omitempty"`
}

type ListPipelineResponse struct {
	Items     []Pipeline `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}
