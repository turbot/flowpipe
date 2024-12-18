package types

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/color"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type ParsedHeader struct {
	ExecutionId string `json:"execution_id"`
	IsStale     bool   `json:"is_stale"`
	LastLoaded  string `json:"last_loaded"`
}

func (p ParsedHeader) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	out := fmt.Sprintf("%s%s%s %s %s\n", left, au.Cyan("flowpipe"), right, "Execution ID:", p.ExecutionId)
	if p.IsStale {
		out += fmt.Sprintf("%s%s%s %s\n", left, au.Cyan("flowpipe"), right, au.Sprintf(au.Yellow("Warning: Mod is stale, last loaded %s"), p.LastLoaded))
	}
	return out
}

type ParsedEventPrefix struct {
	FullPipelineName string  `json:"full_pipeline_name"`
	PipelineName     string  `json:"pipeline_name"`
	FullStepName     *string `json:"full_step_name,omitempty"`
	StepName         *string `json:"step_name,omitempty"`
	ForEachKey       *string `json:"for_each_key,omitempty"`
	LoopIndex        *int    `json:"loop_index,omitempty"`
	RetryIndex       *int    `json:"retry_index,omitempty"`

	prefix *ServerOutputPrefix
}

func NewPrefix(fullPipelineName string) ParsedEventPrefix {
	return ParsedEventPrefix{
		FullPipelineName: fullPipelineName,
		PipelineName:     strings.Split(fullPipelineName, ".")[len(strings.Split(fullPipelineName, "."))-1],
	}
}

func NewPrefixWithServer(fullPipelineName string, serverPrefix ServerOutputPrefix) ParsedEventPrefix {
	return ParsedEventPrefix{
		FullPipelineName: fullPipelineName,
		PipelineName:     strings.Split(fullPipelineName, ".")[len(strings.Split(fullPipelineName, "."))-1],
		prefix:           &serverPrefix,
	}
}

func NewParsedEventPrefix(fullPipelineName string, fullStepName *string, forEachKey *string, loopIndex *int, retryIndex *int, serverPrefix *ServerOutputPrefix) ParsedEventPrefix {
	prefix := ParsedEventPrefix{
		FullPipelineName: fullPipelineName,
		PipelineName:     strings.Split(fullPipelineName, ".")[len(strings.Split(fullPipelineName, "."))-1],
	}

	if !helpers.IsNil(fullStepName) {
		prefix.FullStepName = fullStepName
		prefix.StepName = &strings.Split(*fullStepName, ".")[len(strings.Split(*fullStepName, "."))-1]
	}

	if !helpers.IsNil(forEachKey) {
		prefix.ForEachKey = forEachKey
	}

	if !helpers.IsNil(loopIndex) {
		prefix.LoopIndex = loopIndex
	}

	if !helpers.IsNil(retryIndex) {
		prefix.RetryIndex = retryIndex
	}

	if !helpers.IsNil(serverPrefix) {
		prefix.prefix = serverPrefix
	}

	return prefix
}

func (p ParsedEventPrefix) getRetryString(au aurora.Aurora) string {
	if p.RetryIndex == nil || *p.RetryIndex <= 1 {
		return ""
	}
	return au.Sprintf(au.Index(8, "#%d"), *p.RetryIndex)
}

func (p ParsedEventPrefix) getPipelineString(au aurora.Aurora, cg *color.DynamicColorGenerator) string {
	c := cg.GetColorForElement(p.PipelineName)
	return au.Sprintf(au.Index(c, p.PipelineName))
}

func (p ParsedEventPrefix) getLoopString(au aurora.Aurora, cg *color.DynamicColorGenerator) string {
	if p.LoopIndex == nil || p.StepName == nil {
		return ""
	}

	key := fmt.Sprintf("%s.%s.%s.%d", p.PipelineName, *p.StepName, types.SafeString(p.ForEachKey), *p.LoopIndex)
	c := cg.GetColorForElement(key)
	return au.Sprintf(au.Index(c, *p.LoopIndex))
}

func (p ParsedEventPrefix) getForEachString(loopString string, au aurora.Aurora, cg *color.DynamicColorGenerator) string {
	if p.ForEachKey == nil || p.StepName == nil {
		return ""
	}

	key := fmt.Sprintf("%s.%s.%s", p.PipelineName, *p.StepName, *p.ForEachKey)
	c := cg.GetColorForElement(key)

	if loopString != "" {
		return au.Sprintf("%s%s%s", au.Index(c, *p.ForEachKey+"["), loopString, au.Index(c, "]"))
	} else {
		return au.Sprintf(au.Index(c, *p.ForEachKey))
	}
}

func (p ParsedEventPrefix) getStepString(eachString string, loopString string, au aurora.Aurora, cg *color.DynamicColorGenerator) string {
	if p.StepName == nil {
		return ""
	}

	key := fmt.Sprintf("%s.%s", p.PipelineName, *p.StepName)
	c := cg.GetColorForElement(key)
	if eachString != "" {
		return au.Sprintf("%s%s%s", au.Index(c, *p.StepName+"["), eachString, au.Index(c, "]"))
	} else if loopString != "" {
		return au.Sprintf("%s%s%s", au.Index(c, *p.StepName+"["), loopString, au.Index(c, "]"))
	} else {
		return fmt.Sprintf("%s", au.Index(c, *p.StepName))
	}
}

func (p ParsedEventPrefix) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}
	cg := opts.ColorGenerator

	retryString := p.getRetryString(au)
	loopString := p.getLoopString(au, cg)
	eachString := p.getForEachString(loopString, au, cg)
	stepString := p.getStepString(eachString, loopString, au, cg)
	pipelineString := p.getPipelineString(au, cg)

	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	sep := au.BrightBlack(".")

	var out string
	if stepString == "" {
		out = fmt.Sprintf("%s%s%s", left, pipelineString, right)
	} else {
		out = fmt.Sprintf("%s%s%s%s%s%s", left, pipelineString, sep, stepString, retryString, right)
	}

	if !helpers.IsNil(p.prefix) {
		pre := *p.prefix
		out = fmt.Sprintf("%s%s", pre.String(sanitizer, opts), out)
	}

	return out
}

type ParsedEvent struct {
	ParsedEventPrefix
	Type     string `json:"event_type"`
	StepType string `json:"step_type"`
	Message  string `json:"message,omitempty"`
	execId   string
}

func NewParsedEvent(prefix ParsedEventPrefix, executionId string, eventType string, stepType string, msg string) ParsedEvent {
	return ParsedEvent{
		ParsedEventPrefix: prefix,
		execId:            executionId,
		Type:              eventType,
		StepType:          stepType,
		Message:           msg,
	}
}

func (p ParsedEvent) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	pre := p.ParsedEventPrefix.String(sanitize.NullSanitizer, opts)
	out := ""

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return out
	}

	out += fmt.Sprintf("%s %s\n", pre, p.Message)
	return out
}

type ParsedEventWithInput struct {
	ParsedEvent
	Input  map[string]any `json:"args"`
	isSkip bool
}

func NewParsedEventWithInput(pe ParsedEvent, input map[string]any, isSkip bool) ParsedEventWithInput {
	return ParsedEventWithInput{
		ParsedEvent: pe,
		Input:       input,
		isSkip:      isSkip,
	}
}

func (p ParsedEventWithInput) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	out := ""
	au := aurora.NewAurora(opts.ColorEnabled)
	pre := p.ParsedEventPrefix.String(sanitize.NullSanitizer, opts)

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return out
	}

	if p.isSkip {
		return fmt.Sprintf("%s %s\n", pre, au.BrightBlack("Skipped"))
	}

	initText := "Starting"
	if p.RetryIndex != nil {
		initText = "Retrying"
	}

	switch p.StepType {
	case "http":
		method, _ := p.Input["method"].(string)
		url, _ := p.Input["url"].(string)
		if method == "" {
			method = "GET"
		} else {
			method = strings.ToUpper(method)
		}

		out += fmt.Sprintf("%s %s %s: %s %s\n", pre, initText, p.StepType, au.BrightBlack(method), au.BrightBlack(url))
	case "sleep":
		duration, _ := p.Input["duration"].(string)
		out += fmt.Sprintf("%s %s %s: %s\n", pre, initText, p.StepType, au.BrightBlack(duration))
	case "input":
		summary, additional := parseInputStepNotifierToLines(p.Input, opts)
		out += fmt.Sprintf("%s %s %s: %s\n", pre, initText, p.StepType, summary)
		if opts.Verbose && !helpers.IsNil(additional) {
			for _, line := range *additional {
				out += fmt.Sprintf("%s %s\n", pre, line)
			}
		}
	case "message":
		text, _ := p.Input["text"].(string)
		displayText := text
		if len(displayText) > 50 {
			displayText = displayText[:50] + "…"
		}
		out += fmt.Sprintf("%s %s %s: %s\n", pre, initText, p.StepType, au.BrightBlack(displayText))
		if !opts.Verbose { // arg will be shown in verbose mode anyway, no need for extra parsing
			if stepNotifierHasHttp(p.Input) {
				out += fmt.Sprintf("%s %s %s = %s\n", pre, "Arg", au.Blue("text"), formatSimpleValue(text, au))
			}
		}
	default:
		out += fmt.Sprintf("%s %s %s\n", pre, initText, p.StepType)
	}

	// args
	if opts.Verbose && len(p.Input) > 0 {
		out += sortAndParseMap(p.Input, "Arg", pre, au, opts)
	}

	return out
}

type ParsedEventWithOutput struct {
	ParsedEvent
	Output         map[string]any `json:"attributes"`
	StepOutput     map[string]any `json:"step_output"`
	Duration       *string        `json:"duration,omitempty"`
	isClosingEvent bool
}

func NewParsedEventWithOutput(parsedEvent ParsedEvent, output map[string]any, stepOutput map[string]any, duration *string, isClosingEvent bool) ParsedEventWithOutput {
	return ParsedEventWithOutput{
		ParsedEvent:    parsedEvent,
		Output:         output,
		StepOutput:     stepOutput,
		Duration:       duration,
		isClosingEvent: isClosingEvent,
	}
}

func (p ParsedEventWithOutput) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	out := ""
	au := aurora.NewAurora(opts.ColorEnabled)
	pre := p.ParsedEventPrefix.String(sanitize.NullSanitizer, opts)

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return out
	}

	// attributes
	if opts.Verbose && len(p.Output) > 0 {
		out += sortAndParseMap(p.Output, "Attr", pre, au, opts)
	}

	// outputs
	out += sortAndParseMap(p.StepOutput, "Output", pre, au, opts)

	duration := ""
	if p.Duration != nil {
		duration = *p.Duration
	}

	switch p.StepType {
	case "http":
		statusCode, _ := p.Output["status_code"].(float64)
		out += fmt.Sprintf("%s %s %g %s\n", pre, au.BrightGreen("Complete:"), au.BrightGreen(statusCode), au.Cyan(duration).Italic())
	case "query":
		if !helpers.IsNil(p.Output["rows"]) {
			rows := len(p.Output["rows"].([]any))
			out += fmt.Sprintf("%s %s %d %s %s\n", pre, au.BrightGreen("Complete:"), au.Cyan(rows), au.BrightGreen("row(s)"), au.Cyan(duration).Italic())
		} else {
			out += fmt.Sprintf("%s %s %s\n", pre, au.BrightGreen("Complete"), au.Cyan(duration).Italic())
		}
	default:
		additionalText := ""
		if p.isClosingEvent {
			additionalText = fmt.Sprintf(" %s", p.execId)
		}
		out += fmt.Sprintf("%s %s %s%s\n", pre, au.BrightGreen("Complete"), au.Cyan(duration).Italic(), au.BrightBlack(additionalText))
	}

	return out
}

// ParsedErrorEvent is a ParsedEvent which Failed.
type ParsedErrorEvent struct {
	ParsedEvent
	Errors          []resources.StepError `json:"errors"`
	Output          map[string]any        `json:"attributes"`
	Duration        *string               `json:"duration,omitempty"`
	isClosingEvent  bool
	retriesComplete bool
}

func NewParsedErrorEvent(parsedEvent ParsedEvent, errors []resources.StepError, output map[string]any, duration *string, isClosingEvent bool, retriesComplete bool) ParsedErrorEvent {
	return ParsedErrorEvent{
		ParsedEvent:     parsedEvent,
		Errors:          errors,
		Output:          output,
		Duration:        duration,
		isClosingEvent:  isClosingEvent,
		retriesComplete: retriesComplete,
	}
}

func (p ParsedErrorEvent) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	pre := p.ParsedEventPrefix.String(sanitize.NullSanitizer, opts)
	out := ""

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return out
	}

	duration := ""
	if p.Duration != nil {
		duration = *p.Duration
	}

	if (p.Type == event.HandlerPipelineFailed || opts.Verbose) && len(p.Output) > 0 {
		out += sortAndParseMap(p.Output, "Output", pre, au, opts)
	}
	for _, e := range p.Errors {
		out += fmt.Sprintf("%s %s %s\n", pre, au.Red(e.Error.Title+":"), au.Red(e.Error.Detail))
	}

	additionalText := ""
	if p.isClosingEvent {
		additionalText = fmt.Sprintf(" %s", p.execId)
	}

	if p.retriesComplete {
		errStr := au.Sprintf(au.Red("Failed").Bold())
		if len(p.Errors) > 1 {
			errStr = au.Sprintf(au.Red("Failed with %d errors").Bold(), len(p.Errors))
		}
		out += fmt.Sprintf("%s %s %s%s\n", pre, errStr, au.Cyan(duration).Italic(), au.BrightBlack(additionalText))
	}
	return out
}

type ParsedEventRegistryItem struct {
	Name    string
	Started time.Time
	Args    *resources.Input
}

type PrintableParsedEvent struct {
	Items             []sanitize.SanitizedStringer
	Registry          map[string]ParsedEventRegistryItem
	initialPipelineId string
}

func NewPrintableParsedEvent(pipelineId string) *PrintableParsedEvent {
	return &PrintableParsedEvent{
		Registry:          make(map[string]ParsedEventRegistryItem),
		initialPipelineId: pipelineId,
	}
}

func (p *PrintableParsedEvent) GetItems() []sanitize.SanitizedStringer {
	return p.Items
}

func (p *PrintableParsedEvent) SetEvents(logs ProcessEventLogs) (string, error) {
	var out []sanitize.SanitizedStringer

	lastStatus := ""

	for _, log := range logs {
		jsonPayload, err := json.Marshal(log.Detail)
		if err != nil {
			slog.Error("Error marshalling JSON", "error", err)
			return lastStatus, perr.InternalWithMessage("Error marshalling JSON")
		}

		switch log.Message {
		case event.HandlerExecutionFinished, event.HandlerExecutionFailed, event.HandlerExecutionPaused, event.HandlerExecutionCancelled:
			lastStatus = log.Message
		case event.HandlerPipelineQueued:
			var e event.PipelineQueued
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, perr.InternalWithMessage("Error unmarshalling JSON for pipeline queued event")
			}
			p.Registry[e.PipelineExecutionID] = ParsedEventRegistryItem{e.Name, e.Event.CreatedAt, &e.Args}
		case event.HandlerPipelineStarted:
			var e event.PipelineStarted
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, perr.InternalWithMessage("Error unmarshalling JSON for pipeline started event")
			}
			fullName := "unknown.unknown"
			var args resources.Input
			if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
				p.Registry[e.PipelineExecutionID] = ParsedEventRegistryItem{entry.Name, e.Event.CreatedAt, entry.Args}
				fullName = entry.Name
				args = *entry.Args
			}
			parsed := ParsedEventWithInput{
				ParsedEvent: ParsedEvent{
					ParsedEventPrefix: NewPrefix(fullName),
					Type:              log.Message,
					StepType:          "pipeline",
					execId:            e.Event.ExecutionID,
				},
				Input:  args,
				isSkip: false,
			}
			out = append(out, parsed)
		case event.HandlerPipelineFinished:
			var e event.PipelineFinished
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, perr.InternalWithMessage("Error unmarshalling JSON for pipeline finished event")
			}
			fullName := "unknown.unknown"
			started := e.Event.CreatedAt
			if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
				fullName = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
				started = entry.Started
			}
			duration := utils.HumanizeDuration(e.Event.CreatedAt.Sub(started))

			parsed := ParsedEventWithOutput{
				ParsedEvent: ParsedEvent{
					ParsedEventPrefix: NewPrefix(fullName),
					Type:              log.Message,
					execId:            e.Event.ExecutionID,
				},
				Duration:       &duration,
				StepOutput:     e.PipelineOutput,
				isClosingEvent: p.initialPipelineId == e.PipelineExecutionID,
			}
			out = append(out, parsed)
		case event.HandlerPipelineFailed:
			var e event.PipelineFailed
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, perr.InternalWithMessage("Error unmarshalling JSON for pipeline failed event")
			}
			fullName := "unknown.unknown"
			started := e.Event.CreatedAt
			if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
				fullName = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
				started = entry.Started
			}
			duration := utils.HumanizeDuration(e.Event.CreatedAt.Sub(started))

			allErrors := e.Errors
			pipelineOutputErrors, ok := e.PipelineOutput["errors"].([]resources.StepError)
			if ok && len(pipelineOutputErrors) > 0 {

				for _, e := range pipelineOutputErrors {
					found := false
					for _, ae := range allErrors {
						if e.Error.ID == ae.Error.ID {
							found = true
							break
						}
					}
					if !found {
						allErrors = append(allErrors, e)
					}
				}
			}

			parsed := ParsedErrorEvent{
				ParsedEvent: ParsedEvent{
					ParsedEventPrefix: ParsedEventPrefix{
						FullPipelineName: fullName,
						PipelineName:     strings.Split(fullName, ".")[len(strings.Split(fullName, "."))-1],
					},
					Type:   log.Message,
					execId: e.Event.ExecutionID,
				},
				Duration:        &duration,
				Errors:          allErrors,
				Output:          e.PipelineOutput,
				retriesComplete: true,
			}
			out = append(out, parsed)
		case event.HandlerStepQueued:
			var e event.StepQueued
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, perr.InternalWithMessage("Error unmarshalling JSON for step queued event")
			}
			p.Registry[e.StepExecutionID] = ParsedEventRegistryItem{
				Name:    e.StepName,
				Started: e.Event.CreatedAt,
			}
		case event.CommandStepStart:
			var e event.StepStart
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, perr.InternalWithMessage("Error unmarshalling JSON for step start event")
			}

			p.Registry[e.StepExecutionID] = ParsedEventRegistryItem{e.StepName, e.Event.CreatedAt, &e.StepInput}

			pipeline := p.Registry[e.PipelineExecutionID]
			fullStepName := e.StepName
			stepType := strings.Split(e.StepName, ".")[0]
			stepName := strings.Split(e.StepName, ".")[1]

			prefix := NewPrefix(pipeline.Name)
			prefix.FullStepName = &fullStepName
			prefix.StepName = &stepName
			if e.StepForEach != nil && e.StepForEach.ForEachStep {
				prefix.ForEachKey = &e.StepForEach.Key
			}
			if e.StepLoop != nil {
				prefix.LoopIndex = &e.StepLoop.Index
			}
			if e.StepRetry != nil {
				i := e.StepRetry.Count + 1
				prefix.RetryIndex = &i
			}

			parsed := ParsedEventWithInput{
				ParsedEvent: ParsedEvent{
					ParsedEventPrefix: prefix,
					Type:              log.Message,
					StepType:          stepType,
					execId:            e.Event.ExecutionID,
				},
				Input:  e.StepInput,
				isSkip: e.NextStepAction == "skip",
			}
			out = append(out, parsed)
		case event.HandlerStepFinished:
			var e event.StepFinished
			err := json.Unmarshal(jsonPayload, &e)
			if err != nil {
				return lastStatus, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
			}

			if e.Output != nil && e.Output.Status != "skipped" {
				pipeline := p.Registry[e.PipelineExecutionID]
				step := p.Registry[e.StepExecutionID]
				stepType := strings.Split(step.Name, ".")[0]
				stepName := strings.Split(step.Name, ".")[1]
				duration := utils.HumanizeDuration(e.Event.CreatedAt.Sub(step.Started))

				prefix := NewPrefix(pipeline.Name)
				prefix.FullStepName = &step.Name
				prefix.StepName = &stepName
				if e.StepForEach != nil && e.StepForEach.ForEachStep {
					prefix.ForEachKey = &e.StepForEach.Key
				}
				if e.StepLoop != nil {
					if e.StepLoop.LoopCompleted {
						prefix.LoopIndex = &e.StepLoop.Index
					} else {
						i := e.StepLoop.Index - 1
						prefix.LoopIndex = &i
					}
				}
				if e.StepRetry != nil {
					i := e.StepRetry.Count
					prefix.RetryIndex = &i

				}
				if helpers.IsNil(e.Output.Data) {
					e.Output.Data = resources.OutputData{}
				}
				if e.Output.Flowpipe != nil {
					e.Output.Data["flowpipe"] = e.Output.Flowpipe
				}
				switch e.Output.Status {
				case "finished":
					parsed := ParsedEventWithOutput{
						ParsedEvent: ParsedEvent{
							ParsedEventPrefix: prefix,
							Type:              log.Message,
							StepType:          stepType,
							execId:            e.Event.ExecutionID,
						},
						Duration:   &duration,
						Output:     e.Output.Data,
						StepOutput: e.StepOutput,
					}
					out = append(out, parsed)
				case "failed":
					rc := true
					if e.StepRetry != nil {
						rc = e.StepRetry.RetryCompleted
					}
					parsed := ParsedErrorEvent{
						ParsedEvent: ParsedEvent{
							ParsedEventPrefix: prefix,
							Type:              log.Message,
							StepType:          stepType,
							execId:            e.Event.ExecutionID,
						},
						Duration:        &duration,
						Errors:          e.Output.Errors,
						Output:          e.Output.Data,
						retriesComplete: rc,
					}
					out = append(out, parsed)
				}
			}

		default:
			// ignore other events
		}
	}

	p.Items = out
	return lastStatus, nil

}

func (p *PrintableParsedEvent) GetTable() (*printers.Table, error) {
	return printers.NewTable(), nil
}

type ProcessEventLogs []event.EventLogImpl

// GetResourceType is used to satisfy the interface requirements of types.PrintableResource Transform function
func (ProcessEventLogs) GetResourceType() string {
	return "ProcessEventLogs"
}

func isSimpleType(input any) bool {
	kind := reflect.TypeOf(input).Kind()
	switch kind {
	case
		reflect.Bool,
		reflect.String,
		reflect.Float32,
		reflect.Float64,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return true
	default:
		return false
	}
}

func formatSimpleValue(input any, au aurora.Aurora) string {
	kind := reflect.TypeOf(input).Kind()
	switch kind {
	case reflect.Bool:
		return au.Sprintf("%t", au.Yellow(input))
	case reflect.String:
		return au.Sprintf("%s", au.Green(input))
	case
		reflect.Float32,
		reflect.Float64:
		return au.Sprintf("%g", au.Cyan(input))
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return au.Sprintf("%d", au.Cyan(input))
	default:
		return ""
	}
}

func sortAndParseMap(input map[string]any, typeString string, prefix string, au aurora.Aurora, opts sanitize.RenderOptions) string {
	out := ""
	sortedKeys := utils.SortedMapKeys(input)
	if typeString != "" {
		typeString = fmt.Sprintf("%s ", typeString)
	}
	if prefix != "" {
		prefix = fmt.Sprintf("%s ", prefix)
	}
	for _, key := range sortedKeys {

		// Nasty .. but form_url is a special case where we "extend the input" (see extendInput function). It need to be removed
		// because it's not a real input to the step.
		if key == constants.FormUrl {
			continue
		}

		v := input[key]
		if v == nil {
			v = ""
		}
		valueString := ""
		if isSimpleType(v) {
			valueString = formatSimpleValue(v, au)
		} else {
			s, err := opts.JsonFormatter.Marshal(v)
			if err != nil {
				valueString = au.Sprintf(au.Red("error parsing value"))
			} else {
				valueString = string(s)
			}
		}

		out += fmt.Sprintf("%s%s%s = %s\n", prefix, typeString, au.Blue(key), valueString)
	}
	return out
}

func parseInputStepNotifierToLines(input resources.Input, opts sanitize.RenderOptions) (string, *[]string) {
	au := aurora.NewAurora(opts.ColorEnabled)
	formUrl, hasFormUrl := input[constants.FormUrl].(string)
	if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
		if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
			switch len(notifies) {
			case 0:
				return formUrl, nil
			case 1: // single notify, give summary
				notify := notifies[0].(map[string]any)
				integration := notify["integration"].(map[string]any)
				integrationType := integration["type"].(string)
				switch integrationType {
				case schema.IntegrationTypeHttp:
					return formUrl, nil
				case schema.IntegrationTypeEmail:
					var to []string
					additionalLines := []string{fmt.Sprintf("Form URL: %s", au.BrightBlack(formUrl))}
					if sTo, ok := input[schema.AttributeTypeTo].([]any); ok {
						for _, t := range sTo {
							to = append(to, t.(string))
						}
					} else if nTo, ok := notify[schema.AttributeTypeTo].([]any); ok {
						for _, t := range nTo {
							to = append(to, t.(string))
						}
					} else if iTo, ok := integration[schema.AttributeTypeTo].([]any); ok {
						for _, t := range iTo {
							to = append(to, t.(string))
						}
					}

					switch len(to) {
					case 0:
						return fmt.Sprintf("email to %s", au.BrightBlack("cc/bcc only")), &additionalLines
					case 1:
						return fmt.Sprintf("email to %s", au.BrightBlack(to[0])), &additionalLines
					default:
						remainder := len(to) - 1
						return fmt.Sprintf("email to %s", au.BrightBlack(fmt.Sprintf("%s + %d others", to[0], remainder))), &additionalLines
					}
				case schema.IntegrationTypeSlack:
					channel := ""
					if sChannel, ok := input[schema.AttributeTypeChannel].(string); ok {
						channel = sChannel
					} else if nChannel, ok := notify[schema.AttributeTypeChannel].(string); ok {
						channel = nChannel
					} else if iChannel, ok := integration[schema.AttributeTypeChannel].(string); ok {
						channel = iChannel
					}
					return fmt.Sprintf("slack to %s", au.BrightBlack(channel)), nil
				case schema.IntegrationTypeMsTeams:
					return fmt.Sprintf("msteams via %s", au.BrightBlack("webhook")), nil
				}

			default: // multiple notifies
				var notifyTypes []string
				var additionalLines []string
				if hasFormUrl {
					additionalLines = append(additionalLines, fmt.Sprintf("Form URL: %s", au.BrightBlack(formUrl)))
				}

				for i, n := range notifies {
					notify := n.(map[string]any)
					integration := notify["integration"].(map[string]any)
					integrationType := integration["type"].(string)
					notifyTypes = append(notifyTypes, integrationType)
					var prefix string
					if i == 0 {
						prefix = fmt.Sprintf("Notified via %s", au.BrightBlack(integrationType))
					} else {
						prefix = fmt.Sprintf("Notified #%d via %s", i+1, au.BrightBlack(integrationType))
					}
					switch integrationType {
					case schema.IntegrationTypeEmail:
						var to []string
						if sTo, ok := input[schema.AttributeTypeTo].([]any); ok {
							for _, t := range sTo {
								to = append(to, t.(string))
							}
						} else if nTo, ok := notify[schema.AttributeTypeTo].([]any); ok {
							for _, t := range nTo {
								to = append(to, t.(string))
							}
						} else if iTo, ok := integration[schema.AttributeTypeTo].([]any); ok {
							for _, t := range iTo {
								to = append(to, t.(string))
							}
						}
						switch len(to) {
						case 0:
						case 1, 2, 3:
							additionalLines = append(additionalLines, fmt.Sprintf("%s to %s", prefix, au.BrightBlack(strings.Join(to, ", "))))
						default:
							r := len(to) - 3
							f3 := to[0:3]
							additionalLines = append(additionalLines, fmt.Sprintf("%s to %s %s", prefix, au.BrightBlack(strings.Join(f3, ", ")), au.BrightBlack(fmt.Sprintf("+ %d others", r))))
						}

						var cc []string
						if sCc, ok := input[schema.AttributeTypeCc].([]any); ok {
							for _, t := range sCc {
								cc = append(cc, t.(string))
							}
						} else if nCc, ok := notify[schema.AttributeTypeCc].([]any); ok {
							for _, t := range nCc {
								cc = append(cc, t.(string))
							}
						} else if iCc, ok := integration[schema.AttributeTypeCc].([]any); ok {
							for _, t := range iCc {
								cc = append(cc, t.(string))
							}
						}
						switch len(cc) {
						case 0:
						case 1, 2, 3:
							additionalLines = append(additionalLines, fmt.Sprintf("%s to %s", prefix, au.BrightBlack(strings.Join(cc, ", "))))
						default:
							r := len(cc) - 3
							f3 := cc[0:3]
							additionalLines = append(additionalLines, fmt.Sprintf("%s to %s %s", prefix, au.BrightBlack(strings.Join(f3, ", ")), au.BrightBlack(fmt.Sprintf("+ %d others", r))))
						}

						var bcc []string
						if sBcc, ok := input[schema.AttributeTypeBcc].([]any); ok {
							for _, t := range sBcc {
								bcc = append(bcc, t.(string))
							}
						} else if nBcc, ok := notify[schema.AttributeTypeBcc].([]any); ok {
							for _, t := range nBcc {
								bcc = append(bcc, t.(string))
							}
						} else if iBcc, ok := integration[schema.AttributeTypeBcc].([]any); ok {
							for _, t := range iBcc {
								bcc = append(bcc, t.(string))
							}
						}
						switch len(bcc) {
						case 0:
						case 1, 2, 3:
							additionalLines = append(additionalLines, fmt.Sprintf("%s to %s", prefix, au.BrightBlack(strings.Join(bcc, ", "))))
						default:
							r := len(bcc) - 3
							f3 := bcc[0:3]
							additionalLines = append(additionalLines, fmt.Sprintf("%s to %s %s", prefix, au.BrightBlack(strings.Join(f3, ", ")), au.BrightBlack(fmt.Sprintf("+ %d others", r))))
						}
					case schema.IntegrationTypeHttp:
						additionalLines = append(additionalLines, prefix)
					case schema.IntegrationTypeSlack:
						var channel string
						if sChannel, ok := input[schema.AttributeTypeChannel].(string); ok {
							channel = sChannel
						} else if nChannel, ok := notify[schema.AttributeTypeChannel].(string); ok {
							channel = nChannel
						} else if iChannel, ok := integration[schema.AttributeTypeChannel].(string); ok {
							channel = iChannel
						}
						additionalLines = append(additionalLines, fmt.Sprintf("%s channel %s", prefix, au.BrightBlack(channel)))
					case schema.IntegrationTypeMsTeams:
						additionalLines = append(additionalLines, fmt.Sprintf("%s via %s", prefix, au.BrightBlack("webhook")))
					}
				}

				return fmt.Sprintf("Notified via %s", au.BrightBlack(strings.Join(notifyTypes, ", "))), &additionalLines
			}
		}
	}
	return formUrl, nil
}

func stepNotifierHasHttp(input resources.Input) bool {
	if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
		if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
			for _, n := range notifies {
				notify := n.(map[string]any)
				integration := notify["integration"].(map[string]any)
				integrationType := integration["type"].(string)
				if integrationType == schema.IntegrationTypeHttp {
					return true
				}
			}
		}
	}
	return false
}
