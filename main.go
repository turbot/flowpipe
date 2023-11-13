package main

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/cmd"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
)

var (
	// This variables will be set by GoReleaser, put it in main package because we put everything else in internal and I couldn't get Go Releaser
	// to modify the internal package
	version = "0.0.1-local.1"
	commit  = "none"
	date    = "unknown"
	builtBy = "local"
)

func main() {
	// Create a single, global context for the application
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)
	ctx, err := config.ContextWithConfig(ctx)

	defer func() {
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()
	if err != nil {
		error_helpers.FailOnError(err)
	}

	cache.InMemoryInitialize(nil)

	localcmdconfig.SetAppSpecificConstants()

	// TODO kai can we pass these into SetAppSpecificConstants?
	//  look into namespacing of config
	viper.SetDefault("main.version", version)
	viper.SetDefault("main.commit", commit)
	viper.SetDefault("main.date", date)
	viper.SetDefault("main.builtBy", builtBy)

	// Run the CLI
	cmd.RunCLI(ctx)
}
