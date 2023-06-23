package handler

import (
	"context"
	"os"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
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

	logger := fplog.Logger(ctx)

	logger.Info("stopped event handler")

	os.Exit(1)
	return nil
}
