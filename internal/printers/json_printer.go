package printers

import (
	"context"
	"github.com/turbot/flowpipe/internal/sanitize"
	"io"

	"github.com/hokaccha/go-prettyjson"
	"github.com/turbot/flowpipe/internal/types"
)

type JsonPrinter struct {
}

func (p JsonPrinter) PrintResource(ctx context.Context, r types.PrintableResource, writer io.Writer, sanitizer *sanitize.Sanitizer) error {
	s, err := prettyjson.Marshal(r.GetItems(sanitizer))
	if err != nil {
		return err
	}
	_, err = writer.Write(s)
	if err != nil {
		return err
	}

	return nil
}
