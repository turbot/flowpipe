package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type StringPrinter struct {
}

func (p StringPrinter) PrintResource(_ context.Context, r types.PrintableResource[fmt.Stringer], writer io.Writer) error {
	// TODO KAI how do we sanitize
	items := r.GetItems()
	for _, item := range items {
		_, err := writer.Write([]byte(item.String()))
		if err != nil {
			return fmt.Errorf("error printing resource")
		}
	}
	return nil
}
