package cmd

import (
	"github.com/spf13/cobra"
)

func serviceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Args:  cobra.NoArgs,
		Short: "Service commands",
	}

	cmd.AddCommand(serviceStartCmd())

	return cmd
}
