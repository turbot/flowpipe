package types

import (
	"encoding/json"
	"fmt"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type ListPipelineResponseItem struct {
	Name          string            `json:"name"`
	Description   *string           `json:"description,omitempty"`
	Mod           string            `json:"mod"`
	Title         *string           `json:"title,omitempty"`
	Documentation *string           `json:"documentation,omitempty"`
	Tags          map[string]string `json:"tags"`
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []ListPipelineResponseItem `json:"items"`
	NextToken *string                    `json:"next_token,omitempty"`
}

type GetPipelineResponse struct {
	Name          string                              `json:"name"`
	Description   *string                             `json:"description,omitempty"`
	Mod           string                              `json:"mod"`
	Title         *string                             `json:"title,omitempty"`
	Documentation *string                             `json:"documentation,omitempty"`
	Tags          map[string]string                   `json:"tags"`
	Steps         []modconfig.IPipelineStep           `json:"steps,omitempty"`
	OutputConfig  []modconfig.PipelineOutput          `json:"outputs,omitempty"`
	Params        map[string]*modconfig.PipelineParam `json:"params,omitempty"`
}

type PipelineExecutionResponse map[string]interface{}

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

	// else if apiResourceType == "GetPipelineResponse" {
	// 	lp, ok := r.(*flowpipeapiclient.GetPipelineResponse)
	// 	if !ok {
	// 		return nil, perr.BadRequestWithMessage("unable to cast to flowpipeapiclient.FpPipeline")
	// 	}
	// 	return []flowpipeapiclient.GetPipelineResponse{*lp}, nil
	// } else {
	// 	return nil, perr.BadRequestWithMessage(fmt.Sprintf("invalid resource type: %s", apiResourceType))
	// }
}

func (p PrintablePipeline) GetItems() interface{} {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.ListPipelineResponseItem)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Pipeline")
	}

	var tableRows []TableRow
	for _, item := range lp {
		var description, documentation, title, tags string
		if item.Description != nil {
			description = *item.Description
		}
		if item.Documentation != nil {
			documentation = *item.Documentation
		}
		if item.Title != nil {
			title = *item.Title
		}
		if item.Tags != nil {
			data, _ := json.Marshal(*item.Tags)
			tags = string(data)
		}

		cells := []interface{}{
			*item.Mod,
			*item.Name,
			title,
			description,
			documentation,
			tags,
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
			Name:        "TITLE",
			Type:        "string",
			Description: "Pipeline title",
		},
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Pipeline description",
		},
		{
			Name:        "DOCUMENTATION",
			Type:        "string",
			Description: "Pipeline documentation",
		},
		{
			Name:        "TAGS",
			Type:        "string",
			Description: "Pipeline tags",
		},
	}
}
