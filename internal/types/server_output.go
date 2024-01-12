package types

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	kitTypes "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
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

func (o ServerOutputPrefix) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	var cat aurora.Value
	switch o.Category {
	case "flowpipe":
		cat = au.Cyan(o.Category)
	case "mod":
		cat = au.Green(o.Category)
	case "pipeline":
		cat = au.Magenta(o.Category)
	case "trigger":
		cat = au.Yellow(o.Category)
	default:
		cat = au.Blue(o.Category)
	}
	return aurora.Sprintf("%s %s%s%s ", au.BrightBlack(o.TimeStamp.Local().Format(time.DateTime)), left, cat, right)
}

type ServerOutputStatusChange struct {
	ServerOutputPrefix
	Status     string
	Additional string
}

func NewServerOutputStatusChange(ts time.Time, status string, additional string) ServerOutputStatusChange {
	return ServerOutputStatusChange{
		ServerOutputPrefix: ServerOutputPrefix{
			TimeStamp: ts,
			Category:  "flowpipe",
		},
		Status:     status,
		Additional: additional,
	}
}

func (o ServerOutputStatusChange) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	pre := o.ServerOutputPrefix.String(sanitizer, opts)

	switch o.Status {
	case "started":
		return fmt.Sprintf("%sserver %s\n", pre, au.Green(o.Status))
	case "stopped":
		return fmt.Sprintf("%sserver %s\n", pre, au.Red(o.Status))
	case "listening":
		return fmt.Sprintf("%sserver %s on %s\n", pre, au.Yellow(o.Status), au.Yellow(o.Additional))
	default:
		return fmt.Sprintf("%sserver %s\n", pre, o.Status)
	}
}

type ServerOutputLoaded struct {
	ServerOutputPrefix
	ModName  string
	IsReload bool
}

func NewServerOutputLoaded(serverOutputPrefix ServerOutputPrefix, modName string, isReload bool) *ServerOutputLoaded {
	return &ServerOutputLoaded{
		ServerOutputPrefix: serverOutputPrefix,
		ModName:            modName,
		IsReload:           isReload,
	}
}

func (o ServerOutputLoaded) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	pre := o.ServerOutputPrefix.String(nil, opts)
	text := "loaded"
	if o.IsReload {
		text = "reloaded"
	}

	return fmt.Sprintf("%s%s mod %s\n", pre, text, au.Green(o.ModName))
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

func (o ServerOutput) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	return fmt.Sprintf("%s%s\n", o.ServerOutputPrefix.String(sanitizer, opts), o.Message)
}

type ServerOutputError struct {
	ServerOutputPrefix
	Message string
	Error   error
}

func NewServerOutputError(serverOutputPrefix ServerOutputPrefix, message string, error error) *ServerOutputError {
	return &ServerOutputError{
		ServerOutputPrefix: serverOutputPrefix,
		Message:            message,
		Error:              error,
	}
}

func (o ServerOutputError) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	suffix := ""
	if opts.Verbose {
		suffix = fmt.Sprintf("\n%s", au.BrightRed(o.Error.Error()))
	}

	return fmt.Sprintf("%s%s %s%s\n",
		o.ServerOutputPrefix.String(sanitizer, opts),
		au.Red("error"),
		au.Red(o.Message),
		suffix)
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

func (o ServerOutputPipelineExecution) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var lines []string
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	c := opts.ColorGenerator.GetColorForElement(o.ExecutionID)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	dot := au.BrightBlack(".")

	pre := fmt.Sprintf("%s%s%s%s%s%s",
		o.ServerOutputPrefix.String(sanitizer, opts),
		left,
		au.Sprintf(au.Index(c, o.PipelineName)),
		dot,
		au.Sprintf(au.Index(c, o.ExecutionID)),
		right,
	)
	var status string
	switch o.Status {
	case "started":
		status = au.Cyan(o.Status).String()
	case "finished":
		status = au.Green(o.Status).String()
	case "failed":
		status = au.Red(o.Status).String()
	case "canceled":
		status = au.Blue(o.Status).String()
	case "queued":
		status = au.Yellow(o.Status).String()
	default:
		status = o.Status
	}
	lines = append(lines, fmt.Sprintf("%s pipeline %s\n", pre, status))

	if opts.Verbose {
		if len(o.Output) > 0 {
			outputs := sortAndParseMap(o.Output, "output:", " ", au, opts)
			lines = append(lines, fmt.Sprintf("%s pipeline outputs\n%s\n", pre, outputs))
		}
	}

	if len(o.Errors) > 0 {
		for _, e := range o.Errors {
			errLine := fmt.Sprintf("error on step %s: %s\n", e.Step, e.Error.Error())
			lines = append(lines, fmt.Sprintf("%s %s", pre, au.Red(errLine)))
		}
	}

	out := strings.Join(lines, "")
	// trim double new line ending
	if strings.HasSuffix(out, "\n\n") {
		out = strings.TrimSuffix(out, "\n")
	}
	return out
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

func (o ServerOutputTriggerExecution) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	return fmt.Sprintf("%strigger %s fired, executing Pipeline %s (%s)\n", o.ServerOutputPrefix.String(sanitizer, opts), o.TriggerName, o.PipelineName, o.ExecutionID)
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

func (o ServerOutputStepExecution) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
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

	c := opts.ColorGenerator.GetColorForElement(o.ExecutionID)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	dot := au.BrightBlack(".")

	pre := fmt.Sprintf("%s%s%s%s%s%s",
		o.ServerOutputPrefix.String(sanitizer, opts),
		left,
		au.Sprintf(au.Index(c, o.PipelineName)),
		dot,
		au.Sprintf(au.Index(c, o.ExecutionID)),
		right,
	)

	var status string
	switch o.Status {
	case "started":
		status = au.Cyan(o.Status).String()
	case "finished":
		status = au.Green(o.Status).String()
	case "failed":
		status = au.Red(o.Status).String()
	case "retrying":
		status = au.Yellow(o.Status).String()
	default:
		status = o.Status
	}

	lines = append(lines, fmt.Sprintf("%s %s step %s %s\n", pre, au.Blue(o.StepType), au.BrightBlue(o.StepName), status))

	if len(o.Output) > 0 {
		outputs := sortAndParseMap(o.Output, "output", " ", au, opts)
		lines = append(lines, fmt.Sprintf("%s step outputs\n%s\n", pre, outputs))
	}

	if len(o.Errors) > 0 {
		for _, e := range o.Errors {
			errLine := fmt.Sprintf("error on %s step %s: %s\n", o.StepType, o.StepName, e.Error.Error())
			lines = append(lines, fmt.Sprintf("%s %s", pre, au.Red(errLine)))
		}
	}

	out := strings.Join(lines, "")
	// trim double new line ending
	if strings.HasSuffix(out, "\n\n") {
		out = strings.TrimSuffix(out, "\n")
	}
	return out
}

type ServerOutputTrigger struct {
	ServerOutputPrefix
	Name     string
	Type     string
	Schedule *string
	Method   *string
	Url      *string
	Sql      *string
}

func NewServerOutputTrigger(prefix ServerOutputPrefix, n string, t string) *ServerOutputTrigger {
	return &ServerOutputTrigger{
		ServerOutputPrefix: prefix,
		Name:               n,
		Type:               t,
	}
}

func (o ServerOutputTrigger) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)

	// deliberately skip sanitizer as want to keep Url

	pre := o.ServerOutputPrefix.String(sanitizer, opts)
	var suffix string
	switch o.Type {
	case "http":
		m := strings.ToUpper(kitTypes.SafeString(o.Method))
		u := kitTypes.SafeString(o.Url)

		suffix = fmt.Sprintf("%s %s", au.BrightBlack(m), au.Yellow(u))
	case "schedule", "interval":
		s := kitTypes.SafeString(o.Schedule)
		suffix = fmt.Sprintf("%s", au.Yellow(s))
	case "query":
		s := kitTypes.SafeString(o.Schedule)
		q := kitTypes.SafeString(o.Sql)
		suffix = fmt.Sprintf("schedule %s - query %s", au.Yellow(s), au.Yellow(q))
	default:
		suffix = "loaded"
	}

	return fmt.Sprintf("%s%s %s - %s\n", pre, au.BrightBlue(o.Name), o.Type, suffix)
}

type PrintableServerOutput struct {
	Items []sanitize.SanitizedStringer
}

func NewPrintableServerOutput() *PrintableServerOutput {
	return &PrintableServerOutput{}
}

func (p *PrintableServerOutput) GetItems() []sanitize.SanitizedStringer {
	return p.Items
}

func (p *PrintableServerOutput) GetTable() (printers.Table, error) {
	return printers.Table{}, nil
}
