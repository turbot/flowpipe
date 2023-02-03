package handler

import (
	"context"
	"fmt"
	"os"

	"github.com/turbot/steampipe-pipelines/es/event"
)

type Stopped EventHandler

func (h Stopped) HandlerName() string {
	return "handler.stopped"
}

func (Stopped) NewEvent() interface{} {
	return &event.Stopped{}
}

func (h Stopped) Handle(ctx context.Context, ei interface{}) error {
	e := ei.(*event.Stopped)
	fmt.Printf("[%-20s] %v\n", h.HandlerName(), e)
	os.Exit(1)
	return nil
}
