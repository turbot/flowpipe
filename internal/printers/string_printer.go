package printers

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/color"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
	"io"
)

type StringPrinter struct {
	colorGenerator *color.DynamicColorGenerator
}

func NewStringPrinter() (*StringPrinter, error) {
	colorGenerator, err := color.NewDynamicColorGenerator(0, 16)
	if err != nil {
		return nil, err
	}

	return &StringPrinter{
		colorGenerator: colorGenerator,
	}, nil
}

func (p StringPrinter) PrintResource(_ context.Context, r types.PrintableResource[types.SanitizedStringer], writer io.Writer) error {
	items := r.GetItems()
	for _, item := range items {
		str := item.String(sanitize.Instance, p.colorGenerator)

		_, err := writer.Write([]byte(str))
		if err != nil {
			return fmt.Errorf("error printing resource")
		}
	}
	return nil
}
