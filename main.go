package main

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/files"
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

	installDir, err := files.Tildefy("~/.flowpipe")
	if err != nil {
		panic(err)
	}

	app_specific.AppName = "flowpipe"
	// TODO unify version logic with steampipe and powerpipe
	//app_specific.AppVersion
	app_specific.AutoVariablesExtension = ".auto.pvars"
	//app_specific.ClientConnectionAppNamePrefix
	//app_specific.ClientSystemConnectionAppNamePrefix
	app_specific.DefaultInstallDir = installDir
	app_specific.DefaultVarsFileName = "flowpipe.pvars"
	//app_specific.DefaultWorkspaceDatabase
	//app_specific.EnvAppPrefix
	app_specific.EnvInputVarPrefix = "P_VAR_"
	//app_specific.InstallDir
	app_specific.ModDataExtension = ".hcl"
	app_specific.ModFileName = "mod.hcl"
	app_specific.VariablesExtension = ".pvars"
	//app_specific.ServiceConnectionAppNamePrefix
	//app_specific.WorkspaceIgnoreFile
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	//app_specific.WorkspaceDataDir

}
