package printers

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
)

// Inspired by Kubernetes
// TablePrinter decodes table objects into typed objects before delegating to another printer.
// Non-table types are simply passed through
type TablePrinter[T any] struct {
	Sanitizer sanitize.Sanitizer
}

func NewTablePrinter[T any]() (*TablePrinter[T], error) {
	return &TablePrinter[T]{
		Sanitizer: *sanitize.NullSanitizer,
	}, nil
}

func (p TablePrinter[T]) PrintResource(_ context.Context, items types.PrintableResource[T], writer io.Writer) error {
	table, err := items.GetTable()

	if err != nil {
		return err
	}
	err = p.PrintTable(table, writer)
	return err
}

func (p TablePrinter[T]) PrintTable(table types.Table, writer io.Writer) error {
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
		// format the row
		str := fmt.Sprintf(tableFormatter, r.Cells...)
		// sanitize
		str = p.Sanitizer.SanitizeString(str)

		// write
		//nolint:forbidigo // this is how the tabwriter works
		_, err := fmt.Fprint(w, str)
		if err != nil {
			return err
		}
	}

	// Flush and display the table
	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}
