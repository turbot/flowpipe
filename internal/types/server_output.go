package types

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/pipe-fittings/modconfig"
	"strings"
	"time"
)

type ServerOutput struct {
	TimeStamp time.Time
	Category  string
	Message   string
}

func NewServerOutput(ts time.Time, category string, msg string) ServerOutput {
	return ServerOutput{
		TimeStamp: ts,
		Category:  category,
		Message:   msg,
	}
}

func (o ServerOutput) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	return fmt.Sprintf("%s [%s] %s\n", o.TimeStamp.Format(time.RFC3339), o.Category, o.Message)
}

type ServerOutputPipelineExecution struct {
	ServerOutput
	ExecutionID  string
	PipelineName string
	Output       map[string]any
	Errors       []modconfig.StepError
}

func NewServerOutputPipelineExecution(serverOutput ServerOutput, execId string, name string) *ServerOutputPipelineExecution {
	return &ServerOutputPipelineExecution{
		ServerOutput: serverOutput,
		ExecutionID:  execId,
		PipelineName: name,
	}
}

func (o ServerOutputPipelineExecution) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var lines []string
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}
	n := o.PipelineName
	if n != "" {
		n = fmt.Sprintf("[%s]", n)
	}
	pre := fmt.Sprintf("%s [%s] [%s%s]",
		o.TimeStamp.Format(time.RFC3339),
		o.Category,
		o.ExecutionID,
		n,
	)
	lines = append(lines, fmt.Sprintf("%s Pipeline %s\n", pre, o.Message))

	if opts.Verbose {
		if len(o.Output) > 0 {
			outputs := sortAndParseMap(o.Output, "Output", " ", au, opts)
			lines = append(lines, fmt.Sprintf("%s Outputs\n%s\n", pre, outputs))
		}
	}

	if len(o.Errors) > 0 {
		for _, e := range o.Errors {
			lines = append(lines, fmt.Sprintf("%s error on step %s: %s\n", pre, e.Step, e.Error.Error()))
		}
	}
	return strings.Join(lines, "")
}

type PrintableServerOutput struct {
	Items []SanitizedStringer
}

func NewPrintableServerOutput() *PrintableServerOutput {
	return &PrintableServerOutput{}
}

func (p *PrintableServerOutput) GetItems() []SanitizedStringer {
	return p.Items
}

func (p *PrintableServerOutput) GetTable() (Table, error) {
	return Table{}, nil
}
