package types

import (
	"github.com/turbot/flowpipe/internal/sanitize"
	"time"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/pipe-fittings/perr"
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
	Items any
}

func (PrintableProcess) Transform(r flowpipeapiclient.FlowpipeAPIResource) (any, error) {

	// apiResourceType := r.GetResourceType()

	// if apiResourceType != "ListProcessResponse" {
	// 	return nil, perr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
	// }

	lp, ok := r.(*flowpipeapiclient.ListProcessResponse)
	if !ok {
		return nil, perr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListProcessResponse")
	}

	return lp.Items, nil
}

func (p PrintableProcess) GetItems(sanitizer *sanitize.Sanitizer) any {
	items, ok := p.Items.([]Process)
	if !ok {
		// not expected
		return []any{}
	}

	sanitizedItems := make([]any, len(items))
	for i, item := range items {
		sanitizedItems[i] = sanitizer.SanitizeStruct(item)
	}
	return p.Items
}

func (p PrintableProcess) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.Process)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Process")
	}

	var tableRows []TableRow
	for _, item := range lp {
		cells := []any{
			*item.ExecutionId,
			*item.Pipeline,
			*item.CreatedAt,
			*item.Status,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return Table{
		Rows:    tableRows,
		Columns: p.GetColumns(),
	}, nil
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
