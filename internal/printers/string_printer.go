package printers

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/color"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
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
		colorOpts := types.ColorOptions{
			ColorGenerator: p.colorGenerator,
			ColourEnabled:  viper.GetString(constants.ArgOutput) == constants.OutputFormatPretty,
		}
		str := item.String(sanitize.Instance, colorOpts)

		if _, err := writer.Write([]byte(str)); err != nil {
			return fmt.Errorf("error printing resource")
		}
	}
	return nil
}
