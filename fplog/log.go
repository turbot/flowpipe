package fplog

import (
	//nolint:depguard // can't get the exclude list to work in the config (not sure why this is easy enough)
	"context"

	"go.uber.org/zap"
)

type loggerContextKey struct{}

func NewLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Unable to create zap logger")
	}
	return logger
}

func ContextWithLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, NewLogger())
}

func Logger(ctx context.Context) *zap.Logger {
	return ctx.Value(loggerContextKey{}).(*zap.Logger)
}
