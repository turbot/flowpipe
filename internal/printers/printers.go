package printers

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
	"io"
	"slices"
)

// Inspired by Kubernetes
//
// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter[T any] interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, types.PrintableResource[T], io.Writer) error
}

func GetPrinter[T any](cmd *cobra.Command) (ResourcePrinter[T], error) {
	format := cmd.Flags().Lookup(constants.ArgOutput).Value.String()
	key := cmdconfig.CommandFullKey(cmd)
	useTable := []string{
		"flowpipe.trigger.list",
		"flowpipe.pipeline.list",
		"flowpipe.variable.list",
	}

	switch format {
	case constants.OutputFormatPretty:
		if slices.Contains(useTable, key) {
			return NewTablePrinter[T]()
		}
		return NewStringPrinter[T]()
	case constants.OutputFormatPlain:
		if slices.Contains(useTable, key) {
			return NewTablePrinter[T]()
		}
		return NewStringPrinter[T]()
	case constants.OutputFormatJSON:
		return NewJsonPrinter[T]()
	case constants.OutputFormatYAML:
		return NewYamlPrinter[T]()
	}
	return nil, nil
}
