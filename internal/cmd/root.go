package cmd

import (
	"context"
	"github.com/turbot/flowpipe/internal/sanitize"
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

// TODO where should this be defined
var sanitizer = sanitize.NewSanitizer(sanitize.SanitizerOptions{
	ExcludeFields: []string{
		"pipeline_execution_id",
		"pipeline_name",
		"mod",
		"step_type",
		"name",
		"pipeline_name",
	},
	ExcludePatterns: []string{
		"echo_one",
	},
})

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

	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)

	cmdconfig.
		OnCmd(rootCmd).
		// Flowpipe API
		AddPersistentStringFlag(constants.ArgHost, "", "API server host, including the port number").
		AddPersistentBoolFlag(localconstants.ArgTlsInsecure, false, "Skip TLS verification").
		// Common (steampipe, flowpipe) flags
		AddPersistentFilepathFlag(constants.ArgInstallDir, app_specific.DefaultInstallDir, "Path to the Config Directory").
		AddPersistentFilepathFlag(constants.ArgModLocation, cwd, "Path to the workspace working directory").
		AddPersistentStringFlag(constants.ArgWorkspaceProfile, "default", "The workspace to use").
		// Define the CLI flag parameters for wrapped enum flag.
		AddPersistentVarFlag(enumflag.New(&outputMode, constants.ArgOutput, types.OutputModeIds, enumflag.EnumCaseInsensitive),
			constants.ArgOutput,
			"Output format; one of: table, yaml, json")

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
