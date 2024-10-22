package cmd

import (
	"context"
	"log/slog"
	"os"

	"github.com/turbot/flowpipe/internal/fperr"
)

// RunCLI executes the root command.
func RunCLI(ctx context.Context) {
	cmd := rootCommand()

	if err := cmd.ExecuteContext(ctx); err != nil {
		slog.Debug("Error executing command", "error", err)

		exitCode := fperr.GetExitCode(err, false)
		os.Exit(exitCode)
	}
}
