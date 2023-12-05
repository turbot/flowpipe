package printers

import (
	"context"
	"encoding/json"
	"fmt"

	"io"

	"github.com/fatih/color"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/printer"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/flowpipe/internal/types"
)

// Inspired by https://github.com/goccy/go-yaml/blob/master/cmd/ycat/ycat.go
type YamlPrinter[T any] struct {
	Sanitizer sanitize.Sanitizer
}

func NewYamlPrinter[T any]() (*YamlPrinter[T], error) {
	return &YamlPrinter[T]{
		Sanitizer: *sanitize.NullSanitizer,
	}, nil
}

const escape = "\x1b"

func format(attr color.Attribute) string {
	return fmt.Sprintf("%s[%dm", escape, attr)
}

func (px YamlPrinter[T]) PrintResource(ctx context.Context, r types.PrintableResource[T], writer io.Writer) error {
	// marshal to json to avoid having to put yaml tags on all structs
	s, err := json.Marshal(r.GetItems())
	if err != nil {
		return err
	}

	// sanitize
	s = []byte(px.Sanitizer.SanitizeString(string(s)))

	// convert to yaml
	yamlBytes, err := yaml.JSONToYAML(s)
	if err != nil {
		return err
	}

	// this is a copy of https://github.com/goccy/go-yaml/blob/master/cmd/ycat/ycat.go
	tokens := lexer.Tokenize(string(yamlBytes))
	var p printer.Printer
	p.LineNumber = false
	// p.LineNumberFormat = func(num int) string {
	// 	fn := color.New(color.Bold, color.FgHiWhite).SprintFunc()
	// 	return fn(fmt.Sprintf("%2d | ", num))
	// }
	p.Bool = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiMagenta),
			Suffix: format(color.Reset),
		}
	}
	p.Number = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiMagenta),
			Suffix: format(color.Reset),
		}
	}
	p.MapKey = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiCyan),
			Suffix: format(color.Reset),
		}
	}
	p.Anchor = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiYellow),
			Suffix: format(color.Reset),
		}
	}
	p.Alias = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiYellow),
			Suffix: format(color.Reset),
		}
	}
	p.String = func() *printer.Property {
		return &printer.Property{
			Prefix: format(color.FgHiGreen),
			Suffix: format(color.Reset),
		}
	}
	// stdOut := colorable.NewColorableStdout()
	_, err = writer.Write([]byte(p.PrintTokens(tokens) + "\n"))
	if err != nil {
		return err
	}
	return nil
}
