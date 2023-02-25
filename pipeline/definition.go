package pipeline

type Pipeline struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Steps    []PipelineStep         `json:"steps"`
	Parallel bool                   `json:"parallel"`
	Input    map[string]interface{} `json:"input"`
}

type PipelineStep struct {
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Input     map[string]interface{} `json:"input"`
	DependsOn []int                  `json:"depends_on"`
}
