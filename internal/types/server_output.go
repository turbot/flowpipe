package types

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/pipe-fittings/modconfig"
	"strings"
	"time"
)

type ServerOutputPrefix struct {
	TimeStamp time.Time
	Category  string
}

func NewServerOutputPrefix(ts time.Time, category string) ServerOutputPrefix {
	return ServerOutputPrefix{
		TimeStamp: ts,
		Category:  category,
	}
}

func (o ServerOutputPrefix) String() string {
	return fmt.Sprintf("%s [%s]", o.TimeStamp.Format(time.RFC3339), o.Category)
}

type ServerOutput struct {
	ServerOutputPrefix
	Message string
}

func NewServerOutput(ts time.Time, category string, msg string) ServerOutput {
	return ServerOutput{
		ServerOutputPrefix: ServerOutputPrefix{
			TimeStamp: ts,
			Category:  category,
		},
		Message: msg,
	}
}

func (o ServerOutput) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	return fmt.Sprintf("%s %s\n", o.ServerOutputPrefix.String(), o.Message)
}

type ServerOutputPipelineExecution struct {
	ServerOutputPrefix
	ExecutionID  string
	PipelineName string
	Status       string
	Output       map[string]any
	Errors       []modconfig.StepError
}

func NewServerOutputPipelineExecution(prefix ServerOutputPrefix, execId string, name string, status string) *ServerOutputPipelineExecution {
	return &ServerOutputPipelineExecution{
		ServerOutputPrefix: prefix,
		ExecutionID:        execId,
		PipelineName:       name,
		Status:             status,
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
	pre := fmt.Sprintf("%s [%s%s]",
		o.ServerOutputPrefix.String(),
		o.ExecutionID,
		n,
	)
	lines = append(lines, fmt.Sprintf("%s Pipeline %s\n", pre, o.Status))

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

type ServerOutputTriggerExecution struct {
	ServerOutputPrefix
	ExecutionID  string
	TriggerName  string
	PipelineName string
}

func NewServerOutputTriggerExecution(prefix ServerOutputPrefix, execId string, name string, pipeline string) *ServerOutputTriggerExecution {
	return &ServerOutputTriggerExecution{
		ServerOutputPrefix: prefix,
		ExecutionID:        execId,
		TriggerName:        name,
		PipelineName:       pipeline,
	}
}

func (o ServerOutputTriggerExecution) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	return fmt.Sprintf("%s Trigger %s fired, executing Pipeline %s (%s)\n", o.ServerOutputPrefix.String(), o.TriggerName, o.PipelineName, o.ExecutionID)
}

type ServerOutputStepExecution struct {
	ServerOutputPrefix
	ExecutionID  string
	PipelineName string
	StepName     string
	StepType     string
	Status       string
	Output       map[string]any
	Errors       []modconfig.StepError
}

func NewServerOutputStepExecution(prefix ServerOutputPrefix, execId string, pipelineName string, stepName string, stepType string, status string) *ServerOutputStepExecution {
	return &ServerOutputStepExecution{
		ServerOutputPrefix: prefix,
		ExecutionID:        execId,
		PipelineName:       pipelineName,
		StepName:           stepName,
		StepType:           stepType,
		Status:             status,
	}
}

func (o ServerOutputStepExecution) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var lines []string
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	// put Steps behind verbose flag
	if !opts.Verbose {
		return ""
	}

	p := o.PipelineName
	if p != "" {
		p = fmt.Sprintf("[%s]", p)
	}

	pre := fmt.Sprintf("%s [%s%s]", o.ServerOutputPrefix.String(), o.ExecutionID, p)

	lines = append(lines, fmt.Sprintf("%s %s step %s %s\n", pre, o.StepType, o.StepName, o.Status))

	if len(o.Output) > 0 {
		outputs := sortAndParseMap(o.Output, "Output", " ", au, opts)
		lines = append(lines, fmt.Sprintf("%s Outputs\n%s\n", pre, outputs))
	}

	if len(o.Errors) > 0 {
		for _, e := range o.Errors {
			lines = append(lines, fmt.Sprintf("%s error on %s step %s: %s\n", pre, o.StepType, o.StepName, e.Error.Error()))
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
