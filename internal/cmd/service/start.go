package service

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/docker"
	serviceConfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

func ServiceStartCmd(ctx context.Context) (*cobra.Command, error) {

	var serviceStartCmd = &cobra.Command{
		Use:  "start",
		Args: cobra.NoArgs,
		Run:  startManagerFunc(ctx),
	}

	serviceStartCmd.Flags().String(constants.ArgModLocation, ".", "The directory to load pipelines from. Defaults to the current directory.")
	serviceStartCmd.Flags().String(constants.ArgOutputDir, "~/.flowpipe/output", "The directory path to dump the snapshot file.")
	serviceStartCmd.Flags().String(constants.ArgLogDir, "~/.flowpipe/log", "The directory path to the log file for the execution.")
	serviceStartCmd.Flags().Bool(constants.ArgNoScheduler, false, "Disable the scheduler.")
	serviceStartCmd.Flags().Bool(constants.ArgRetainArtifacts, false, "Retains Docker container artifacts for container step. [EXPERIMENTAL]")
	serviceStartCmd.Flags().Bool(constants.ArgInput, true, "Enable interactive prompts")

	err := viper.BindPFlag(constants.ArgModLocation, serviceStartCmd.Flags().Lookup(constants.ArgModLocation))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	err = viper.BindPFlag(constants.ArgOutputDir, serviceStartCmd.Flags().Lookup(constants.ArgOutputDir))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	err = viper.BindPFlag(constants.ArgLogDir, serviceStartCmd.Flags().Lookup(constants.ArgLogDir))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	err = viper.BindPFlag(constants.ArgNoScheduler, serviceStartCmd.Flags().Lookup(constants.ArgNoScheduler))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	err = viper.BindPFlag(constants.ArgRetainArtifacts, serviceStartCmd.Flags().Lookup(constants.ArgRetainArtifacts))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	err = viper.BindPFlag(constants.ArgInput, serviceStartCmd.Flags().Lookup(constants.ArgInput))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	return serviceStartCmd, nil
}

func startManagerFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		err := docker.Initialize(ctx)
		if err != nil {
			error_helpers.FailOnError(err)
		}

		serviceConfig.Initialize(ctx)

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
