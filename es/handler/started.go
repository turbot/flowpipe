package handler

import (
	"context"

	"github.com/turbot/flowpipe/es/event"
)

type Started EventHandler

func (h Started) HandlerName() string {
	return "handler.started"
}

func (Started) NewEvent() interface{} {
	return &event.Started{}
}

func (h Started) Handle(ctx context.Context, ei interface{}) error {

	// Note: The mod is now listening for trigger events. It is stopped by a
	// Ctrl-C handler hooked to the Stop command.

	return nil
}
