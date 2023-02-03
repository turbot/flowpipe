package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type PipelineLoadHandler CommandHandler

func (h PipelineLoadHandler) HandlerName() string {
	return "command.pipeline_load"
}

func (h PipelineLoadHandler) NewCommand() interface{} {
	return &event.PipelineLoad{}
}

func (h PipelineLoadHandler) Handle(ctx context.Context, c interface{}) error {
	cmd := c.(*event.PipelineLoad)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	//defn, err := PipelineDefinition(cmd.Name)
	defn, err := PipelineDefinition("my_pipeline_0")
	if err != nil {
		e := event.PipelineRunFailed{
			RunID:        cmd.RunID,
			SpanID:       cmd.SpanID,
			CreatedAt:    time.Now(),
			ErrorMessage: err.Error(),
		}
		return h.EventBus.Publish(ctx, &e)
	}

	e := event.PipelineLoaded{
		RunID:     cmd.RunID,
		SpanID:    cmd.SpanID,
		CreatedAt: time.Now(),
		Pipeline:  *defn,
	}
	return h.EventBus.Publish(ctx, &e)
}

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
