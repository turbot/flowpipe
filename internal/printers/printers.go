package printers

import (
	"context"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
)

// Inspired by Kubernetes
//
// ResourcePrinter is an interface that knows how to print runtime objects.
type ResourcePrinter[T any] interface {
	// PrintResource receives a runtime object, formats it and prints it to a writer.
	PrintResource(context.Context, types.PrintableResource[T], io.Writer) error
}

func GetPrinter[T any](cmd *cobra.Command) ResourcePrinter[T] {
	format := cmd.Flags().Lookup(constants.ArgOutput).Value.String()

	// TODO: devise a more robust approach to determine a "list" command
	isListCmd := strings.Contains(cmd.Use, "list")

	switch format {
	case constants.OutputFormatPretty:
		if isListCmd {
			return TablePrinter[T]{}
		}
		return StringPrinter[T]{}
	case constants.OutputFormatPlain:
		if isListCmd {
			return TablePrinter[T]{}
		}
		return StringPrinter[T]{}
	case constants.OutputFormatJSON:
		return JsonPrinter[T]{}
	case constants.OutputFormatYAML:
		return YamlPrinter[T]{}
	}
	return nil
}
