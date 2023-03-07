package config

import "context"

// ConfigContextKey is the key used to store the config in the context.
type ConfigContextKey struct{}

// Config returns the session ID from the context.
func Get(ctx context.Context) *Config {
	if v := ctx.Value(ConfigContextKey{}); v != nil {
		return v.(*Config)
	}
	panic("No config in context")
}

// ContextWithSession returns a new context with a new session ID.
func Set(ctx context.Context, c *Config) context.Context {
	return context.WithValue(ctx, ConfigContextKey{}, c)
}
