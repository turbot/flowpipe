package types

import (
	"time"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

// The definition of a single Flowpipe Process
type Process struct {
	ID        string    `json:"execution_id"`
	Pipeline  string    `json:"pipeline"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Process log payload definition
type ProcessPayload struct {
	PipelineName        string              `json:"name"`
	PipelineExecutionID string              `json:"pipeline_execution_id"`
	Event               ProcessPayloadEvent `json:"event"`
}

type ProcessPayloadEvent struct {
	CreatedAt time.Time `json:"created_at"`
}

type ProcessOutputData struct {
	ID     string         `json:"process_id"`
	Output map[string]any `json:"output"`
}

// Identical to the EventLogEntry struct in internal/types/execution.go
// Using the EventLogEntry returned an error at the time of openapi generation:
// cannot find type definition: json.RawMessage
// TODO - Recheck to use the EventLogEntry struct
type ProcessEventLog struct {
	EventType string     `json:"event_type"`
	Timestamp *time.Time `json:"ts"`
	// Setting the type as string for now, as the CLI need to print the payload
	Payload string `json:"payload"`
}

type PrintableProcess struct {
	// todo should we map to internal types
	Items []flowpipeapiclient.Process
}

func NewPrintableProcess(resp *flowpipeapiclient.ListProcessResponse) *PrintableProcess {
	return &PrintableProcess{
		Items: resp.Items,
	}
}

func (p PrintableProcess) GetItems() []flowpipeapiclient.Process {
	return p.Items
}

func (p PrintableProcess) GetTable() (Table, error) {
	var tableRows []TableRow
	for _, item := range p.Items {
		cells := []any{
			*item.ExecutionId,
			*item.Pipeline,
			*item.CreatedAt,
			*item.Status,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return NewTable(tableRows, p.GetColumns()), nil
}

func (PrintableProcess) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "EXECUTION_ID",
			Type:        "string",
			Description: "The id of the process execution",
		},
		{
			Name:        "PIPELINE",
			Type:        "string",
			Description: "The name of the pipeline",
		},
		{
			Name:        "CREATED_AT",
			Type:        "string",
			Description: "The name of the pipeline",
		},
		{
			Name:        "STATUS",
			Type:        "string",
			Description: "The status of the process execution",
		},
	}
}

// This type is used by the API to return a list of processs.
type ListProcessResponse struct {
	Items     []Process `json:"items"`
	NextToken *string   `json:"next_token,omitempty"`
}

type ListProcessLogJSONResponse struct {
	Items     []ProcessEventLog `json:"items"`
	NextToken *string           `json:"next_token,omitempty"`
}

// This type is used by the API to return a list of proces logs.
type ListProcessLogResponse struct {
	Items     []EventLogEntry `json:"items"`
	NextToken *string         `json:"next_token,omitempty"`
}

type CmdProcess struct {
	Command             string `json:"command" binding:"required,oneof=run cancel pause resume"`
	PipelineExecutionID string `json:"pipeline_execution_id,omitempty" format:"^(pexec|exec)_[0-9a-v]{20}$"`
	Reason              string `json:"reason,omitempty"`
}
