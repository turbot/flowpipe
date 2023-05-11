package fplog

import (
	"context"
)

type loggerContextKey struct{}

func ContextWithLogger(ctx context.Context) context.Context {
	opts := []LoggerOption{}
	// TODO - how do we get these from the command line or config?
	opts = append(opts,
		WithLevelFromEnvironment(),
		WithFormatFromEnvironment(),
		WithColor(true),
	)
	lgr, err := NewLogger(ctx, opts...)
	if err != nil {
		panic(err)
	}
	return context.WithValue(ctx, loggerContextKey{}, lgr)
}

func Logger(ctx context.Context) *FlowpipeLogger {
	return ctx.Value(loggerContextKey{}).(*FlowpipeLogger)
}
