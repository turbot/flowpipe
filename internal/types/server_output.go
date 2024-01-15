package types

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/go-kit/helpers"
	kitTypes "github.com/turbot/go-kit/types"
	"strings"
	"time"
)

type ServerOutputPrefix struct {
	TimeStamp time.Time
	Category  string
	execId    *string
}

func NewServerOutputPrefix(ts time.Time, category string) ServerOutputPrefix {
	return ServerOutputPrefix{
		TimeStamp: ts,
		Category:  category,
	}
}

func NewServerOutputPrefixWithExecId(ts time.Time, category string, execId *string) ServerOutputPrefix {
	return ServerOutputPrefix{
		TimeStamp: ts,
		Category:  category,
		execId:    execId,
	}
}

func (o ServerOutputPrefix) String(_ *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	dot := au.BrightBlack(".")
	var cat string
	switch o.Category {
	case "flowpipe":
		cat = au.Cyan(o.Category).String()
	case "mod":
		cat = au.Green(o.Category).String()
	case "pipeline":
		if !helpers.IsNil(o.execId) {
			c := opts.ColorGenerator.GetColorForElement(*o.execId)
			cat = aurora.Sprintf("%s%s%s", au.Magenta(o.Category), dot, au.Index(c, *o.execId))
		} else {
			cat = au.Magenta(o.Category).String()
		}
	case "trigger":
		cat = au.Yellow(o.Category).String()
	default:
		cat = au.Blue(o.Category).String()
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

func (o ServerOutputStatusChange) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
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

func (o ServerOutputLoaded) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
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

func (o ServerOutput) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
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

func (o ServerOutputError) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
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

	return fmt.Sprintf("%strigger %s fired, executing Pipeline %s (%s)\n", o.ServerOutputPrefix.String(sanitizer, opts), o.TriggerName, o.PipelineName, o.ExecutionID)
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

func (o ServerOutputTrigger) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
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
