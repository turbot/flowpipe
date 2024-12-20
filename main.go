package main

import (
	"context"
	"os"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cmd"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/fperr"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/utils"
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
	utils.LogTime("main start")
	defer func() {
		utils.LogTime("main end")
		utils.DisplayProfileDataJsonl(os.Stderr)

		var err error
		if r := recover(); r != nil {
			err = helpers.ToError(r)
			error_helpers.ShowError(ctx, err)
			exitCode := fperr.GetExitCode(err, true)
			os.Exit(exitCode)
		}
	}()

	viper.SetDefault(constants.ArgProcessRetention, 604800) // 7 days
	viper.SetDefault("main.version", version)
	viper.SetDefault("main.commit", commit)
	viper.SetDefault("main.date", date)
	viper.SetDefault("main.builtBy", builtBy)

	localcmdconfig.SetAppSpecificConstants()
	log.SetDefaultLogger()

	// Run the CLI
	cmd.RunCLI(ctx)
}
