package printers

import (
	"context"
	"io"

	"github.com/hokaccha/go-prettyjson"
	"github.com/turbot/flowpipe/internal/types"
)

type JsonPrinter struct {
}

func (p JsonPrinter) PrintResource(ctx context.Context, r types.PrintableResource, writer io.Writer) error {
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
