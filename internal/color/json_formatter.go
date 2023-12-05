package color

import (
	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
)

func NewJsonFormatter() *prettyjson.Formatter {
	return &prettyjson.Formatter{
		KeyColor:        color.New(color.FgBlue),
		StringColor:     color.New(color.FgGreen),
		BoolColor:       color.New(color.FgYellow),
		NumberColor:     color.New(color.FgCyan),
		NullColor:       color.New(color.FgBlack),
		StringMaxLength: 0,
		DisabledColor:   false,
		Indent:          2,
		Newline:         "\n",
	}
}
