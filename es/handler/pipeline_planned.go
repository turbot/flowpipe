package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
	"github.com/turbot/steampipe-pipelines/pipeline"
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
			if stepStatus.Progress() < 100 {
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

		stepOutputs, err := ex.PipelineStepOutputs(e.PipelineExecutionID)
		if err != nil {
			e := event.PipelineFailed{
				Event:        event.NewFlowEvent(e.Event),
				ErrorMessage: err.Error(),
			}
			return h.CommandBus.Send(ctx, &e)
		}

		items := []pipeline.StepInput{}

		stepDefn := defn.Steps[stepName]
		if stepDefn.Input != "" {
			// Use go template with the step outputs to generate the items
			t, err := template.New("input").Parse(stepDefn.Input)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
			var itemsBuffer bytes.Buffer
			err = t.Execute(&itemsBuffer, stepOutputs)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
			fmt.Println(stepName, ".input = ", itemsBuffer.String())
			var item pipeline.StepInput
			err = json.Unmarshal(itemsBuffer.Bytes(), &item)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
			items = append(items, item)
		} else if stepDefn.For != "" {
			// Use go template with the step outputs to generate the items
			stepOutputs, err := ex.PipelineStepOutputs(e.PipelineExecutionID)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
			t, err := template.New("for").Parse(stepDefn.For)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
			var itemsBuffer bytes.Buffer
			err = t.Execute(&itemsBuffer, stepOutputs)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
			fmt.Println(stepName, ".for = ", itemsBuffer.String())
			err = json.Unmarshal(itemsBuffer.Bytes(), &items)
			if err != nil {
				e := event.PipelineFailed{
					Event:        event.NewFlowEvent(e.Event),
					ErrorMessage: err.Error(),
				}
				return h.CommandBus.Send(ctx, &e)
			}
		}

		for _, item := range items {
			// Start each step in parallel
			go func(stepName string, item pipeline.StepInput) {
				cmd, err := event.NewPipelineStepStart(event.ForPipelinePlanned(e), event.WithStep(stepName, item))
				if err != nil {
					e := event.PipelineFailed{
						Event:        event.NewFlowEvent(e.Event),
						ErrorMessage: err.Error(),
					}
					h.CommandBus.Send(ctx, &e)
					return
				}
				if err := h.CommandBus.Send(ctx, &cmd); err != nil {
					e := event.PipelineFailed{
						Event:        event.NewFlowEvent(e.Event),
						ErrorMessage: err.Error(),
					}
					h.CommandBus.Send(ctx, &e)
					return
				}
			}(stepName, item)
		}
	}

	return nil
}
