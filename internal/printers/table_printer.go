package printers

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
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

	// Create a tabwriter
	w := tabwriter.NewWriter(writer, 1, 1, 4, ' ', tabwriter.TabIndent)

	// Print the table headers
	var tableHeaders string
	var tableFormatter string
	for i, c := range table.Columns {
		if i > 0 {
			tableHeaders += "\t"
			tableFormatter += "\t"
		}
		tableHeaders += c.Name
		tableFormatter += c.Formatter()
	}
	tableHeaders += "\n"
	tableFormatter += "\n"

	//nolint:forbidigo // this is how the tabwriter works
	_, err := fmt.Fprint(w, tableHeaders)
	if err != nil {
		return err
	}

	// Print each struct in the array as a row in the table
	for _, r := range table.Rows {
		//nolint:forbidigo // this is how the tabwriter works
		_, err := fmt.Fprintf(w, tableFormatter, r.Cells...)
		if err != nil {
			return err
		}
	}

	// Flush and display the table
	w.Flush()

	return nil
}
