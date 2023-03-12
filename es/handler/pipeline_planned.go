package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
	}

	defn, err := ex.PipelineDefinition(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
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
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}
			return h.CommandBus.Send(ctx, &cmd)
		}

		return nil
	}

	// PRE: The planner has told us what steps to run next, our job is to start them

	for _, stepName := range e.NextSteps {

		stepOutputs, err := ex.PipelineStepOutputs(e.PipelineExecutionID)
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
		}

		forInputs := []interface{}{}

		stepDefn := defn.Steps[stepName]
		if stepDefn.For != "" {
			// Use go template with the step outputs to generate the items
			t, err := template.New("for").Parse(stepDefn.For)
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}
			var itemsBuffer bytes.Buffer
			err = t.Execute(&itemsBuffer, stepOutputs)
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}
			err = json.Unmarshal(itemsBuffer.Bytes(), &forInputs)
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}
			if len(forInputs) == 0 {
				// A for loop was defined, but it returned no items, so we can
				// skip this step
				return nil
			}
		}

		inputs := []pipeline.StepInput{}

		if stepDefn.Input == "" {
			// No input, so just use an empty input for each step execution.

			// There is always one input (e.g. no for loop). If the for loop had
			// no items, then we would have returned above.
			inputs = append(inputs, pipeline.StepInput{})

			// Add extra items if the for loop required them, skipping the one
			// we added already above.
			for i := 0; i < len(forInputs)-1; i++ {
				inputs = append(inputs, pipeline.StepInput{})
			}
		} else {
			// We have an input

			// Parse the input template once
			t, err := template.New("input").Parse(stepDefn.Input)
			if err != nil {
				return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
			}

			if stepDefn.For == "" {
				// No for loop

				var itemsBuffer bytes.Buffer
				err = t.Execute(&itemsBuffer, stepOutputs)
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				var input pipeline.StepInput
				err = json.Unmarshal(itemsBuffer.Bytes(), &input)
				if err != nil {
					return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
				}
				inputs = append(inputs, input)

			} else {
				// Create a step for each input in forInputs
				for i, forInput := range forInputs {

					// TODO - this updates the same map each time ... is that safe?
					var stepOutputsWithEach = stepOutputs
					stepOutputsWithEach["each"] = map[string]interface{}{"key": i, "value": forInput}

					var itemsBuffer bytes.Buffer
					err = t.Execute(&itemsBuffer, stepOutputsWithEach)
					if err != nil {
						return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					}
					var input pipeline.StepInput
					err = json.Unmarshal(itemsBuffer.Bytes(), &input)
					if err != nil {
						return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					}
					inputs = append(inputs, input)
				}

			}

		}

		for _, input := range inputs {
			// Start each step in parallel
			go func(stepName string, input pipeline.StepInput) {
				cmd, err := event.NewPipelineStepStart(event.ForPipelinePlanned(e), event.WithStep(stepName, input))
				if err != nil {
					h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					return
				}
				if err := h.CommandBus.Send(ctx, &cmd); err != nil {
					h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelinePlannedToPipelineFail(e, err)))
					return
				}
			}(stepName, input)
		}
	}

	return nil
}
