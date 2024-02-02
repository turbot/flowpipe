package main

import (
	"context"

	"github.com/spf13/viper"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/cmd"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

var (
	// These variables will be set by GoReleaser. We have them in main package because we put everything else in internal
	// and  I couldn't get Go Release to modify the internal packages
	version = "0.0.1-local.1"
	commit  = "none"
	date    = "unknown"
	builtBy = "local"
)

func main() {
	// Create a single, global context for the application
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	viper.SetDefault(constants.ArgProcessRetention, 604800) // 7 days
	viper.SetDefault("main.version", version)
	viper.SetDefault("main.commit", commit)
	viper.SetDefault("main.date", date)
	viper.SetDefault("main.builtBy", builtBy)

	localcmdconfig.SetAppSpecificConstants()
	log.SetDefaultLogger()
	cache.InMemoryInitialize(nil)

	// Run the CLI
	cmd.RunCLI(ctx)
}
