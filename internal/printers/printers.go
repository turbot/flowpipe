package printers

import (
	"context"
	"io"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/types"
)

// Inspired by Kubernetes
//
// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, types.PrintableResource, io.Writer) error
}

func GetPrinter(cmd *cobra.Command) ResourcePrinter {

	format := cmd.Flags().Lookup(constants.ArgOutput).Value.String()

	switch format {
	case "table":
		return TablePrinter{
			Delegate: HumanReadableTablePrinter{},
		}
	case "json":
		return JsonPrinter{}
	case "yaml":
		return YamlPrinter{}
	}
	return nil
}
