package cmd

import (
	"github.com/spf13/cobra"
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
		Short: "Server commands",
	}

	cmd.AddCommand(serverStartCmd())

	return cmd
}

func serverStartCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use: "start",
		Run: startManagerFunc(),
	}

	cmdconfig.
		OnCmd(cmd).
		AddFilepathFlag(constants.ArgModLocation, ".", "The directory to load pipelines from. Defaults to the current directory.").
		AddFilepathFlag(constants.ArgOutputDir, "~/.flowpipe/output", "The directory path to dump the snapshot file.").
		AddFilepathFlag(constants.ArgLogDir, "~/.flowpipe/log", "The directory path to the log file for the execution.").
		AddIntFlag(constants.ArgPort, localconstants.DefaultServerPort, "Server port.").
		// TODO KAI check default
		AddIntFlag(constants.ArgPortHttps, localconstants.DefaultServerPort, "HTTPS server port.").
		AddIntFlag(constants.ArgListen, localconstants.DefaultServerPort, "HTTPS server port.").
		AddBoolFlag(constants.ArgNoScheduler, false, "Disable the scheduler.").
		AddBoolFlag(constants.ArgRetainArtifacts, false, "Retains Docker container artifacts for container step. [EXPERIMENTAL]").
		AddBoolFlag(constants.ArgInput, true, "Enable interactive prompts")

	return cmd
}

func startManagerFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		err := docker.Initialize(ctx)
		if err != nil {
			error_helpers.FailOnError(err)
		}

		serviceConfig.Initialize()

		m, err := manager.NewManager(ctx)

		if err != nil {
			error_helpers.FailOnError(err)
		}

		err = m.Initialize()
		if err != nil {
			error_helpers.FailOnError(err)
		}

		err = m.Start()
		if err != nil {
			error_helpers.FailOnError(err)
		}

		// Block until we receive a signal
		m.InterruptHandler()
	}
}
