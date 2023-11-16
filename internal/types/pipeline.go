package types

import (
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type ListPipelineResponseItem struct {
	Name          string            `json:"name"`
	Description   *string           `json:"description,omitempty"`
	Mod           string            `json:"mod"`
	Title         *string           `json:"title,omitempty"`
	Documentation *string           `json:"documentation,omitempty"`
	Tags          map[string]string `json:"tags"`
}

func ListPipelineResponseItemFromAPI(apiItem flowpipeapiclient.ListPipelineResponseItem) ListPipelineResponseItem {
	res := ListPipelineResponseItem{
		Name:          typehelpers.SafeString(apiItem.Name),
		Description:   apiItem.Description,
		Mod:           typehelpers.SafeString(apiItem.Mod),
		Title:         apiItem.Title,
		Documentation: apiItem.Documentation,
		Tags:          make(map[string]string),
	}
	if apiItem.Tags != nil {
		res.Tags = *apiItem.Tags
	}
	return res
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []ListPipelineResponseItem `json:"items"`
	NextToken *string                    `json:"next_token,omitempty"`
}

func ListPipelineResponseFromAPI(apiResp *flowpipeapiclient.ListPipelineResponse) *ListPipelineResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListPipelineResponse{
		Items:     make([]ListPipelineResponseItem, len(apiResp.Items)),
		NextToken: apiResp.NextToken,
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = ListPipelineResponseItemFromAPI(apiItem)
	}
	return res
}

func (o ListPipelineResponse) GetResourceType() string {
	return "ListPipelineResponse"
}

type GetPipelineResponse struct {
	Name          string                     `json:"name"`
	Description   *string                    `json:"description,omitempty"`
	Mod           string                     `json:"mod"`
	Title         *string                    `json:"title,omitempty"`
	Documentation *string                    `json:"documentation,omitempty"`
	Tags          map[string]string          `json:"tags"`
	Steps         []modconfig.PipelineStep   `json:"steps,omitempty"`
	OutputConfig  []modconfig.PipelineOutput `json:"outputs,omitempty"`
	Params        []FpPipelineParam          `json:"params,omitempty"`
}

func GetPipelineResponseFromAPI(apiResp *flowpipeapiclient.GetPipelineResponse) *GetPipelineResponse {
	if apiResp == nil {
		return nil
	}

	res := &GetPipelineResponse{
		Name:          typehelpers.SafeString(apiResp.Name),
		Description:   apiResp.Description,
		Mod:           typehelpers.SafeString(apiResp.Mod),
		Title:         apiResp.Title,
		Documentation: apiResp.Documentation,
		Tags:          make(map[string]string),

		Steps:        make([]modconfig.PipelineStep, 0, len(apiResp.Steps)),
		Params:       make([]FpPipelineParam, 0, len(apiResp.Params)),
		OutputConfig: make([]modconfig.PipelineOutput, 0, len(apiResp.Outputs)),
	}

	// TODO KAI >???????
	//for _, s := range apiResp.Steps {
	//	res.Steps = append(res.Steps)
	//}
	if apiResp.Tags != nil {
		res.Tags = *apiResp.Tags
	}
	return res
}

type FpPipelineParam struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Optional    *bool       `json:"optional,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Type        string      `json:"type"`
}

type PipelineExecutionResponse map[string]interface{}

type CmdPipeline struct {
	Command       string                 `json:"command" binding:"required,oneof=run"`
	Args          map[string]interface{} `json:"args,omitempty"`
	ArgsString    map[string]string      `json:"args_string,omitempty"`
	ExecutionMode *string                `json:"execution_mode,omitempty" binding:"omitempty,oneof=synchronous asynchronous"`
	WaitRetry     *int                   `json:"wait_retry,omitempty" binding:"omitempty"`
}

type PrintablePipeline struct {
	Items interface{}
}

func (PrintablePipeline) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {
	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListPipelineResponse" {

		return nil, perr.BadRequestWithMessage(fmt.Sprintf("invalid resource type: %s", apiResourceType))
	}

	lp, ok := r.(*ListPipelineResponse)
	if !ok {
		return nil, perr.BadRequestWithMessage("unable to cast to ListPipelineResponse")
	}

	return lp.Items, nil
}

func (p PrintablePipeline) GetItems() interface{} {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]ListPipelineResponseItem)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("Unable to cast to []ListPipelineResponseItem")
	}

	var tableRows []TableRow
	for _, item := range lp {
		var description string
		if item.Description != nil {
			description = *item.Description
		}

		cells := []interface{}{
			item.Mod,
			item.Name,
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
