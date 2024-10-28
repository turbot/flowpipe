package types

import (
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
)

type ProgressOutput struct {
	ExecutionId string                `json:"execution_id"`
	Output      map[string]any       `json:"output"`
	Errors      []flowpipe.StepError `json:"errors"`
}

func NewProgressOutput(executionId string, output map[string]any, errors []flowpipe.StepError) ProgressOutput {
	return ProgressOutput{
		ExecutionId: executionId,
		Output:      output,
		Errors:      errors,
	}
}

func (p ProgressOutput) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var out string
	if len(p.Output) > 0 {
		out += au.Bold("Outputs:").String() + "\n"
		out += sortAndParseMap(p.Output, "", "", au, opts)
	}
	if len(p.Errors) > 0 {
		if len(p.Output) > 0 {
			out += "\n"
		}

		for _, e := range p.Errors {
			out += fmt.Sprintf("%s%s\n%s\n\n", au.Red("Error: ").Bold().String(), au.Red(e.Error.Title).String(), au.Red(e.Error.Detail))
		}
	}
	return out
}

type PrintableProgressOutput struct {
	Items []sanitize.SanitizedStringer
}

func NewPrintableProgressOutput() *PrintableProgressOutput {
	return &PrintableProgressOutput{}
}

func (p *PrintableProgressOutput) GetItems() []sanitize.SanitizedStringer {
	return p.Items
}

func (p *PrintableProgressOutput) GetTable() (*printers.Table, error) {
	return printers.NewTable(), nil
}
