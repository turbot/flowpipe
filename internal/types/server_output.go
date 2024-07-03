package types

import (
	"fmt"

	"github.com/logrusorgru/aurora"
	kitTypes "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"

	"github.com/turbot/go-kit/helpers"

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

func (o ServerOutputPrefix) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	timeStamp := au.BrightBlack(o.TimeStamp.Local().Format(time.DateTime))

	if !helpers.IsNil(o.execId) {
		c := opts.ColorGenerator.GetColorForElement(*o.execId)
		return au.Sprintf("%s %s ", timeStamp, au.Index(c, *o.execId))
	}

	if o.Category == "flowpipe" {
		return au.Sprintf("%s %s%s%s ", timeStamp, left, au.Cyan(o.Category), right)
	}

	return au.Sprintf("%s ", timeStamp)
}

type ServerOutputStatusChange struct {
	ServerOutputPrefix
	Status     string
	Content    string
	Additional int
}

func NewServerOutputStatusChange(ts time.Time, status string, content string) ServerOutputStatusChange {
	return ServerOutputStatusChange{
		ServerOutputPrefix: ServerOutputPrefix{
			TimeStamp: ts,
			Category:  "flowpipe",
		},
		Status:     status,
		Content:    content,
		Additional: -1,
	}
}

func NewServerOutputStatusChangeWithAdditional(ts time.Time, status string, content string, additional int) ServerOutputStatusChange {
	return ServerOutputStatusChange{
		ServerOutputPrefix: ServerOutputPrefix{
			TimeStamp: ts,
			Category:  "flowpipe",
		},
		Status:     status,
		Content:    content,
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

	switch strings.ToLower(o.Status) {
	case "started":
		return fmt.Sprintf("%s%s v%s\n", pre, au.Green(o.Status), o.Content)
	case "stopped":
		return fmt.Sprintf("%s%s\n", pre, au.Red(o.Status))
	case "listening":
		i := "all network interfaces"
		if o.Content != "" {
			i = o.Content
		}
		return fmt.Sprintf("%s%s on %s, port %d\n", pre, au.Yellow(o.Status), au.Yellow(i), au.Cyan(o.Additional))
	default:
		return fmt.Sprintf("%s%s %s\n", pre, o.Status, o.Content)
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
	text := "Loaded"
	if o.IsReload {
		text = "Reloaded"
	}

	return fmt.Sprintf("%s%s %s\n", pre, text, au.Green(o.ModName))
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

func NewServerOutputError(serverOutputPrefix ServerOutputPrefix, message string, err error) *ServerOutputError {
	return &ServerOutputError{
		ServerOutputPrefix: serverOutputPrefix,
		Message:            message,
		Error:              err,
	}
}

func (o ServerOutputError) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}

	var errorMessage string
	if o.Error != nil {
		errorMessage = o.Error.Error()
	} else {
		errorMessage = "unknown error"
	}

	return fmt.Sprintf("%s%s %s %s\n",
		o.ServerOutputPrefix.String(sanitizer, opts),
		au.Red("error"),
		au.Red(o.Message+":"),
		au.BrightRed(errorMessage))
}

type ServerOutputTriggerExecution struct {
	ServerOutputPrefix
	TriggerName  string
	PipelineName string
}

func NewServerOutputTriggerExecution(ts time.Time, execId string, name string, pipeline string) *ServerOutputTriggerExecution {
	prefix := NewServerOutputPrefixWithExecId(ts, "trigger", &execId)
	return &ServerOutputTriggerExecution{
		ServerOutputPrefix: prefix,
		TriggerName:        name,
		PipelineName:       pipeline,
	}
}

func (o ServerOutputTriggerExecution) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if o, err = sanitize.SanitizeStruct(sanitizer, o); err != nil {
		return ""
	}
	triggerSplit := strings.Split(o.TriggerName, ".")
	triggerType := triggerSplit[len(triggerSplit)-2]
	triggerName := triggerSplit[len(triggerSplit)-1]
	shortTrigger := fmt.Sprintf("trigger.%s.%s", triggerType, triggerName)
	triggerColor := opts.ColorGenerator.GetColorForElement(shortTrigger)

	shortPipeline := strings.Split(o.PipelineName, ".")[len(strings.Split(o.PipelineName, "."))-1]
	c := opts.ColorGenerator.GetColorForElement(shortPipeline)
	return fmt.Sprintf("%s%s%s%s fired, executing %s\n", o.ServerOutputPrefix.String(sanitize.NullSanitizer, opts), left, au.Index(triggerColor, shortTrigger), right, au.Index(c, shortPipeline))
}

type ServerOutputQueryTriggerRun struct {
	ServerOutputPrefix
	TriggerName string
	Inserted    int
	Updated     int
	Deleted     int
}

func NewServerOutputQueryTriggerRun(name string, inserted int, updated int, deleted int) ServerOutputQueryTriggerRun {
	prefix := NewServerOutputPrefixWithExecId(time.Now().Local(), "trigger", nil)
	return ServerOutputQueryTriggerRun{
		ServerOutputPrefix: prefix,
		TriggerName:        name,
		Inserted:           inserted,
		Updated:            updated,
		Deleted:            deleted,
	}
}

func (o ServerOutputQueryTriggerRun) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")

	triggerSplit := strings.Split(o.TriggerName, ".")
	triggerType := triggerSplit[len(triggerSplit)-2]
	triggerName := triggerSplit[len(triggerSplit)-1]
	shortTrigger := fmt.Sprintf("trigger.%s.%s", triggerType, triggerName)
	triggerColor := opts.ColorGenerator.GetColorForElement(shortTrigger)

	return fmt.Sprintf("%s%s%s%s trigger run, obtained %d inserted, %d updated, %d deleted row(s)\n",
		o.ServerOutputPrefix.String(sanitize.NullSanitizer, opts),
		left, au.Index(triggerColor, shortTrigger), right,
		au.Cyan(o.Inserted),
		au.Cyan(o.Updated),
		au.Cyan(o.Deleted),
	)
}

type ServerOutputTrigger struct {
	ServerOutputPrefix
	Name     string
	Type     string
	Enabled  *bool
	Schedule *string
	Method   *string
	Url      *string
	Sql      *string
}

func NewServerOutputTrigger(prefix ServerOutputPrefix, n string, t string, e *bool) *ServerOutputTrigger {
	return &ServerOutputTrigger{
		ServerOutputPrefix: prefix,
		Name:               n,
		Type:               t,
		Enabled:            e,
	}
}

func (o ServerOutputTrigger) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")

	// deliberately skip sanitizer as want to keep Url

	pre := o.ServerOutputPrefix.String(sanitizer, opts)
	shortName := strings.Split(o.Name, ".")[len(strings.Split(o.Name, "."))-1]
	shortTrigger := fmt.Sprintf("trigger.%s.%s", o.Type, shortName)
	triggerColor := opts.ColorGenerator.GetColorForElement(shortTrigger)

	if !helpers.IsNil(o.Enabled) && !*o.Enabled {
		return fmt.Sprintf("%s%s%s%s %s\n", pre, left, au.Index(triggerColor, shortTrigger), right, au.Red("Disabled"))
	}
	var suffix string
	switch o.Type {
	case "http":
		m := strings.ToUpper(kitTypes.SafeString(o.Method))
		u := kitTypes.SafeString(o.Url)

		suffix = fmt.Sprintf("HTTP %s %s", au.BrightBlack(m), au.Blue(u))
	case "schedule", "interval":
		s := kitTypes.SafeString(o.Schedule)
		suffix = fmt.Sprintf("Schedule: %s", au.Blue(s))
	case "query":
		s := kitTypes.SafeString(o.Schedule)
		q := kitTypes.SafeString(o.Sql)
		suffix = fmt.Sprintf("Schedule: %s - Query: %s", au.Blue(s), au.Blue(q))
	default:
		suffix = "loaded"
	}

	return fmt.Sprintf("%s%s%s%s %s %s\n", pre, left, au.Index(triggerColor, shortTrigger), right, au.Green("Enabled"), suffix)
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

func (p *PrintableServerOutput) GetTable() (*printers.Table, error) {
	return printers.NewTable(), nil
}
