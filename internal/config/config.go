package config

import (
	"context"
)

// Configuration represents the configuration as set by command-line flags.
// All variables will be set, unless explicitly noted.
type Configuration struct {
	ctx        context.Context
	ConfigPath string
}

// ConfigOption defines a type of function to configures the Config.
type ConfigOption func(*Configuration) error

// NewConfig creates a new Config.
func NewConfig(ctx context.Context, opts ...ConfigOption) (*Configuration, error) {
	// Defaults
	c := &Configuration{
		ctx: ctx,
	}
	// Set options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}
