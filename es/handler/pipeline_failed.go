package handler

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/es/event"
)

type PipelineFailed EventHandler

func (h PipelineFailed) HandlerName() string {
	return "handler.pipeline_failed"
}

func (PipelineFailed) NewEvent() interface{} {
	return &event.PipelineFailed{}
}

func (h PipelineFailed) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineFailed)
	fmt.Println("pipeline_failed", e)
	return nil
}
