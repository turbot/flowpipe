package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
)

// Configuration represents the configuration as set by command-line flags.
// All variables will be set, unless explicitly noted.
type Configuration struct {
	ctx        context.Context
	Viper      *viper.Viper
	ConfigPath string

	// TODO: Directory where log files will be written.
	LogDir string `json:"log_dir,omitempty"`
}

// ConfigOption defines a type of function to configures the Config.
type ConfigOption func(*Configuration) error

// NewConfig creates a new Config.
func NewConfig(ctx context.Context, opts ...ConfigOption) (*Configuration, error) {
	// Defaults
	c := &Configuration{
		ctx:   ctx,
		Viper: viper.New(),
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

func (c *Configuration) InitializeViper() error {

	// Convenience
	v := c.Viper

	if c.ConfigPath != "" {
		// User has provided a specific config file location, so use that.
		// We do not look in other (default) locations in this case.
		v.SetConfigFile(c.ConfigPath)
	} else {
		// Look for a config file in standard locations.
		// First, the current working directory.
		v.AddConfigPath(".")
		// Second, the user's home directory.
		v.AddConfigPath("$HOME/.flowpipe")

		// Set the base name of the config file, without the file extension.
		// This means they can use a variety of formats, like HCL or YAML or JSON.
		v.SetConfigName("flowpipe")
	}

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fperr.WrapWithMessage(err, "error reading config file")
		}
	}
	fplog.Logger(c.ctx).Debug("Using config file:", v.ConfigFileUsed())

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable FLOWPIPE_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix("FLOWPIPE")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to FLOWPIPE_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	return nil
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
