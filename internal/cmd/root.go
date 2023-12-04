package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
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
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				error_helpers.FailOnError(err)
			}
		},
	}
	rootCmd.SetVersionTemplate("Flowpipe v{{.Version}}\n")

	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)

	cmdconfig.
		OnCmd(rootCmd).
		// Flowpipe API
		AddPersistentStringFlag(constants.ArgHost, "", "API server host, including the port number").
		AddPersistentBoolFlag(localconstants.ArgTlsInsecure, false, "Skip TLS verification").
		AddPersistentStringSliceFlag(constants.ArgConfigPath, nil, "Comma separated list of paths to search for workspace files").
		// Common (steampipe, flowpipe) flags
		AddPersistentFilepathFlag(constants.ArgModLocation, cwd, "Path to the workspace working directory").
		AddPersistentStringFlag(constants.ArgWorkspaceProfile, "default", "The workspace to use").
		// Define the CLI flag parameters for wrapped enum flag.
		AddPersistentVarFlag(enumflag.New(&outputMode, constants.ArgOutput, types.OutputModeIds, enumflag.EnumCaseInsensitive),
			constants.ArgOutput,
			"Output format; one of: pretty, plain, yaml, json").
		AddPersistentStringSliceFlag(constants.ArgVarFile, nil, "Specify an .fpvar file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddPersistentStringArrayFlag(constants.ArgVariable, nil, "Specify the value of a variable")

	// disable auto completion generation, since we don't want to support
	// powershell yet - and there's no way to disable powershell in the default generator
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// add all the subcommands
	rootCmd.AddCommand(serverCmd())
	rootCmd.AddCommand(pipelineCmd())
	rootCmd.AddCommand(triggerCmd())
	rootCmd.AddCommand(processCmd())
	rootCmd.AddCommand(modCmd())

	return rootCmd
}
