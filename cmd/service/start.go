package service

import (
	"context"

	"github.com/spf13/cobra"
	serviceConfig "github.com/turbot/flowpipe/service/config"
	"github.com/turbot/flowpipe/service/manager"
)

func ServiceStartCmd(ctx context.Context) (*cobra.Command, error) {

	var serviceStartCmd = &cobra.Command{
		Use:  "start",
		Args: cobra.NoArgs,
		Run:  startManagerFunc(ctx),
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
