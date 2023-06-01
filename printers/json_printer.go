package printers

import (
	"context"
	"fmt"
	"io"

	"github.com/turbot/flowpipe/types"

	"github.com/hokaccha/go-prettyjson"
)

type JsonPrinter struct {
}

func (h *JsonPrinter) PrintObj(ctx context.Context, resource types.FlowpipeResources, writer io.Writer) error {
	s, _ := prettyjson.Marshal(resource)
	fmt.Println(string(s))

	return nil
}
