package service

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	serviceStartCmd.Flags().String(constants.ArgPipelineDir, ".", "The directory to load pipelines from. Defaults to the current directory.")
	serviceStartCmd.Flags().String(constants.ArgWorkDir, "./flowpipe/pipelines", "Working directory for pipeline execution")
	serviceStartCmd.Flags().String(constants.ArgOutputDir, "~/.flowpipe/output", "The directory path to dump the snapshot file")
	serviceStartCmd.Flags().String(constants.ArgLogDir, "~/.flowpipe/log", "The directory path to the log file for the execution")

	pipelineDir, err := serviceStartCmd.Flags().GetString(constants.ArgPipelineDir)
	if err != nil {
		return nil, err
	}

	// Check if the specified file exists in the pipeline directory.
	filePath := fmt.Sprintf("%s/mod.hcl", pipelineDir)
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("mod.hcl not found in the current directory")
	}

	err = viper.BindPFlag(constants.ArgPipelineDir, serviceStartCmd.Flags().Lookup(constants.ArgPipelineDir))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag(constants.ArgWorkDir, serviceStartCmd.Flags().Lookup(constants.ArgWorkDir))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag(constants.ArgOutputDir, serviceStartCmd.Flags().Lookup(constants.ArgOutputDir))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag(constants.ArgLogDir, serviceStartCmd.Flags().Lookup(constants.ArgLogDir))
	if err != nil {
		log.Fatal(err)
	}

	return serviceStartCmd, nil
}

func startManagerFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

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
