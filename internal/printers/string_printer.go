package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type StringPrinter struct {
}

func (p StringPrinter) PrintResource(_ context.Context, r types.PrintableResource, writer io.Writer, sanitizer *sanitize.Sanitizer) error {
	// TODO kai does this cast work for things which return []Trigger for example
	if items, ok := r.GetItems(sanitizer).([]any); ok {
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
