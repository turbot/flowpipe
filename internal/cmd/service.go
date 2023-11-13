package cmd

import (
	"github.com/spf13/cobra"
)

func serviceCmd() *cobra.Command {

	serviceCmd := &cobra.Command{
		Use:   "service",
		Args:  cobra.NoArgs,
		Short: "Service commands",
	}

	serviceCmd.AddCommand(ServiceStartCmd())

	return serviceCmd
}
