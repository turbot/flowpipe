package handler

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Finished EventHandler

func (h Finished) HandlerName() string {
	return "handler.finished"
}

func (Finished) NewEvent() interface{} {
	return &event.Finished{}
}

func (h Finished) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.Finished)
	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)
	return nil
}
