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

	err := viper.BindPFlag("pipeline.dir", serviceStartCmd.Flags().Lookup("pipeline-dir"))
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
