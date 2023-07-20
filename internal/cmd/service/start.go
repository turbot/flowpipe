package service

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	serviceConfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/flowpipe/internal/service/manager"
)

func ServiceStartCmd(ctx context.Context) (*cobra.Command, error) {

	var serviceStartCmd = &cobra.Command{
		Use:  "start",
		Args: cobra.NoArgs,
		Run:  startManagerFunc(ctx),
	}

	serviceStartCmd.Flags().String("pipeline-dir", "./flowpipe/pipelines", "The directory to load pipelines from")
	serviceStartCmd.Flags().String("work-dir", "./flowpipe/pipelines", "Working directory for pipeline execution")
	serviceStartCmd.Flags().String("output-dir", "./tmp", "The directory path to dump the snapshot file")
	serviceStartCmd.Flags().String("log-dir", "./tmp", "The directory path to the log file for the execution")

	err := viper.BindPFlag("pipeline.dir", serviceStartCmd.Flags().Lookup("pipeline-dir"))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("work.dir", serviceStartCmd.Flags().Lookup("work-dir"))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("output.dir", serviceStartCmd.Flags().Lookup("output-dir"))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("log.dir", serviceStartCmd.Flags().Lookup("log-dir"))
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
			panic(err)
		}

		// Start the manager
		err = m.Start()
		if err != nil {
			panic(err)
		}

		// Block until we receive a signal
		m.InterruptHandler()
	}
}
