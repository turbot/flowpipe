package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type StringPrinter struct {
	sanitizer *sanitize.Sanitizer
}

func NewStringPrinter(sanitizer *sanitize.Sanitizer) StringPrinter {
	return StringPrinter{
		sanitizer: sanitizer,
	}
}

func (p StringPrinter) PrintResource(_ context.Context, r types.PrintableResource[fmt.Stringer], writer io.Writer) error {
	items := r.GetItems()
	for _, item := range items {
		_, err := writer.Write([]byte(item.String()))
		if err != nil {
			return fmt.Errorf("error printing resource")
		}
	}
	return nil
}
