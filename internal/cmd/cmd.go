package cmd

import (
	"context"
	"os"
)

// RunCLI executes the root command.
func RunCLI(ctx context.Context) {
	cmd := rootCommand()

	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
