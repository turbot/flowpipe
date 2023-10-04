package types

import (
	"fmt"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type FpPipeline struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Mod         string  `json:"mod"`
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []FpPipeline `json:"items"`
	NextToken *string      `json:"next_token,omitempty"`
}

type RunPipelineResponse struct {
	ExecutionID           string `json:"execution_id"`
	PipelineExecutionID   string `json:"pipeline_execution_id"`
	ParentStepExecutionID string `json:"parent_step_execution_id"`
}

type CmdPipeline struct {
	Command       string                 `json:"command" binding:"required,oneof=run"`
	Args          map[string]interface{} `json:"args,omitempty"`
	ArgsString    map[string]string      `json:"args_string,omitempty"`
	ExecutionMode *string                `json:"execution_mode,omitempty" binding:"omitempty,oneof=synchronous asynchronous"`
}

type PrintablePipeline struct {
	Items interface{}
}

func (PrintablePipeline) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListPipelineResponse" {

		return nil, perr.BadRequestWithMessage(fmt.Sprintf("invalid resource type: %s", apiResourceType))
	}

	lp, ok := r.(*flowpipeapiclient.ListPipelineResponse)
	if !ok {

		return nil, perr.BadRequestWithMessage("unable to cast to flowpipeapiclient.ListPipelineResponse")
	}

	return lp.Items, nil
}

func (p PrintablePipeline) GetItems() interface{} {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.FpPipeline)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Pipeline")
	}

	var tableRows []TableRow
	for _, item := range lp {

		description := ""
		if item.Description != nil {
			description = *item.Description
		}
		cells := []interface{}{
			*item.Mod,
			*item.Name,
			description,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return Table{
		Rows:    tableRows,
		Columns: p.GetColumns(),
	}, nil
}

func (PrintablePipeline) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "MOD",
			Type:        "string",
			Description: "Mod name",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "Pipeline name",
		},
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Pipeline description",
		},
	}
}
