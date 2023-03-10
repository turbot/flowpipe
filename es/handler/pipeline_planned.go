package handler

import (
	"context"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
)

type PipelinePlanned EventHandler

func (h PipelinePlanned) HandlerName() string {
	return "handler.pipeline_planned"
}

func (PipelinePlanned) NewEvent() interface{} {
	return &event.PipelinePlanned{}
}

func (h PipelinePlanned) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelinePlanned)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		e := event.PipelineFailed{
			Event:        event.NewFlowEvent(e.Event),
			ErrorMessage: err.Error(),
		}
		return h.CommandBus.Send(ctx, &e)
	}

	// Convenience
	pe := ex.PipelineExecutions[e.PipelineExecutionID]

	if len(e.NextSteps) == 0 {

		// PRE: No new steps to execute, so the planner should just check to see if
		// all existing steps are complete.

		complete := true
		for _, stepStatus := range pe.StepStatus {
			if stepStatus.Status != "finished" && stepStatus.Status != "failed" {
				complete = false
				break
			}
		}

		if complete {
			cmd, err := event.NewPipelineFinish(event.ForPipelinePlannedToPipelineFinish(e))
			if err != nil {
				return err
			}
			return h.CommandBus.Send(ctx, &cmd)
		}

		return nil
	}

	// PRE: The planner has told us what steps to run next, our job is to start them

	for _, stepName := range e.NextSteps {
		stepDefn := defn.Steps[stepName]
		cmd, err := event.NewPipelineStepStart(event.ForPipelinePlanned(e), event.WithStep(stepName, stepDefn.Input))
		if err != nil {
			return err
		}
		if err := h.CommandBus.Send(ctx, &cmd); err != nil {
			return err
		}
	}

	return nil
}
