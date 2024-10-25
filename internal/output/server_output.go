package output

import (
	"context"
	"os"

	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"

	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
)

var IsServerMode bool
var serverOutputPrinter *printers.StringPrinter[sanitize.SanitizedStringer]

func RenderServerOutput(ctx context.Context, outputs ...sanitize.SanitizedStringer) {
	if !IsServerMode {
		return
	}

	// TODO: determine if we should set this up once when server command is started...
	if serverOutputPrinter == nil {
		printer, err := printers.NewStringPrinter[sanitize.SanitizedStringer]()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}
		printer.Sanitizer = sanitize.ServerInstance
		serverOutputPrinter = printer
	}
	printableResource := types.NewPrintableServerOutput()
	printableResource.Items = outputs

	_ = serverOutputPrinter.PrintResource(ctx, printableResource, os.Stdout)
}
