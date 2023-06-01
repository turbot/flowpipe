package printers

import (
	"context"
	"io"

	"github.com/turbot/flowpipe/types"
)

// Inspired by Kubernetes
//
// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter interface {
	// PrintObj receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, types.FlowpipeResources, io.Writer) error
}
