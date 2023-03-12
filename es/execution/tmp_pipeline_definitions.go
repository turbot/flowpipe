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
		"series_of_for_loop_steps": {
			Type: "pipeline",
			Name: "series_of_for_loop_steps",
			Steps: map[string]*pipeline.PipelineStep{
				"sleep_1": {
					Type:      "sleep",
					Name:      "sleep_1",
					DependsOn: []string{},
					For:       `["1s", "2s", "300ms", "600ms"]`,
					Input:     `{"duration": "{{.each.value}}"}`,
				},
				"http_1": {
					Type:      "http_request",
					Name:      "http_1",
					DependsOn: []string{"sleep_1"},
					For:       `["http://api.open-notify.org/astros.json", "http://api.open-notify.org/iss-now.json"]`,
					Input:     `{"url": "{{.each.value}}"}`,
				},
			},
		},
		"simple_parallel": {
			Type: "pipeline",
			Name: "simple_parallel",
			Steps: map[string]*pipeline.PipelineStep{
				"http_1": {
					Type:  "http_request",
					Name:  "http_1",
					Input: `{"url": "http://api.open-notify.org/astros.json"}`,
				},
				"sleep_1": {
					Type:  "sleep",
					Name:  "sleep_1",
					Input: `{"duration": "2s"}`,
				},
			},
		},
		"for_loop_using_http_request_body_json": {
			Type: "pipeline",
			Name: "for_loop_using_map",
			Steps: map[string]*pipeline.PipelineStep{
				"astros": {
					Type:  "http_request",
					Name:  "astros",
					Input: `{"url": "http://api.open-notify.org/astros.json"}`,
				},
				"echo_astros": {
					Type:      "exec",
					Name:      "echo_astros",
					DependsOn: []string{"astros"},
					For:       `[{{ range $i, $person := .astros.body_json.people }}{{if $i}}, {{end}}"{{$person.name}}"{{ end }}]`,
					Input:     `{"command": "echo '{{.each.value}}'"}`,
				},
			},
		},
		"for_loop_using_map": {
			Type: "pipeline",
			Name: "for_loop_using_map",
			Steps: map[string]*pipeline.PipelineStep{
				"echo_map": {
					Type:  "exec",
					Name:  "echo_map",
					For:   `{"foo": 1, "bar": "baz"}`,
					Input: `{"command": "echo '{{.each.key}}={{.each.value}}'"}`,
				},
				"echo_array": {
					Type:  "exec",
					Name:  "echo_array",
					For:   `["foo", "bar"]`,
					Input: `{"command": "echo '{{.each.key}}={{.each.value}}'"}`,
				},
			},
		},
		"complex_sequence": {
			Type: "pipeline",
			Name: "complex_sequence",
			Steps: map[string]*pipeline.PipelineStep{
				"query_accounts": {
					Type:  "query",
					Name:  "query_accounts",
					Input: `{"sql": "select account_id, title from aws_account"}`,
				},
				"exec_1": {
					Type:      "exec",
					Name:      "exec_1",
					DependsOn: []string{"sleep_1"},
					Input:     `{"command": "ls"}`,
				},
				"sleep_1": {
					Type:      "sleep",
					Name:      "sleep_1",
					DependsOn: []string{"query_accounts"},
					For:       `["300ms", "600ms"]`,
					Input:     `{"duration": "{{.each.value}}"}`,
				},
				"pipeline_a": {
					Type:      "pipeline",
					Name:      "pipeline_a",
					DependsOn: []string{"sleep_1"},
					Input:     `{"name": "my_pipeline_1"}`,
				},
				"pipeline_b": {
					Type:      "pipeline",
					Name:      "pipeline_b",
					DependsOn: []string{"pipeline_a"},
					Input:     `{"name": "my_pipeline_1"}`,
				},
			},
		},
		"for_loop_for_parallel_exec": {
			Type: "pipeline",
			Name: "for_loop_for_parallel_exec",
			Steps: map[string]*pipeline.PipelineStep{
				"exec_1": {
					Type:  "exec",
					Name:  "exec_1",
					For:   `["pwd","ls","ls /crap","uname -a"]`,
					Input: `{"command": "{{.each.value}}"}`,
				},
			},
		},
		"pass_data_between_steps": {
			Type: "pipeline",
			Name: "pass_data_between_steps",
			Steps: map[string]*pipeline.PipelineStep{
				"list_root_dir": {
					Type:  "exec",
					Name:  "list_root_dir",
					Input: `{"command": "ls /"}`,
				},
				"list_each_subdir_of_root_dir": {
					Type:      "exec",
					Name:      "list_each_subdir_of_root_dir",
					DependsOn: []string{"list_root_dir"},
					For:       `[{{range $i, $e := .list_root_dir.stdout_lines}}{{ if $i }}, {{end}}"{{$e}}"{{end}}]`,
					Input:     `{"command": "ls /{{.each.value}}"}`,
				},
			},
		},
		"chained_steampipe_queries": {
			Type: "pipeline",
			Name: "chained_steampipe_queries",
			Steps: map[string]*pipeline.PipelineStep{
				"accounts": {
					Type:  "query",
					Name:  "accounts",
					Input: `{"sql": "select account_id, title from aws_account"}`,
				},
				"account_details": {
					Type:      "query",
					Name:      "account_details",
					DependsOn: []string{"accounts"},
					For:       `[{{range $i, $row := .accounts.rows}}{{ if $i }}, {{end}}{"account_id":"{{$row.account_id}}","title":"{{$row.title}}"}{{end}}]`,
					Input:     `{"sql":"select * from aws_account where account_id = '{{ .each.value.account_id }}'"}`,
				},
			},
		},
		"chained_input": {
			Type: "pipeline",
			Name: "chained_input",
			Steps: map[string]*pipeline.PipelineStep{
				"accounts": {
					Type:  "query",
					Name:  "accounts",
					Input: `{"sql": "select count(*) from aws_account"}`,
				},
				"echo_count": {
					Type:      "exec",
					Name:      "echo_count",
					DependsOn: []string{"accounts"},
					Input:     `{"command": "echo {{ (index .accounts.rows 0).count }}"}`,
				},
			},
		},
		"call_pipelines_in_for_loop": {
			Type: "pipeline",
			Name: "call_pipeline_in_for_loop",
			Steps: map[string]*pipeline.PipelineStep{
				"pipeline_caller": {
					Type:  "pipeline",
					Name:  "pipeline_caller",
					For:   `["chained_input", "chained_steampipe_queries"]`,
					Input: `{"name": "{{.each.value}}"}`,
				},
			},
		},
	}

	if d, ok := definitions[pe.Name]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("pipeline_not_found: %s", pe.Name)
}
