package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hokaccha/go-prettyjson"
	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/color"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
)

type ColorOptions struct {
	ColorEnabled   bool
	ColorGenerator *color.DynamicColorGenerator
}

type SanitizedStringer interface {
	String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string
}

// const grayScaleIndex = uint8(3)

type ParsedHeader struct {
	ExecutionId string `json:"execution_id"`
	IsStale     bool   `json:"is_stale"`
	LastLoaded  string `json:"last_loaded"`
}

func (p ParsedHeader) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)

	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	left := "["  // au.Gray(grayScaleIndex, "[")
	right := "]" // au.Gray(grayScaleIndex, "]")
	out := fmt.Sprintf("%s%s%s %s\n", left, au.BrightGreen("Execution"), right, p.ExecutionId)
	if p.IsStale {
		out += fmt.Sprintf("%s%s%s %s\n", left, au.BrightRed("Stale"), right, au.Sprintf(au.Red("Mod is stale, last loaded: %s"), p.LastLoaded))
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
}

func NewPrefix(fullPipelineName string) ParsedEventPrefix {
	return ParsedEventPrefix{
		FullPipelineName: fullPipelineName,
		PipelineName:     strings.Split(fullPipelineName, ".")[len(strings.Split(fullPipelineName, "."))-1],
	}
}

func (p ParsedEventPrefix) getRetryString(au aurora.Aurora) string {
	if p.RetryIndex == nil || *p.RetryIndex == 0 {
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

func (p ParsedEventPrefix) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
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

	left := "["  // au.Gray(grayScaleIndex, "[")
	right := "]" // au.Gray(grayScaleIndex, "]")
	sep := "."   // au.Gray(grayScaleIndex, ".")

	if stepString == "" {
		return fmt.Sprintf("%s%s%s", left, pipelineString, right)
	} else {
		return fmt.Sprintf("%s%s%s%s%s%s", left, pipelineString, sep, stepString, retryString, right)
	}
}

type ParsedEvent struct {
	ParsedEventPrefix
	Type     string `json:"event_type"`
	StepType string `json:"step_type"`
	Message  string `json:"message"`
}

func (p ParsedEvent) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	out := ""
	pre := p.ParsedEventPrefix.String(sanitizer, opts)

	out += fmt.Sprintf("%s %s\n", pre, p.Message)
	return out
}

type ParsedEventWithInput struct {
	ParsedEvent
	Input map[string]any `json:"input"`
}

func (p ParsedEventWithInput) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	out := ""
	pre := p.ParsedEventPrefix.String(sanitizer, opts)

	stepString := ""
	if p.StepType != "" {
		stepString = fmt.Sprintf(" %s step", au.Blue(p.StepType))
	}

	out += fmt.Sprintf("%s Starting%s\n", pre, stepString)
	for k, v := range p.Input {
		if v == nil {
			v = ""
		}
		valueString := ""
		if isSimpleType(v) {
			valueString = fmt.Sprintf("%v", v)
		} else {
			s, err := prettyjson.Marshal(v)
			if err != nil {
				valueString = au.Sprintf(au.Red("error parsing value"))
			} else {
				valueString = string(s)
			}
		}
		out += fmt.Sprintf("%s Arg %s = %s\n", pre, au.Blue(k), au.BrightBlue(valueString))
	}
	return out
}

type ParsedEventWithArgs struct {
	ParsedEvent
	Args map[string]any `json:"args"`
}

func (p ParsedEventWithArgs) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	out := ""
	pre := p.ParsedEventPrefix.String(sanitizer, opts)

	out += fmt.Sprintf("%s Starting\n", pre)
	for k, v := range p.Args {
		if v == nil {
			v = ""
		}
		valueString := ""
		if isSimpleType(v) {
			valueString = fmt.Sprintf("%v", au.BrightBlue(v))
		} else {
			s, err := prettyjson.Marshal(v)
			if err != nil {
				valueString = au.Sprintf(au.Red("error parsing value"))
			} else {
				valueString = string(s)
			}
		}
		out += fmt.Sprintf("%s Arg %s = %s\n", pre, au.Blue(k), valueString)
	}
	return out
}

type ParsedEventWithOutput struct {
	ParsedEvent
	Output   map[string]any
	Duration *string
}

func (p ParsedEventWithOutput) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	out := ""
	pre := p.ParsedEventPrefix.String(sanitizer, opts)

	if p.Type == event.HandlerPipelineFinished {
		for k, v := range p.Output {
			if v == nil {
				v = ""
			}
			valueString := ""
			if isSimpleType(v) {
				valueString = fmt.Sprintf("%v", au.BrightBlue(v))
			} else {
				s, err := prettyjson.Marshal(v)
				if err != nil {
					valueString = au.Sprintf(au.Red("error parsing value"))
				} else {
					valueString = string(s)
				}
			}
			out += fmt.Sprintf("%s %s %s = %s\n", pre, "Output", au.Blue(k), valueString)
		}
	}
	duration := ""
	if p.Duration != nil {
		duration = *p.Duration
	}
	out += fmt.Sprintf("%s %s %s\n", pre, au.BrightGreen("Complete"), au.Cyan(duration).Italic())
	return out
}

// ParsedErrorEvent is a ParsedEvent which Failed.
type ParsedErrorEvent struct {
	ParsedEvent
	Errors   []modconfig.StepError `json:"errors"`
	Output   map[string]any
	Duration *string `json:"duration,omitempty"`
}

func (p ParsedErrorEvent) String(sanitizer *sanitize.Sanitizer, opts ColorOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	out := ""
	pre := p.ParsedEventPrefix.String(sanitizer, opts)
	duration := ""
	if p.Duration != nil {
		duration = *p.Duration
	}

	if p.Type == event.HandlerPipelineFailed {
		for k, v := range p.Output {
			if v == nil {
				v = ""
			}
			valueString := ""
			if isSimpleType(v) {
				valueString = fmt.Sprintf("%v", au.BrightBlue(v))
			} else {
				s, err := prettyjson.Marshal(v)
				if err != nil {
					valueString = au.Sprintf(au.Red("error parsing value"))
				} else {
					valueString = string(s)
				}
			}
			out += fmt.Sprintf("%s %s %s = %s\n", pre, "Output", au.Blue(k), valueString)
		}
	}

	for _, e := range p.Errors {
		out += fmt.Sprintf("%s %s: %s\n", pre, au.BrightRed(e.Error.Title), au.Red(e.Error.Detail))
	}
	out += fmt.Sprintf("%s %s %s\n", pre, au.Sprintf(au.Red("Failed with %d error(s)"), len(p.Errors)), au.Cyan(duration).Italic())
	return out
}

type ParsedEventRegistryItem struct {
	Name    string
	Started time.Time
}

type PrintableParsedEvent struct {
	Items    []SanitizedStringer
	Registry map[string]ParsedEventRegistryItem
}

func NewPrintableParsedEvent() *PrintableParsedEvent {
	return &PrintableParsedEvent{
		Registry: make(map[string]ParsedEventRegistryItem),
	}
}

func (p *PrintableParsedEvent) GetItems() []SanitizedStringer {
	return p.Items
}

func (p *PrintableParsedEvent) SetEvents(logs ProcessEventLogs) error {
	var out []SanitizedStringer

	for _, log := range logs {
		switch log.EventType {
		case event.HandlerPipelineQueued:
			var e event.PipelineQueued
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
			}
			p.Registry[e.PipelineExecutionID] = ParsedEventRegistryItem{e.Name, e.Event.CreatedAt}
		case event.HandlerPipelineStarted:
			var e event.PipelineStarted
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
			}
			fullName := "unknown.unknown"
			if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
				p.Registry[e.PipelineExecutionID] = ParsedEventRegistryItem{entry.Name, e.Event.CreatedAt}
				fullName = entry.Name
			}
			parsed := ParsedEvent{
				ParsedEventPrefix: NewPrefix(fullName),
				Type:              log.EventType,
				Message:           fmt.Sprintf("Starting: %s", e.PipelineExecutionID),
			}
			out = append(out, parsed)
		case event.HandlerPipelineFinished:
			var e event.PipelineFinished
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
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
					Type:              log.EventType,
				},
				Duration: &duration,
				Output:   e.PipelineOutput,
			}
			out = append(out, parsed)
		case event.HandlerPipelineFailed:
			var e event.PipelineFailed
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
			}
			fullName := "unknown.unknown"
			started := e.Event.CreatedAt
			if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
				fullName = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
				started = entry.Started
			}
			duration := utils.HumanizeDuration(e.Event.CreatedAt.Sub(started))

			allErrors := e.Errors
			pipelineOutputErrors, ok := e.PipelineOutput["errors"].([]modconfig.StepError)
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
					Type: log.EventType,
				},
				Duration: &duration,
				Errors:   allErrors,
				Output:   e.PipelineOutput,
			}
			out = append(out, parsed)
		case event.HandlerStepQueued:
			var e event.StepQueued
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
			}
			p.Registry[e.StepExecutionID] = ParsedEventRegistryItem{
				Name:    e.StepName,
				Started: e.Event.CreatedAt,
			}
		case event.CommandStepStart:
			var e event.StepStart
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
			}
			if e.NextStepAction == "start" { // TODO: handle 'skip' steps?
				p.Registry[e.StepExecutionID] = ParsedEventRegistryItem{e.StepName, e.Event.CreatedAt}

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
					prefix.RetryIndex = &e.StepRetry.Count
				}

				parsed := ParsedEventWithInput{
					ParsedEvent: ParsedEvent{
						ParsedEventPrefix: prefix,
						Type:              log.EventType,
						StepType:          stepType,
					},
					Input: e.StepInput,
				}
				out = append(out, parsed)
			}
		case event.HandlerStepFinished:
			var e event.StepFinished
			err := json.Unmarshal([]byte(log.Payload), &e)
			if err != nil {
				return fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
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
					prefix.RetryIndex = &e.StepRetry.Count
				}

				switch e.Output.Status {
				case "finished":
					parsed := ParsedEventWithOutput{
						ParsedEvent: ParsedEvent{
							ParsedEventPrefix: prefix,
							Type:              log.EventType,
							StepType:          stepType,
						},
						Duration: &duration,
						Output:   e.Output.Data,
					}
					out = append(out, parsed)
				case "failed":
					parsed := ParsedErrorEvent{
						ParsedEvent: ParsedEvent{
							ParsedEventPrefix: prefix,
							Type:              log.EventType,
							StepType:          stepType,
						},
						Duration: &duration,
						Errors:   e.Output.Errors,
					}
					out = append(out, parsed)
				}
			}
		default:
			// ignore other events
		}
	}

	p.Items = out
	return nil

}

func (p *PrintableParsedEvent) GetTable() (Table, error) {
	return Table{}, nil
}

type ProcessEventLogs []ProcessEventLog

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
