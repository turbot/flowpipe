package fplog

import (
	"context"
)

type uiLoggerContextKey struct{}

type UILogger struct{}

/*
func (UILogger) Command(ci any) {
	c := ci.(command.Command)
}

func (UILogger) Event(ei any) {
	e := ei.(event.Command)
}
*/

func ContextWithLogger(ctx context.Context) context.Context {
	uiLogger := &UILogger{}
	return context.WithValue(ctx, uiLoggerContextKey{}, uiLogger)
}

func Logger(ctx context.Context) *UILogger {
	return ctx.Value(uiLoggerContextKey{}).(*UILogger)
}
