package printers

import (
	"context"
	"github.com/turbot/flowpipe/internal/sanitize"
	"io"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
)

// Inspired by Kubernetes
//
// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter[T any] interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, types.PrintableResource[T], io.Writer, *sanitize.Sanitizer) error
}

func GetPrinter[T any](cmd *cobra.Command) ResourcePrinter[T] {

	format := cmd.Flags().Lookup(constants.ArgOutput).Value.String()

	switch format {
	case "table":
		return TablePrinter[T]{}
	case "json":
		return JsonPrinter[T]{}
	case "yaml":
		return YamlPrinter[T]{}
	}
	return nil
}
