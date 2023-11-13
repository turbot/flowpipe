package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/flowpipe/internal/config"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

// variable used to assign the output mode flag
var outputMode types.OutputMode

// Build the cobra command that handles our command line tool.
func rootCommand(ctx context.Context) *cobra.Command {
	// Define our command
	rootCmd := &cobra.Command{
		Use:     app_specific.AppName,
		Short:   localconstants.FlowpipeShortDescription,
		Long:    localconstants.FlowpipeLongDescription,
		Version: viper.GetString("main.version"),
	}
	rootCmd.SetVersionTemplate("Flowpipe v{{.Version}}\n")

	c := config.GetConfigFromContext(ctx)

	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)

	rootCmd.Flags().StringVar(&c.ConfigPath, "config-path", "", "config file (default is $HOME/.flowpipe/flowpipe.yaml)")

	cmdconfig.
		OnCmd(rootCmd).
		// Flowpipe API
		AddPersistentStringFlag(localconstants.ArgApiHost, "http://localhost", "API server host").
		AddPersistentIntFlag(localconstants.ArgApiPort, 7103, "API server port").
		AddPersistentBoolFlag(localconstants.ArgTlsInsecure, false, "Skip TLS verification").
		// Common (steampipe, flowpipe) flags
		AddPersistentFilepathFlag(constants.ArgInstallDir, app_specific.DefaultInstallDir, "Path to the Config Directory").
		AddPersistentFilepathFlag(constants.ArgModLocation, cwd, "Path to the workspace working directory").
		// Define the CLI flag parameters for wrapped enum flag.
		AddPersistentVarFlag(enumflag.New(&outputMode, constants.ArgOutput, types.OutputModeIds, enumflag.EnumCaseInsensitive),
			constants.ArgOutput,
			"Output format; one of: table, yaml, json")

	// disable auto completion generation, since we don't want to support
	// powershell yet - and there's no way to disable powershell in the default generator
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// add all the subcommands
	rootCmd.AddCommand(serviceCmd())
	rootCmd.AddCommand(pipelineCmd())
	rootCmd.AddCommand(triggerCmd())
	rootCmd.AddCommand(processCmd())
	rootCmd.AddCommand(modCmd())

	return rootCmd
}
