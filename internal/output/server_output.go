package output

import (
	"context"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
	"os"
)

var IsServerMode bool

func RenderServerOutput(ctx context.Context, outputs ...types.SanitizedStringer) {
	if !IsServerMode {
		return
	}
	printer, err := printers.NewStringPrinter[types.SanitizedStringer]()
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}
	printer.Sanitizer = sanitize.Instance
	printableResource := types.NewPrintableServerOutput()
	printableResource.Items = outputs

	err = printer.PrintResource(ctx, printableResource, os.Stdout)
}
