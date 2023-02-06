package fplog

import (
	//nolint:depguard // can't get the exclude list to work in the config (not sure why this is easy enough)
	"context"
	"fmt"

	"github.com/turbot/steampipe-pipelines/utils"
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

func NewProcessEventSourceLogger(RunID string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.OutputPaths = []string{fmt.Sprintf("logs/%s.jsonl", RunID)}
	return zap.Must(cfg.Build())
}

func ContextWithLogger(ctx context.Context) context.Context {
	sessionID := utils.Session(ctx)
	cfg := zap.NewProductionConfig()
	cfg.Sampling = nil
	cfg.OutputPaths = []string{fmt.Sprintf("logs/%s.jsonl", sessionID)}
	return context.WithValue(ctx, loggerContextKey{}, zap.Must(cfg.Build()))
}

func Logger(ctx context.Context) *zap.Logger {
	return ctx.Value(loggerContextKey{}).(*zap.Logger)
}
