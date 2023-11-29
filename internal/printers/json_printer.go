package printers

import (
	"context"
	"github.com/turbot/flowpipe/internal/sanitize"
	"io"

	"github.com/hokaccha/go-prettyjson"
	"github.com/turbot/flowpipe/internal/types"
)

type JsonPrinter[T any] struct {
}

func (p JsonPrinter[T]) PrintResource(ctx context.Context, r types.PrintableResource[T], writer io.Writer, sanitizer *sanitize.Sanitizer) error {
	s, err := prettyjson.Marshal(r.GetItems())
	if err != nil {
		return err
	}
	_, err = writer.Write(s)
	if err != nil {
		return err
	}

	return nil
}
