package handler

import (
	"context"
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
	//e := ei.(*event.Stopped)
	os.Exit(1)
	return nil
}
