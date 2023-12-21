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

	// TODO: Build custom output strings based on scenarios
	return fmt.Sprintf("%s [%s] %s\n", o.TimeStamp.Format(time.RFC3339), o.Category, o.Message)
}

type ServerOutputPipelineExecution struct {
	ServerOutput
	ExecutionID  string
	PipelineName string
	IsStart      bool
	Output       map[string]any
	Errors       []modconfig.StepError
}

func NewServerOutputPipelineExecution(serverOutput ServerOutput, execId string, name string, start bool) *ServerOutputPipelineExecution {
	return &ServerOutputPipelineExecution{
		ServerOutput: serverOutput,
		ExecutionID:  execId,
		PipelineName: name,
		IsStart:      start,
	}
}

func (o ServerOutputPipelineExecution) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var lines []string
	status := "started"
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	if !o.IsStart {
		if len(o.Errors) > 0 {
			status = fmt.Sprintf("failed with %d error(s)", len(o.Errors))
		} else {
			status = "finished"
		}
	}
	pre := fmt.Sprintf("%s [%s][%s[%s]]",
		o.TimeStamp.Format(time.RFC3339),
		o.Category,
		o.ExecutionID,
		o.PipelineName,
	)
	lines = append(lines, fmt.Sprintf("%s Pipeline %s", pre, status))

	if opts.Verbose {
		if len(o.Output) > 0 {
			outputs := sortAndParseMap(o.Output, "Output", " ", au, opts)
			lines = append(lines, fmt.Sprintf("%s Outputs\n%s", pre, outputs))
		}
	}

	if len(o.Errors) > 0 {
		for _, e := range o.Errors {
			lines = append(lines, fmt.Sprintf("%s error on step %s: %s", pre, e.Step, e.Error.Error()))
		}
	}
	return strings.Join(lines, "\n")
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
