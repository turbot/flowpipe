package handler

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/es/event"
)

type PipelineCanceled EventHandler

func (h PipelineCanceled) HandlerName() string {
	return "handler.pipeline_canceled"
}

func (PipelineCanceled) NewEvent() interface{} {
	return &event.PipelineCanceled{}
}

func (h PipelineCanceled) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.PipelineCanceled)
	fmt.Println("pipeline_canceled", e)
	return nil
}
