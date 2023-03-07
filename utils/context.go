package utils

import (
	"context"
)

type SessionContextKey struct{}

// ContextWithSession returns a new context with a new session ID.
func ContextWithSession(ctx context.Context) context.Context {
	sessionID := NewSessionID()
	return ContextWithSessionID(ctx, sessionID)
}

// ContextWithSessionID returns a new context with the given session ID.
// This is useful for testing.
func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionContextKey{}, sessionID)
}

// Session returns the session ID from the context.
func Session(ctx context.Context) string {
	if v := ctx.Value(SessionContextKey{}); v != nil {
		s := v.(string)
		return s
	}
	panic("No session in context")
}
