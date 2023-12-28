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

type StringPrinter[T any] struct {
	colorGenerator *color.DynamicColorGenerator
	Sanitizer      *sanitize.Sanitizer
}

func NewStringPrinter[T any]() (*StringPrinter[T], error) {
	colorGenerator, err := color.NewDynamicColorGenerator(0, 16)
	if err != nil {
		return nil, err
	}

	p := &StringPrinter[T]{
		colorGenerator: colorGenerator,
		Sanitizer:      sanitize.NullSanitizer,
	}
	return p, nil
}

func (p StringPrinter[T]) PrintResource(_ context.Context, r types.PrintableResource[T], writer io.Writer) error {
	items := r.GetItems()
	enableColor := viper.GetString(constants.ArgOutput) == constants.OutputFormatPretty
	for _, item := range items {
		if item, isSanitizedStringer := any(item).(types.SanitizedStringer); isSanitizedStringer {
			colorOpts := types.RenderOptions{
				ColorGenerator: p.colorGenerator,
				ColorEnabled:   enableColor,
				Verbose:        viper.GetBool(constants.ArgVerbose),
				JsonFormatter:  color.NewJsonFormatter(!enableColor),
			}

			str := item.String(sanitize.Instance, colorOpts)

			if _, err := writer.Write([]byte(str)); err != nil {
				return fmt.Errorf("error printing resource")
			}
		}
	}
	return nil
}
