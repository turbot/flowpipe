package service

import (
	"context"

	"github.com/spf13/cobra"
)

func ServiceCmd(ctx context.Context) (*cobra.Command, error) {

	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Service commands",
	}

	serviceStartCmd, err := ServiceStartCmd(ctx)
	if err != nil {
		return nil, err
	}
	serviceCmd.AddCommand(serviceStartCmd)

	return serviceCmd, nil
}
