package pipeline

type Pipeline struct {
	Name  string         `json:"name"`
	Steps []PipelineStep `json:"steps"`
}

type PipelineStep struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
