package types

import (
	"encoding/json"
	"fmt"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
	"reflect"
	"strings"
	"time"
)

type LogLine struct {
	Name       string
	Message    string
	IsError    bool
	StepName   *string
	ForEachKey *string
	LoopIndex  *int
	RetryIndex *int // TODO: This isn't implemented in EventLog yet, placeholder.
}

type LogLineRegistryItem struct {
	Name    string
	Started time.Time
}

func LogLinesFromEventLog(eventLogEntry flowpipeapiclient.ProcessEventLog, registry map[string]LogLineRegistryItem) ([]LogLine, error) {
	var out []LogLine
	eventType := *eventLogEntry.EventType

	switch eventType {
	case event.HandlerPipelineQueued:
		var p event.PipelineQueued
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}
		registry[p.PipelineExecutionID] = LogLineRegistryItem{
			Name:    p.Name,
			Started: p.Event.CreatedAt,
		}
	case event.HandlerPipelineStarted:
		var p event.PipelineStarted
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		name := "unknown"
		if entry, exists := registry[p.PipelineExecutionID]; exists {
			name = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
			registry[p.PipelineExecutionID] = LogLineRegistryItem{
				Name:    entry.Name,
				Started: p.Event.CreatedAt,
			}
		}

		out = append(out, LogLine{
			Name:    name,
			Message: fmt.Sprintf("Starting pipeline: %s", p.PipelineExecutionID),
		})
	case event.HandlerPipelineFinished:
		var p event.PipelineFinished
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		name := "unknown"
		started := p.Event.CreatedAt
		if entry, exists := registry[p.PipelineExecutionID]; exists {
			name = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
			started = entry.Started
		}
		duration := p.Event.CreatedAt.Sub(started)

		for k, v := range p.PipelineOutput {
			var value string
			if !isSimpleType(v) {
				jsonData, err := json.Marshal(v)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal pipeline output %s: %v", k, err)
				}
				value = string(jsonData)
			} else {
				value = v.(string)
			}
			out = append(out, LogLine{
				Name:    name,
				Message: fmt.Sprintf("Output: %s = %s", k, value),
			})
		}
		out = append(out, LogLine{
			Name:    name,
			Message: fmt.Sprintf("Complete: %s %s", utils.HumanizeDuration(duration), p.PipelineExecutionID),
		})
	case event.HandlerPipelineFailed:
		var p event.PipelineFailed
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		name := "unknown"
		started := p.Event.CreatedAt
		if entry, exists := registry[p.PipelineExecutionID]; exists {
			name = strings.Split(entry.Name, ".")[len(strings.Split(entry.Name, "."))-1]
			started = entry.Started
		}
		duration := p.Event.CreatedAt.Sub(started)

		output := p.PipelineOutput
		errors := output["errors"].([]interface{})
		out = append(out, LogLine{
			Name:    name,
			Message: fmt.Sprintf("Failed with %d error(s): %s %s", len(errors), utils.HumanizeDuration(duration), p.PipelineExecutionID),
			IsError: true,
		})
	case event.HandlerStepQueued:
		var p event.StepQueued
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}
		registry[p.StepExecutionID] = LogLineRegistryItem{
			Name:    p.StepName,
			Started: p.Event.CreatedAt,
		}

	case event.CommandStepStart:
		var p event.StepStart
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		if p.NextStepAction == "start" { // TODO: Add line for skipped steps?
			pipeline := registry[p.PipelineExecutionID]
			pipelineName := strings.Split(pipeline.Name, ".")[len(strings.Split(pipeline.Name, "."))-1]

			registry[p.StepExecutionID] = LogLineRegistryItem{
				Name:    p.StepName,
				Started: p.Event.CreatedAt,
			}

			stepType := strings.Split(p.StepName, ".")[0]
			stepName := strings.Split(p.StepName, ".")[1]
			stepText := "Starting"

			isForEach := false
			isLoop := false
			if p.StepForEach != nil && p.StepForEach.ForEachStep {
				isForEach = true
			}

			if p.StepLoop != nil {
				isLoop = true
			}

			switch stepType {
			case "echo":
				stepText = fmt.Sprintf("Starting %s: %s", stepType, p.StepInput["text"].(string))
			case "sleep":
				stepText = fmt.Sprintf("Starting %s: %s", stepType, p.StepInput["duration"].(string))
			case "http":
				stepText = fmt.Sprintf("Starting %s: %s %s", stepType, strings.ToUpper(p.StepInput["method"].(string)), p.StepInput["url"].(string))
			}

			startLine := LogLine{
				Name:     pipelineName,
				StepName: &stepName,
				Message:  stepText,
			}

			if isForEach {
				startLine.ForEachKey = &p.StepForEach.Key
			}
			if isLoop {
				startLine.LoopIndex = &p.StepLoop.Index
			}

			// TODO: Add lines for inputs?
			out = append(out, startLine)
		}

	case event.HandlerStepFinished:
		var p event.StepFinished
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		if p.Output != nil && p.Output.Status != "skipped" {
			pipeline := registry[p.PipelineExecutionID]
			pipelineName := strings.Split(pipeline.Name, ".")[len(strings.Split(pipeline.Name, "."))-1]
			step := registry[p.StepExecutionID]
			stepType := strings.Split(step.Name, ".")[0]
			stepName := strings.Split(step.Name, ".")[1]
			duration := p.Event.CreatedAt.Sub(step.Started)

			if p.Output != nil && p.Output.Status == "failed" {
				line := LogLine{
					Name:     pipelineName,
					StepName: &stepName,
					Message:  fmt.Sprintf("Failed: %s %s", utils.HumanizeDuration(duration), p.Output.Errors[0].Error.Detail),
					IsError:  true,
				}
				if p.StepForEach != nil && p.StepForEach.ForEachStep {
					line.ForEachKey = &p.StepForEach.Key
				}
				if p.StepLoop != nil {
					line.LoopIndex = &p.StepLoop.Index
				}
				out = append(out, line)
			} else {
				output := p.Output.Data
				suffix := ""
				switch stepType {
				case "echo":
					suffix = output["text"].(string)
				case "http":
					suffix = fmt.Sprintf("%v", output["status_code"].(float64))
				}

				line := LogLine{
					Name:     pipelineName,
					StepName: &stepName,
					Message:  fmt.Sprintf("Complete: %s %s", utils.HumanizeDuration(duration), suffix),
				}
				if p.StepForEach != nil && p.StepForEach.ForEachStep {
					line.ForEachKey = &p.StepForEach.Key
				}
				if p.StepLoop != nil {
					line.LoopIndex = &p.StepLoop.Index
				}
				out = append(out, line)
			}
		}
	}

	return out, nil
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

type PrintableLogLine struct {
	Items interface{}
}

func (p PrintableLogLine) GetItems() interface{} {
	return p.Items
}

func (PrintableLogLine) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {
	resourceType := r.GetResourceType()
	if resourceType != "ListProcessLogJSONResponse" {
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("invalid resource type: %s", resourceType))
	}

	temp, ok := r.(*flowpipeapiclient.ListProcessLogJSONResponse)
	if !ok {
		return nil, perr.BadRequestWithMessage("unable to cast to flowpipeapiclient.ListProcessLogJSONResponse")
	}

	var logLines []LogLine
	mappingRegistry := make(map[string]LogLineRegistryItem)
	for _, i := range temp.Items {
		parsed, err := LogLinesFromEventLog(i, mappingRegistry)
		if err != nil {
			return nil, perr.BadRequestWithMessage("unable to parse event log entry")
		}
		logLines = append(logLines, parsed...)
	}

	return logLines, nil
}

func (p PrintableLogLine) GetTable() (Table, error) {
	return Table{}, nil
}

func (PrintableLogLine) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{}
}
