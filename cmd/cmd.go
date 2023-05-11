package cmd

import (
	"context"
	"os"
)

// Execute executes the root command.
func RunCLI(ctx context.Context) error {
	cmd, err := RootCommand(ctx)
	if err != nil {
		return err
	}
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	return nil
}
