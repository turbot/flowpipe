package main

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/cmd"
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

	appInit()

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

// set app specific constants defined in pipe-fittings
func appInit() {

	app_specific.WorkspaceDataDir = ".flowpipe"
	app_specific.ModFileName = "mod.hcl"
	app_specific.DefaultVarsFileName = "flowpipe.pvars"
	app_specific.DefaultInstallDir = "~/.flowpipe"

	app_specific.ModDataExtension = ".hcl"
	app_specific.VariablesExtension = ".pvars"
	app_specific.AutoVariablesExtension = ".auto.pvars"
	app_specific.EnvInputVarPrefix = "P_VAR_"
	app_specific.AppName = "flowpipe"

	//// set the default install dir
	//installDir, err := files.Tildefy("~/.steampipe")
	//if err != nil {
	//	panic(err)
	//}
	//constants.DefaultInstallDir = installDir
	//constants.AppName = "steampipe"
	//constants.ClientConnectionAppNamePrefix = "steampipe_client"
	//constants.ServiceConnectionAppNamePrefix = "steampipe_service"
	//constants.ClientSystemConnectionAppNamePrefix = "steampipe_client_system"
	//constants.AppVersion = steampipe_version.SteampipeVersion
	//constants.DefaultWorkspaceDatabase = "local"
	//
	//constants.ModDataExtension = ".sp"
	//
	//// set the command pre and post hooks
	//cmdconfig.CustomPreRunHook = localcmdconfig.PreRunHook
	//cmdconfig.CustomPostRunHook = localcmdconfig.PostRunHook
}
