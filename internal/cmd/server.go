package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/docker"
	serviceConfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Args:  cobra.NoArgs,
		Short: "Start the Flowpipe server",
		Run:   startServerFunc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {

			// TODO KAI look at whether this is really needed
			serviceConfig.Initialize()
			return nil
		},
	}

	cmdconfig.
		OnCmd(cmd).
		AddFilepathFlag(constants.ArgModLocation, ".", "The directory to load pipelines from. Defaults to the current directory.").
		AddIntFlag(constants.ArgPort, localconstants.DefaultServerPort, "Server port.").
		AddStringFlag(constants.ArgListen, localconstants.DefaultListen, "listen address port.").
		AddBoolFlag(constants.ArgNoScheduler, false, "Disable the scheduler.").
		AddBoolFlag(constants.ArgRetainArtifacts, false, "Retains Docker container artifacts for container step. [EXPERIMENTAL]").
		AddBoolFlag(constants.ArgInput, true, "Enable interactive prompts")

	return cmd
}

func startServerFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		error_helpers.FailOnError(docker.Initialize(ctx))

		// start manager, passing server config
		// (this will ensure manager starts API, ES, Scheduling and docker services
		m, err := manager.NewManager(ctx,
			manager.WithServerConfig(viper.GetString(constants.ArgListen), viper.GetInt(constants.ArgPort)),
		).Start()
		error_helpers.FailOnError(err)

		// Block until we receive a signal
		m.InterruptHandler()
	}
}
