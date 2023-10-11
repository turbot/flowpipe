package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

// The definition of a single Flowpipe Process
type Process struct {
	ID       string `json:"execution_id"`
	Pipeline string `json:"pipeline"`
	Status   string `json:"status"`
}

// Process log payload definition
type ProcessPayload struct {
	PipelineName        string `json:"name"`
	PipelineExecutionID string `json:"pipeline_execution_id"`
}

type ProcessOutputData struct {
	ID     string                 `json:"process_id"`
	Output map[string]interface{} `json:"output"`
}

type PrintableProcess struct {
	Items interface{}
}

func (PrintableProcess) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	return nil, nil
	// apiResourceType := r.GetResourceType()
	// if apiResourceType != "ListProcessResponse" {
	// 	return nil, fperr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
	// }

	// lp, ok := r.(*flowpipeapiclient.ListProcessResponse)
	// if !ok {
	// 	return nil, fperr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListProcessResponse")
	// }

	// return lp.Items, nil
}

func (p PrintableProcess) GetItems() interface{} {
	return p.Items
}

func (p PrintableProcess) GetTable() (Table, error) {
	return Table{}, nil
	// lp, ok := p.Items.([]flowpipeapiclient.Process)

	// if !ok {
	// 	return Table{}, fperr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Process")
	// }

	// var tableRows []TableRow
	// for _, item := range lp {
	// 	cells := []interface{}{
	// 		*item.Type,
	// 		*item.Name,
	// 		*item.Parallel,
	// 	}
	// 	tableRows = append(tableRows, TableRow{Cells: cells})
	// }

	// return Table{
	// 	Rows:    tableRows,
	// 	Columns: p.GetColumns(),
	// }, nil
}

func (PrintableProcess) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "ID",
			Type:        "string",
			Description: "The id of the process",
		},
	}
}

// This type is used by the API to return a list of processs.
type ListProcessResponse struct {
	Items     []Process `json:"items"`
	NextToken *string   `json:"next_token,omitempty"`
}

// This type is used by the API to return a list of pipelines.
type ListProcessLogResponse struct {
	Items     []EventLogEntry `json:"items"`
	NextToken *string         `json:"next_token,omitempty"`
}

type CmdProcess struct {
	Command             string `json:"command" binding:"required,oneof=run cancel pause resume"`
	PipelineExecutionID string `json:"pipeline_execution_id,omitempty" format:"^(pexec|exec)_[0-9a-v]{20}$"`
	Reason              string `json:"reason,omitempty"`
}
