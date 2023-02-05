package command

import (
	"fmt"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

func PipelineDefinition(name string) (*pipeline.Pipeline, error) {
	definitions := map[string]*pipeline.Pipeline{
		"my_pipeline_0": {
			Type: "pipeline",
			Name: "my_pipeline_0",
			Steps: []pipeline.PipelineStep{
				{Type: "exec", Name: "exec_1", Input: map[string]interface{}{"command": "ls"}},
				{Type: "http_request", Name: "http_1", Input: map[string]interface{}{"url": "http://api.open-notify.org/astros.json"}},
			},
		},
		"my_pipeline_1": {
			Type: "pipeline",
			Name: "my_pipeline_1",
			Steps: []pipeline.PipelineStep{
				{Type: "http_request", Name: "http_1", Input: map[string]interface{}{"url": "http://api.open-notify.org/astros.json"}},
				//{Type: "pipeline", Name: "pipeline_1", Input: map[string]interface{}{"name": "my_pipeline_0"}},
			},
		},
	}
	if d, ok := definitions[name]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("pipeline_not_found: %s", name)
}
