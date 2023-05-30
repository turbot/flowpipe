package main

import (
	"context"

	"github.com/turbot/flowpipe/cmd"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/fplog"

	flowpipeapi "github.com/turbot/flowpipe-sdk-go"
)

func main() {

	flowpipeapi.Hello("you")

	// Create a single, global context for the application
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)
	ctx, err := config.ContextWithConfig(ctx)
	if err != nil {
		// TODO - don't panic
		panic(err)
	}

	// Run the CLI
	err = cmd.RunCLI(ctx)
	if err != nil {
		panic(err)
	}

}
