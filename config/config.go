package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
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

// WithFlags parses the command line, and populates the configuration.
func WithFlags() ConfigOption {
	return func(c *Configuration) error {
		if flag.Parsed() {
			return fmt.Errorf("command-line flags already parsed")
		}

		flag.StringVar(&c.ConfigPath, "config-path", "~/.flowpipe/flowpipe.yaml", "Location of config file")

		flag.Usage = func() {
			//nolint:forbidigo // TODO
			fmt.Fprintf(os.Stderr, "\n%s\n\n", "Pipelines and workflows for DevSecOps.")
			//nolint:forbidigo // TODO
			fmt.Fprintf(os.Stderr, "Usage: %s [flags] <data directory>\n", "flowpipe")
			flag.PrintDefaults()
		}

		flag.Parse()

		/*
			if showVersion {
				msg := fmt.Sprintf("%s %s %s %s %s sqlite%s (commit %s, branch %s, compiler %s)",
					name, build.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), build.SQLiteVersion,
					build.Commit, build.Branch, runtime.Compiler)
				errorExit(0, msg)
			}
		*/

		// Ensure, if set explicitly, that reap times are not too low.
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "raft-reap-node-timeout" || f.Name == "raft-reap-read-only-node-timeout" {
				d, err := time.ParseDuration(f.Value.String())
				if err != nil {
					errorExit(1, fmt.Sprintf("failed to parse duration: %s", err.Error()))
				}
				if d <= 0 {
					errorExit(1, fmt.Sprintf("-%s must be greater than 0", f.Name))
				}
			}
		})

		return nil
	}
}

func errorExit(code int, msg string) {
	if code != 0 {
		//nolint:forbidigo // TODO
		fmt.Fprintf(os.Stderr, "fatal: ")
	}
	//nolint:forbidigo // TODO
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(code)
}
