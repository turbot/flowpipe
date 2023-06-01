package printers

import (
	"context"
	"io"

	"github.com/turbot/flowpipe/types"
)

// Inspired by Kubernetes
//
// HumanReadablePrinter is an implementation of ResourcePrinter which attempts to provide
// more elegant output. It is not threadsafe, but you may call PrintObj repeatedly; headers
// will only be printed if the object type changes. This makes it useful for printing items
// received from watches.
type HumanReadablePrinter struct {
}

func (h *HumanReadablePrinter) PrintObj(context.Context, types.FlowpipeResources, io.Writer) error {
	return nil
}
