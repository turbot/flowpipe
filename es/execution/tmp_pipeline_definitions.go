package execution

import (
	"fmt"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

func (ex *Execution) PipelineDefinition(pipelineExecutionID string) (*pipeline.Pipeline, error) {
	pe, ok := ex.PipelineExecutions[pipelineExecutionID]
	if !ok {
		return nil, fmt.Errorf("pipeline execution %s not found", pipelineExecutionID)
	}
	// TODO - total hack that this is hardcoded here
	definitions := map[string]*pipeline.Pipeline{
		"my_pipeline_0": {
			Type: "pipeline",
			Name: "my_pipeline_0",
			Steps: map[string]*pipeline.PipelineStep{
				"sleep_1": {
					Type:      "sleep",
					Name:      "sleep_1",
					For:       `[{"duration": "1s"}, {"duration": "2s"}, {"duration": "300ms"}, {"duration": "600ms"}]`,
					DependsOn: []string{},
					Input:     map[string]interface{}{"duration": "2s"},
				},
				"http_1": {
					Type:      "http_request",
					Name:      "http_1",
					DependsOn: []string{"sleep_1"},
					For:       `[{"url": "http://api.open-notify.org/astros.json"}, {"url": "http://api.open-notify.org/iss-now.json"}]`,
				},
			},
		},
		"my_pipeline_1": {
			Type: "pipeline",
			Name: "my_pipeline_1",
			Steps: map[string]*pipeline.PipelineStep{
				"http_1": {
					Type:  "http_request",
					Name:  "http_1",
					Input: map[string]interface{}{"url": "http://api.open-notify.org/astros.json"},
				},
				"sleep_1": {
					Type:      "sleep",
					Name:      "sleep_1",
					DependsOn: []string{},
					Input:     map[string]interface{}{"duration": "2s"},
				},
			},
		},
		"my_pipeline_2": {
			Type: "pipeline",
			Name: "my_pipeline_2",
			Steps: map[string]*pipeline.PipelineStep{
				"query_accounts": {
					Type:  "query",
					Name:  "query_accounts",
					Input: map[string]interface{}{"sql": "select account_id, title from aws_account"},
				},
				"exec_1": {
					Type:      "exec",
					Name:      "exec_1",
					DependsOn: []string{"sleep_1"},
					Input:     map[string]interface{}{"command": "ls"},
				},
				"sleep_1": {
					Type:      "sleep",
					Name:      "sleep_1",
					For:       `[{"duration": "1s"}, {"duration": "2s"}]`,
					DependsOn: []string{"query_accounts"},
					Input:     map[string]interface{}{"duration": "2s"},
				},
				"pipeline_a": {
					Type:      "pipeline",
					Name:      "pipeline_a",
					DependsOn: []string{"sleep_1"},
					Input:     map[string]interface{}{"name": "my_pipeline_1"},
				},
				"pipeline_b": {
					Type:      "pipeline",
					Name:      "pipeline_b",
					DependsOn: []string{"pipeline_a"},
					Input:     map[string]interface{}{"name": "my_pipeline_1"},
				},
			},
		},
		"my_pipeline_3": {
			Type: "pipeline",
			Name: "my_pipeline_3",
			Steps: map[string]*pipeline.PipelineStep{
				"exec_1": {
					Type: "exec",
					Name: "exec_1",
					For:  `[{"command":"pwd"},{"command":"ls"},{"command":"ls /crap"},{"command":"uname -a"}]`,
				},
			},
		},
		"pass_data_between_steps": {
			Type: "pipeline",
			Name: "my_pipeline_3",
			Steps: map[string]*pipeline.PipelineStep{
				"list_root_dir": {
					Type:  "exec",
					Name:  "list_root_dir",
					Input: pipeline.StepInput{"command": "ls /"},
				},
				"list_each_subdir_of_root_dir": {
					Type:      "exec",
					Name:      "list_each_subdir_of_root_dir",
					DependsOn: []string{"list_root_dir"},
					For:       `[{{range $i, $e := .list_root_dir.stdout_lines}}{{ if $i }}, {{end}}{"command":"ls /{{$e}}"}{{end}}]`,
				},
			},
		},
	}

	if d, ok := definitions[pe.Name]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("pipeline_not_found: %s", pe.Name)
}
