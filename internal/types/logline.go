package types

import (
	"encoding/json"
	"fmt"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/pipe-fittings/perr"
	"reflect"
	"strings"
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

func LogLinesFromEventLog(eventLogEntry flowpipeapiclient.ProcessEventLog) ([]LogLine, error) {
	var out []LogLine

	eventType := *eventLogEntry.EventType

	switch eventType {
	case event.HandlerPipelineStarted:
		// TODO: Need `pipeline_name` for display
		var p event.PipelineStarted
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}
		out = append(out, LogLine{
			Name:    "TODO-GET-ME",
			Message: fmt.Sprintf("Starting pipeline: %s", p.PipelineExecutionID),
		})
	case event.HandlerPipelineFinished:
		// TODO: Need `pipeline_name` for display
		var p event.PipelineFinished
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}
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
				Name:    "TODO-GET-ME",
				Message: fmt.Sprintf("Output: %s = %s", k, value),
			})
		}
		out = append(out, LogLine{
			Name:    "TODO-GET-ME",
			Message: fmt.Sprintf("Complete: %s", p.PipelineExecutionID),
		})
	case event.HandlerPipelineFailed:
		// TODO: Need `pipeline_name` for display
		var p event.PipelineFailed
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}
		output := p.PipelineOutput
		errors := output["errors"].([]interface{})
		out = append(out, LogLine{
			Name:    "TODO-GET-ME",
			Message: fmt.Sprintf("Failed with %d error(s): %s", len(errors), p.PipelineExecutionID),
			IsError: true,
		})
	case event.CommandStepStart:
		// TODO: Need `pipeline_name` for display
		var p event.StepStart
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		if p.NextStepAction == "start" { // TODO: Add line for skipped steps?
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
				Name:     "TODO-GET-ME",
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
			// TODO: Check for errors & output
			out = append(out, startLine)
		}

	case event.HandlerStepFinished:
		// TODO: Need `pipeline_name` for display
		// TODO: Need `step_name` for display (could extract from error on failure... but no option on success)
		var p event.StepFinished
		err := json.Unmarshal([]byte(*eventLogEntry.Payload), &p)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s event: %v", p.HandlerName(), err)
		}

		name := "TODO-GET-ME"
		stepName := "TODO_GET_STEP_NAME"

		if p.Output != nil && p.Output.Status == "failed" {
			out = append(out, LogLine{
				Name:     name,
				StepName: &stepName,
				Message:  fmt.Sprintf("Failed: %s", p.Output.Errors[0].Error.Detail),
				IsError:  true,
			})
		} else {
			suffix := ""
			// TODO: Set suffix based on stepType (once we have stepName to obtain the type)
			out = append(out, LogLine{
				Name:     name,
				StepName: &stepName,
				Message:  fmt.Sprintf("Complete: %s", suffix),
				IsError:  true,
			})
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
	for _, i := range temp.Items {
		parsed, err := LogLinesFromEventLog(i)
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
