package cmd

import (
	"context"
	"os"
)

// RunCLI executes the root command.
func RunCLI(ctx context.Context) error {
	cmd := rootCommand(ctx)

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
	return nil
}
