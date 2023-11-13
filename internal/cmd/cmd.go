package cmd

import (
	"context"
	"os"
)

// RunCLI executes the root command.
func RunCLI(ctx context.Context) error {
	cmd, err := RootCommand(ctx)
	if err != nil {
		return err
	}
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
	return nil
}
