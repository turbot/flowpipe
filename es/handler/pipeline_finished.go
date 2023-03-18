package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
)

type PipelineFinished EventHandler

func (h PipelineFinished) HandlerName() string {
	return "handler.pipeline_finished"
}

func (PipelineFinished) NewEvent() interface{} {
	return &event.PipelineFinished{}
}

func (h PipelineFinished) Handle(ctx context.Context, ei interface{}) error {

	e := ei.(*event.PipelineFinished)

	ex, err := execution.NewExecution(ctx, execution.WithEvent(e.Event))
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
	}

	parentStepExecution, err := ex.ParentStepExecution(e.PipelineExecutionID)
	if err != nil {
		return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
	}
	if parentStepExecution != nil {
		cmd, err := event.NewPipelineStepFinish(
			event.ForPipelineFinished(e),
			event.WithPipelineExecutionID(parentStepExecution.PipelineExecutionID),
			event.WithStepExecutionID(parentStepExecution.ID))
		if err != nil {
			return h.CommandBus.Send(ctx, event.NewPipelineFail(event.ForPipelineFinishedToPipelineFail(e, err)))
		}
		return h.CommandBus.Send(ctx, &cmd)
	} else {
		// Dump the final execution state
		jsonStr, _ := json.MarshalIndent(ex, "", "  ")
		fmt.Println(string(jsonStr))

		// Dump step outputs
		data, err := ex.PipelineData(e.PipelineExecutionID)
		if err != nil {
			fmt.Println(err)
		} else {
			jsonStr, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(jsonStr))
		}

		// Dump the snapshot
		snapshot, err := ex.Snapshot(e.PipelineExecutionID)
		if err != nil {
			fmt.Println(err)
		} else {
			jsonStr, _ := json.MarshalIndent(snapshot, "", "  ")
			_ = os.WriteFile("/Users/nathan/Downloads/pe.sps", jsonStr, 0644)
			//fmt.Println(string(jsonStr))
		}

	}

	return nil
}
