package service

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	docker "github.com/turbot/flowpipe/internal/docker"
	serviceConfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
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
	serviceStartCmd.Flags().Bool(constants.ArgFunctions, false, "Enable function and container steps.")
	serviceStartCmd.Flags().Bool(constants.ArgNoScheduler, false, "Disable the scheduler.")

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

	err = viper.BindPFlag(constants.ArgFunctions, serviceStartCmd.Flags().Lookup(constants.ArgFunctions))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	err = viper.BindPFlag(constants.ArgNoScheduler, serviceStartCmd.Flags().Lookup(constants.ArgNoScheduler))
	if err != nil {
		error_helpers.FailOnError(err)
	}

	return serviceStartCmd, nil
}

func startManagerFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		if viper.GetBool(constants.ArgFunctions) {
			err := docker.Initialize(ctx)
			if err != nil {
				error_helpers.FailOnError(err)
			}
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
