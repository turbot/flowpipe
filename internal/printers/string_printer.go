package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type StringPrinter[T any] struct {
	sanitizer sanitize.Sanitizer
}

func NewStringPrinter[T any](sanitizer *sanitize.Sanitizer) StringPrinter[T] {
	return StringPrinter[T]{
		sanitizer: *sanitizer,
	}
}

func (p StringPrinter[T]) PrintResource(_ context.Context, r types.PrintableResource[T], writer io.Writer) error {
	items := r.GetItems()
	for _, item := range items {
		if s, ok := any(item).(fmt.Stringer); ok {
			str := p.sanitizer.SanitizeString(s.String())
			_, err := writer.Write([]byte(str))
			if err != nil {
				return fmt.Errorf("error printing resource")
			}
		}
	}
	return nil
}
