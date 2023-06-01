package printers

import (
	"context"
	"fmt"
	"io"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/types"
)

// Inspired by Kubernetes
// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through
type TablePrinter struct {
	Delegate ResourcePrinter
}

func (p TablePrinter) PrintResource(ctx context.Context, items types.PrintableResource, writer io.Writer) error {

	table, err := items.GetTable()

	if err != nil {
		return err
	}
	err = p.Delegate.PrintResource(ctx, table, writer)
	return err

}

// Inspired by Kubernetes
type HumanReadableTablePrinter struct {
}

func (p HumanReadableTablePrinter) PrintResource(ctx context.Context, items types.PrintableResource, writer io.Writer) error {

	table, ok := items.(types.Table)

	if !ok {
		return fperr.BadRequestWithMessage("not a table")
	}

	for _, r := range table.Rows {
		for _, c := range r.Cells {
			fmt.Print(c)
			fmt.Print(" ")
		}
		fmt.Println()
	}
	return nil
}
