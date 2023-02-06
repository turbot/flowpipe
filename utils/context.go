package utils

import (
	"context"
)

type SessionContextKey struct{}

func ContextWithSession(ctx context.Context) context.Context {
	return context.WithValue(ctx, SessionContextKey{}, NewSessionID())
}

func Session(ctx context.Context) string {
	if v := ctx.Value(SessionContextKey{}); v != nil {
		s := v.(string)
		return s
	}
	panic("No session in context")
}
