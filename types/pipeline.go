package types

type Input map[string]interface{}

// Output is the output from a pipeline.
type Output map[string]interface{}

func (o *Output) Get(key string) interface{} {
	return (*o)[key]
}

// The definition of a single Flowpipe Pipeline
type Pipeline struct {
	Type     string                   `json:"type"`
	Name     string                   `json:"name"`
	Steps    map[string]*PipelineStep `json:"steps"`
	Parallel bool                     `json:"parallel"`
	Args     Input                    `json:"args"`
	Output   string                   `json:"output,omitempty"`
}

func (p *Pipeline) GetType() string {
	return "pipeline"
}

type PipelineStep struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Input     string   `json:"input_template"`
	DependsOn []string `json:"depends_on"`
	For       string   `json:"for,omitempty"`
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []Pipeline `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}

func (l *ListPipelineResponse) Transform() *ListPipelineResponse {
	resources := []FlowpipeResource{}
	for _, item := range l.Items {
		resources = append(resources, item)
	}
	return resources
}
