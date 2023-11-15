package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type StringPrinter struct {
}

func (p StringPrinter) PrintResource(_ context.Context, r types.PrintableResource, writer io.Writer) error {
	if items, ok := r.GetItems().([]any); ok {
		for _, item := range items {
			if s, ok := item.(fmt.Stringer); ok {
				_, err := writer.Write([]byte(s.String()))
				if err != nil {
					return fmt.Errorf("error printing resource")
				}
			}
		}
	}
	return nil
}
