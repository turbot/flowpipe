package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineRunLoad struct {
	IdentityID    string                 `json:"identity_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	PipelineName  string                 `json:"pipeline_name"`
	PipelineInput map[string]interface{} `json:"pipeline_input"`
	RunID         string                 `json:"run_id"`
}

type PipelineRunLoadHandler CommandHandler

func (h PipelineRunLoadHandler) HandlerName() string {
	return "pipeline.run.load"
}

func (h PipelineRunLoadHandler) NewCommand() interface{} {
	return &PipelineRunLoad{}
}

// Load the requested pipeline into the run. This captures the pipeline
// definition at a specific point in time, ensuring we can detect changes
// to the underlying definition while running.
//
// Events:
// - pipeline.run.loaded
//
// Errors:
// - pipeline_not_found
func (h PipelineRunLoadHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*PipelineRunLoad)

	fmt.Printf("[command] %s: %v\n", h.HandlerName(), c)

	defn, err := PipelineDefinition(cmd.PipelineName)

	if err != nil {
		e := event.PipelineRunFailed{
			IdentityID:    cmd.IdentityID,
			WorkspaceID:   cmd.WorkspaceID,
			PipelineInput: cmd.PipelineInput,
			RunID:         cmd.RunID,
			Timestamp:     time.Now(),
			ErrorMessage:  err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineRunLoaded{
		IdentityID:    cmd.IdentityID,
		WorkspaceID:   cmd.WorkspaceID,
		PipelineName:  cmd.PipelineName,
		PipelineInput: cmd.PipelineInput,
		RunID:         cmd.RunID,
		Timestamp:     time.Now(),
		Pipeline:      *defn,
	}
	return h.EventBus.Publish(ctx, &e)
}

func PipelineDefinition(name string) (*pipeline.Pipeline, error) {
	definitions := map[string]*pipeline.Pipeline{
		"my_pipeline_0": &pipeline.Pipeline{
			Name: "my_pipeline_0",
			Steps: []pipeline.PipelineStep{
				{Type: "http_request", Name: "get_data_0", Input: map[string]interface{}{"url": "http://api.open-notify.org/astros.json"}},
			},
		},
		"my_pipeline_1": &pipeline.Pipeline{
			Name: "my_pipeline_1",
			Steps: []pipeline.PipelineStep{
				{Type: "http_request", Name: "get_data_1", Input: map[string]interface{}{"url": "http://api.open-notify.org/astros.json"}},
				{Type: "http_request", Name: "get_data_2", Input: map[string]interface{}{"url": "http://api.open-notify.org/astros.json"}},
				{Type: "http_request", Name: "get_data_2", Input: map[string]interface{}{}},
			},
		},
	}
	if d, ok := definitions[name]; ok {
		return d, nil
	}
	return nil, errors.New(fmt.Sprintf("pipeline_not_found: %s", name))
}
