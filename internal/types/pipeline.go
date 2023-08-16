package types

import (
	"fmt"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser/pcerr"

	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []pipeline.Pipeline `json:"items"`
	NextToken *string             `json:"next_token,omitempty"`
}

type RunPipelineResponse struct {
	ExecutionID           string `json:"execution_id"`
	PipelineExecutionID   string `json:"pipeline_execution_id"`
	ParentStepExecutionID string `json:"parent_step_execution_id"`
}

type CmdPipeline struct {
	Command string                 `json:"command" binding:"required,oneof=run"`
	Args    map[string]interface{} `json:"args,omitempty"`
}

type PrintablePipeline struct {
	Items interface{}
}

func (PrintablePipeline) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListPipelineResponse" {
		// return nil, fperr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
		return nil, fmt.Errorf("invalid resource type: %s", apiResourceType)
	}

	lp, ok := r.(*flowpipeapiclient.ListPipelineResponse)
	if !ok {
		// return nil, fperr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListPipelineResponse")
		return nil, fmt.Errorf("unable to cast to flowpipeapiclient.ListPipelineResponse")
	}

	return lp.Items, nil
}

func (p PrintablePipeline) GetItems() interface{} {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.Pipeline)

	if !ok {
		return Table{}, pcerr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Pipeline")
	}

	var tableRows []TableRow
	for _, item := range lp {

		description := ""
		if item.Description != nil {
			description = *item.Description
		}
		cells := []interface{}{
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

type PrintableTrigger struct {
	Items interface{}
}

func (PrintableTrigger) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	return nil, nil
}

func (p PrintableTrigger) GetItems() interface{} {
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	return Table{}, nil
}

func (PrintableTrigger) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the trigger",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the trigger",
		},
	}
}

// This type is used by the API to return a list of triggers.
type ListTriggerResponse struct {
	Items     []pipeline.Trigger `json:"items"`
	NextToken *string            `json:"next_token,omitempty"`
}
