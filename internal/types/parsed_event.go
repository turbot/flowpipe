package types

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
	"reflect"
	"strings"
	"time"
)

type ParsedHeader struct {
	ExecutionId string
	IsStale     bool
	LastLoaded  string
}

func (p ParsedHeader) String() string {
	out := fmt.Sprintf("[%s] %s\n", color.HiGreenString("Execution"), p.ExecutionId)
	if p.IsStale {
		out += fmt.Sprintf("[%s] %s\n", color.HiRedString("Stale"), color.RedString("Mod is stale, last loaded: %s", p.LastLoaded))
	}
	return out
}

type ParsedEventPrefix struct {
	FullPipelineName string
	PipelineName     string
	FullStepName     *string
	StepName         *string
	ForEachKey       *string
	LoopIndex        *int
	RetryIndex       *int
}

func (p ParsedEventPrefix) String() string {
	plString := color.GreenString("%s", p.PipelineName)

	retryString := ""
	if p.RetryIndex != nil {
		retryString = color.BlueString("#%d", *p.RetryIndex)
	}

	loopString := ""
	if p.LoopIndex != nil {
		loopString = color.RedString("%d", *p.LoopIndex)
	}

	feString := ""
	if p.ForEachKey != nil {
		if loopString != "" {
			feString = fmt.Sprintf("%s%s%s", color.CyanString("%s[", *p.ForEachKey), loopString, color.CyanString("]"))
		} else {
			feString = color.CyanString("%s", *p.ForEachKey)
		}
	}

	stepString := ""
	if p.StepName != nil {
		if feString != "" {
			stepString = fmt.Sprintf("%s%s%s", color.MagentaString("%s[", *p.StepName), feString, color.MagentaString("]"))
		} else if loopString != "" {
			stepString = fmt.Sprintf("%s%s%s", color.MagentaString("%s[", *p.StepName), loopString, color.MagentaString("]"))
		} else {
			stepString = color.MagentaString("%s", *p.StepName)
		}
	}

	if stepString != "" {
		return fmt.Sprintf("[%s.%s]%s", plString, stepString, retryString)
	} else {
		return fmt.Sprintf("[%s]", plString)
	}
}

type ParsedEvent struct {
	ParsedEventPrefix
	Type     string
	StepType string
	Message  string
}

func (p ParsedEvent) String() string {
	out := ""
	pre := p.ParsedEventPrefix.String()

	out += fmt.Sprintf("%s %s\n", pre, p.Message)
	return out
}

type ParsedEventWithInput struct {
	ParsedEvent
	Input map[string]any
}

func (p ParsedEventWithInput) String() string {
	out := ""
	pre := p.ParsedEventPrefix.String()

	stepString := ""
	if p.StepType != "" {
		stepString = fmt.Sprintf(": %s step.", color.BlueString(p.StepType))
	}

	out += fmt.Sprintf("%s Starting%s\n", pre, stepString)
	for k, v := range p.Input {
		if v == nil {
			v = ""
		}
		valueString := ""
		if isSimpleType(v) {
			valueString = color.HiBlueString("%v", v)
		} else {
			s, err := prettyjson.Marshal(v)
			if err != nil {
				valueString = color.RedString("error parsing value")
			} else {
				valueString = string(s)
			}
		}
		out += fmt.Sprintf("%s Input: %s = %s\n", pre, color.BlueString(k), color.HiBlueString(valueString))
	}
	return out
}

type ParsedEventWithArgs struct {
	ParsedEvent
	Args map[string]any
}

func (p ParsedEventWithArgs) String() string {
	out := ""
	pre := p.ParsedEventPrefix.String()

	out += fmt.Sprintf("%s Starting\n", pre)
	for k, v := range p.Args {
		if v == nil {
			v = ""
		}
		valueString := ""
		if isSimpleType(v) {
			valueString = color.HiBlueString("%v", v)
		} else {
			s, err := prettyjson.Marshal(v)
			if err != nil {
				valueString = color.RedString("error parsing value")
			} else {
				valueString = string(s)
			}
		}
		out += fmt.Sprintf("%s Arg: %s = %s\n", pre, color.BlueString(k), color.HiBlueString(valueString))
	}
	return out
}

type ParsedEventWithOutput struct {
	ParsedEvent
	Output   map[string]any
	Duration *string
}

func (p ParsedEventWithOutput) String() string {
	out := ""
	pre := p.ParsedEventPrefix.String()

	if p.Type == event.HandlerPipelineFinished {
		for k, v := range p.Output {
			if v == nil {
				v = ""
			}
			valueString := ""
			if isSimpleType(v) {
				valueString = color.HiBlueString("%v", v)
			} else {
				s, err := prettyjson.Marshal(v)
				if err != nil {
					valueString = color.RedString("error parsing value")
				} else {
					valueString = string(s)
				}
			}
			out += fmt.Sprintf("%s %s: %s = %s\n", pre, "Output", color.BlueString(k), color.HiBlueString(valueString))
		}
	}
	duration := ""
	if p.Duration != nil {
		duration = *p.Duration
	}
	out += fmt.Sprintf("%s %s: %s\n", pre, color.HiGreenString("Complete"), color.YellowString("%s", duration))
	return out
}

// ParsedErrorEvent is a ParsedEvent which Failed.
type ParsedErrorEvent struct {
	ParsedEvent
	Errors   []modconfig.StepError
	Duration *string
}

func (p ParsedErrorEvent) String() string {
	out := ""
	pre := p.ParsedEventPrefix.String()

	if p.Type != event.HandlerPipelineFailed {
		for _, e := range p.Errors {
			out += fmt.Sprintf("%s %s: %s\n", pre, color.RedString(e.Error.Title), color.RedString(e.Error.Detail))
		}
	}

	duration := ""
	if p.Duration != nil {
		duration = *p.Duration
	}
	out += fmt.Sprintf("%s %s: %s\n", pre, color.HiRedString("Failed with %d error(s)", len(p.Errors)), color.YellowString("%s", duration))
	return out
}

type ParsedEventRegistryItem struct {
	Name    string
	Started time.Time
}

type PrintableParsedEvent struct {
	Items    interface{}
	Registry map[string]ParsedEventRegistryItem
}

func (p PrintableParsedEvent) GetItems() interface{} {
	return p.Items
}

func (p PrintableParsedEvent) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {
	resourceType := r.GetResourceType()
	if resourceType != "ProcessEventLogs" {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("invalid resource type: %s", resourceType))
	}

	var out []any

	if logs, ok := r.(ProcessEventLogs); ok {
		for _, log := range logs {
			switch log.EventType {
			case event.HandlerPipelineQueued:
				var e event.PipelineQueued
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
				}
				p.Registry[e.PipelineExecutionID] = ParsedEventRegistryItem{e.Name, e.Event.CreatedAt}
			case event.HandlerPipelineStarted:
				var e event.PipelineStarted
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
				}
				fullName := "unknown.unknown"
				if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
					p.Registry[e.PipelineExecutionID] = ParsedEventRegistryItem{entry.Name, e.Event.CreatedAt}
					fullName = entry.Name
				}
				parsed := ParsedEvent{
					ParsedEventPrefix: ParsedEventPrefix{
						FullPipelineName: fullName,
						PipelineName:     strings.Split(fullName, ".")[len(strings.Split(fullName, "."))-1],
					},
					Type:    log.EventType,
					Message: fmt.Sprintf("Starting: %s", e.PipelineExecutionID),
				}
				out = append(out, parsed)
			case event.HandlerPipelineFinished:
				var e event.PipelineFinished
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
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
						ParsedEventPrefix: ParsedEventPrefix{
							FullPipelineName: fullName,
							PipelineName:     strings.Split(fullName, ".")[len(strings.Split(fullName, "."))-1],
						},
						Type: log.EventType,
					},
					Duration: &duration,
					Output:   e.PipelineOutput,
				}
				out = append(out, parsed)
			case event.HandlerPipelineFailed:
				var e event.PipelineFailed
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
				}
				fullName := "unknown.unknown"
				started := e.Event.CreatedAt
				if entry, exists := p.Registry[e.PipelineExecutionID]; exists {
					fullName = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
					started = entry.Started
				}
				duration := utils.HumanizeDuration(e.Event.CreatedAt.Sub(started))
				parsed := ParsedErrorEvent{
					ParsedEvent: ParsedEvent{
						ParsedEventPrefix: ParsedEventPrefix{
							FullPipelineName: fullName,
							PipelineName:     strings.Split(fullName, ".")[len(strings.Split(fullName, "."))-1],
						},
						Type: log.EventType,
					},
					Duration: &duration,
					Errors:   e.Errors,
				}
				out = append(out, parsed)
			case event.HandlerStepQueued:
				var e event.StepQueued
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
				}
				p.Registry[e.StepExecutionID] = ParsedEventRegistryItem{
					Name:    e.StepName,
					Started: e.Event.CreatedAt,
				}
			case event.CommandStepStart:
				var e event.StepStart
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
				}
				if e.NextStepAction == "start" { // TODO: handle 'skip' steps?
					p.Registry[e.StepExecutionID] = ParsedEventRegistryItem{e.StepName, e.Event.CreatedAt}

					pl := p.Registry[e.PipelineExecutionID]
					pipelineName := strings.Split(pl.Name, ".")[len(strings.Split(pl.Name, "."))-1]
					fullStepName := e.StepName
					stepType := strings.Split(e.StepName, ".")[0]
					stepName := strings.Split(e.StepName, ".")[1]

					prefix := ParsedEventPrefix{
						FullPipelineName: pl.Name,
						PipelineName:     pipelineName,
						FullStepName:     &fullStepName,
						StepName:         &stepName,
					}
					if e.StepForEach != nil && e.StepForEach.ForEachStep {
						prefix.ForEachKey = &e.StepForEach.Key
					}
					if e.StepLoop != nil {
						prefix.LoopIndex = &e.StepLoop.Index
					}
					if e.StepRetry != nil {
						prefix.RetryIndex = &e.StepRetry.Index
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
					return nil, fmt.Errorf("failed to unmarshal %s event: %v", e.HandlerName(), err)
				}

				if e.Output != nil && e.Output.Status != "skipped" {
					pipeline := p.Registry[e.PipelineExecutionID]
					step := p.Registry[e.StepExecutionID]
					pipelineName := strings.Split(pipeline.Name, ".")[len(strings.Split(pipeline.Name, "."))-1]
					stepType := strings.Split(step.Name, ".")[0]
					stepName := strings.Split(step.Name, ".")[1]
					duration := utils.HumanizeDuration(e.Event.CreatedAt.Sub(step.Started))

					prefix := ParsedEventPrefix{
						FullPipelineName: pipeline.Name,
						PipelineName:     pipelineName,
						FullStepName:     &step.Name,
						StepName:         &stepName,
					}
					if e.StepForEach != nil && e.StepForEach.ForEachStep {
						prefix.ForEachKey = &e.StepForEach.Key
					}
					if e.StepLoop != nil {
						prefix.LoopIndex = &e.StepLoop.Index
					}
					if e.StepRetry != nil {
						prefix.RetryIndex = &e.StepRetry.Index
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
	} else {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("error parsing resource type: %s", resourceType))
	}

	return out, nil
}

func (p PrintableParsedEvent) GetTable() (Table, error) {
	return Table{}, nil
}

func (PrintableParsedEvent) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{}
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
