package types

import (
	"fmt"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
	"time"

	"github.com/logrusorgru/aurora"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	typehelpers "github.com/turbot/go-kit/types"
)

// The definition of a single Flowpipe Process
type Process struct {
	ID        string    `json:"execution_id"`
	Pipeline  string    `json:"pipeline"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (p Process) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	keyWidth := 14
	output := ""
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Execution ID:"), p.ID)
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Pipeline:"), p.Pipeline)
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Status:"), p.Status)
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Created:"), p.CreatedAt.Local().Format(time.DateTime))
	return output
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
	ID     string                 `json:"process_id"`
	Output map[string]interface{} `json:"output"`
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
	Items []Process
}

func NewPrintableProcess(resp *ListProcessResponse) *PrintableProcess {
	return &PrintableProcess{
		Items: resp.Items,
	}
}

func NewPrintableProcessFromSingle(input *Process) *PrintableProcess {
	return &PrintableProcess{
		Items: []Process{*input},
	}
}

func (p PrintableProcess) GetItems() []Process {
	return p.Items
}

func (p PrintableProcess) GetTable() (printers.Table, error) {
	var tableRows []printers.TableRow
	for _, item := range p.Items {
		cells := []any{
			item.ID,
			item.Pipeline,
			item.CreatedAt.Local().Format(time.DateTime),
			item.Status,
		}
		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable(tableRows, p.getColumns()), nil
}

func (PrintableProcess) getColumns() (columns []printers.TableColumnDefinition) {
	return []printers.TableColumnDefinition{
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

func ListProcessResponseFromAPIResponse(apiResp *flowpipeapiclient.ListProcessResponse) (*ListProcessResponse, error) {
	if apiResp == nil {
		return nil, nil
	}

	var res = &ListProcessResponse{
		Items:     make([]Process, len(apiResp.Items)),
		NextToken: apiResp.NextToken,
	}

	for i, apiItem := range apiResp.Items {
		item, err := ProcessFromAPIResponse(apiItem)
		if err != nil {
			return nil, err
		}
		res.Items[i] = *item
	}
	return res, nil
}

func ProcessFromAPIResponse(apiResp flowpipeapiclient.Process) (*Process, error) {
	createdAt, _ := time.Parse(time.RFC3339Nano, *apiResp.CreatedAt)
	res := &Process{
		ID:        typehelpers.SafeString(apiResp.ExecutionId),
		Pipeline:  typehelpers.SafeString(apiResp.Pipeline),
		Status:    typehelpers.SafeString(apiResp.Status),
		CreatedAt: createdAt,
	}

	return res, nil
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
