package main

import (
	"context"

	"github.com/spf13/viper"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/cmd"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
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

	filepaths.PipesComponentWorkspaceDataDir = ".flowpipe"
	filepaths.PipesComponentModsFileName = "mod.hcl"
	filepaths.PipesComponentDefaultVarsFileName = "flowpipe.pvars"
	filepaths.PipesComponentDefaultInstallDir = "~/.flowpipe"

	constants.PipesComponentModDataExtension = ".hcl"
	constants.PipesComponentVariablesExtension = ".pvars"
	constants.PipesComponentAutoVariablesExtension = ".auto.pvars"
	constants.PipesComponentEnvInputVarPrefix = "P_VAR_"
	constants.PipesComponentAppName = "flowpipe"

	viper.SetDefault("main.version", version)
	viper.SetDefault("main.commit", commit)
	viper.SetDefault("main.date", date)
	viper.SetDefault("main.builtBy", builtBy)

	// Run the CLI
	err = cmd.RunCLI(ctx)
	if err != nil {
		panic(err)
	}
}
