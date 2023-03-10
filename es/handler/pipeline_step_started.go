package handler

import (
	"context"

	"github.com/pkg/errors"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
)

type PipelineStepStarted EventHandler

func (h PipelineStepStarted) HandlerName() string {
	return "handler.pipeline_step_started"
}

func (PipelineStepStarted) NewEvent() interface{} {
	return &event.PipelineStepStarted{}
}

func (h PipelineStepStarted) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineStepStarted)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	stepDefn, err := ex.StepDefinition(e.StepExecutionID)
	if err != nil {
		return err
	}

	switch stepDefn.Type {
	case "pipeline":
		cmd, err := event.NewPipelineQueue(event.ForPipelineStepStartedToPipelineQueue(e))
		if err != nil {
			return err
		}
		return h.CommandBus.Send(ctx, &cmd)
	case "sleep":
		// TODO - implement
	default:
		return errors.Errorf("step type cannot be started: %s", stepDefn.Type)
	}

	return nil
}
