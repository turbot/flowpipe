package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/sanitize"
	"io"
	"text/tabwriter"

	"github.com/turbot/flowpipe/internal/types"
)

// Inspired by Kubernetes
// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through
type TablePrinter[T any] struct {
	Delegate ResourcePrinter[types.TableRow]
}

func (p TablePrinter[T]) PrintResource(ctx context.Context, items types.PrintableResource[T], writer io.Writer, sanitizer *sanitize.Sanitizer) error {
	table, err := items.GetTable()

	if err != nil {
		return err
	}
	err = p.PrintTable(ctx, table, writer, sanitizer)
	return err
}

func (p TablePrinter[T]) PrintTable(ctx context.Context, table types.Table, writer io.Writer, sanitizer *sanitize.Sanitizer) error {
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
