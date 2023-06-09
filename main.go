package main

import (
	"context"

	"github.com/turbot/flowpipe/cache"
	"github.com/turbot/flowpipe/cmd"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/fplog"
)

func main() {

	// Create a single, global context for the application
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)
	ctx, err := config.ContextWithConfig(ctx)
	if err != nil {
		panic(err)
	}

	cache.InMemoryInitialize(nil)

	// Run the CLI
	err = cmd.RunCLI(ctx)
	if err != nil {
		panic(err)
	}
}
