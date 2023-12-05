package printers

import (
	"context"
	"encoding/json"
	"github.com/turbot/flowpipe/internal/sanitize"
	"io"

	"github.com/hokaccha/go-prettyjson"
	"github.com/turbot/flowpipe/internal/types"
)

type JsonPrinter[T any] struct {
	Sanitizer *sanitize.Sanitizer
}

func NewJsonPrinter[T any]() (*JsonPrinter[T], error) {
	return &JsonPrinter[T]{
		Sanitizer: sanitize.NullSanitizer,
	}, nil
}

func (p JsonPrinter[T]) PrintResource(ctx context.Context, r types.PrintableResource[T], writer io.Writer) error {
	// marshal
	s, err := json.Marshal(r.GetItems())
	if err != nil {
		return err
	}

	// sanitize
	s = []byte(p.Sanitizer.SanitizeString(string(s)))

	// format
	s, err = prettyjson.Format(s)
	if err != nil {
		return err
	}
	_, err = writer.Write(s)
	if err != nil {
		return err
	}

	return nil
}
