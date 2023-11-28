package types

import (
	"encoding/json"
	"fmt"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/sanitize"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []FpPipeline `json:"items"`
	NextToken *string      `json:"next_token,omitempty"`
}

func ListPipelineResponseFromAPIResponse(apiResp *flowpipeapiclient.ListPipelineResponse) (*ListPipelineResponse, error) {
	if apiResp == nil {
		return nil, nil
	}

	var res = &ListPipelineResponse{
		Items:     make([]FpPipeline, len(apiResp.Items)),
		NextToken: apiResp.NextToken,
	}

	for i, apiItem := range apiResp.Items {
		item, err := FpPipelineFromAPIResponse(apiItem)
		if err != nil {
			return nil, err
		}
		res.Items[i] = *item
	}
	return res, nil
}

func (o ListPipelineResponse) GetResourceType() string {
	return "ListPipelineResponse"
}

type FpPipeline struct {
	Name          string                     `json:"name"`
	Description   *string                    `json:"description,omitempty"`
	Mod           string                     `json:"mod"`
	Title         *string                    `json:"title,omitempty"`
	Documentation *string                    `json:"documentation,omitempty"`
	Tags          map[string]string          `json:"tags,omitempty"`
	Steps         []modconfig.PipelineStep   `json:"steps,omitempty"`
	OutputConfig  []modconfig.PipelineOutput `json:"outputs,omitempty"`
	Params        []FpPipelineParam          `json:"params,omitempty"`
}

func FpPipelineFromModPipeline(pipeline *modconfig.Pipeline) (*FpPipeline, error) {
	resp := &FpPipeline{
		Name:          pipeline.Name(),
		Description:   pipeline.Description,
		Mod:           pipeline.GetMod().FullName,
		Title:         pipeline.Title,
		Tags:          pipeline.Tags,
		Documentation: pipeline.Documentation,
		Steps:         pipeline.Steps,
		OutputConfig:  pipeline.OutputConfig,
	}

	var pipelineParams []FpPipelineParam
	for _, param := range pipeline.Params {

		paramDefault := map[string]any{}
		if !param.Default.IsNull() {
			paramDefaultGoVal, err := hclhelpers.CtyToGo(param.Default)
			if err != nil {
				return nil, perr.BadRequestWithMessage("unable to convert param default to go value: " + param.Name)
			}
			paramDefault[param.Name] = paramDefaultGoVal
		}

		pipelineParams = append(pipelineParams, FpPipelineParam{
			Name:        param.Name,
			Description: utils.ToStringPointer(param.Description),
			Optional:    &param.Optional,
			Type:        param.Type.FriendlyName(),
			Default:     paramDefault,
		})

		resp.Params = pipelineParams
	}
	return resp, nil
}

func FpPipelineFromAPIResponse(apiResp flowpipeapiclient.FpPipeline) (*FpPipeline, error) {
	res := &FpPipeline{
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

	for _, s := range apiResp.Steps {
		step, err := pipelineStepFromApiResponse(s)
		if err != nil {
			return nil, err
		}
		res.Steps = append(res.Steps, step)
	}
	if apiResp.Tags != nil {
		res.Tags = *apiResp.Tags
	}
	return res, nil
}

// pipelineStepFromApiResponse converts the API response steps to the internal representation.
func pipelineStepFromApiResponse(apiStep map[string]any) (modconfig.PipelineStep, error) {
	stepType := apiStep["step_type"].(string)
	var step modconfig.PipelineStep
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		step = &modconfig.PipelineStepHttp{}
	case schema.BlockTypePipelineStepSleep:
		step = &modconfig.PipelineStepSleep{}
	case schema.BlockTypePipelineStepEmail:
		step = &modconfig.PipelineStepEmail{}
	case schema.BlockTypePipelineStepTransform:
		step = &modconfig.PipelineStepTransform{}
	case schema.BlockTypePipelineStepQuery:
		step = &modconfig.PipelineStepQuery{}
	case schema.BlockTypePipelineStepPipeline:
		step = &modconfig.PipelineStepPipeline{}
	case schema.BlockTypePipelineStepFunction:
		step = &modconfig.PipelineStepFunction{}
	case schema.BlockTypePipelineStepContainer:
		step = &modconfig.PipelineStepContainer{}
	case schema.BlockTypePipelineStepInput:
		step = &modconfig.PipelineStepInput{}
	default:
		// Handle unknown step type
		return nil, perr.BadRequestWithMessage(fmt.Sprintf("unknown step type: %s", stepType))
	}
	jsonBytes, err := json.Marshal(apiStep)
	if err != nil {
		return nil, perr.Internal(err)
	}
	err = json.Unmarshal(jsonBytes, step)
	if err != nil {
		return nil, perr.Internal(err)
	}

	return step, nil
}

type FpPipelineParam struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Optional    *bool   `json:"optional,omitempty"`
	Default     any     `json:"default,omitempty"`
	Type        string  `json:"type"`
}

type PipelineExecutionResponse map[string]any

type CmdPipeline struct {
	Command       string            `json:"command" binding:"required,oneof=run"`
	Args          map[string]any    `json:"args,omitempty"`
	ArgsString    map[string]string `json:"args_string,omitempty"`
	ExecutionMode *string           `json:"execution_mode,omitempty" binding:"omitempty,oneof=synchronous asynchronous"`
	WaitRetry     *int              `json:"wait_retry,omitempty" binding:"omitempty"`
}

func (c *CmdPipeline) GetExecutionMode() string {
	executionMode := localconstants.DefaultExecutionMode
	if c.ExecutionMode != nil {
		executionMode = *c.ExecutionMode
	}
	return executionMode
}

func (c *CmdPipeline) GetWaitRetry() int {
	if c.WaitRetry != nil {
		return *c.WaitRetry
	}
	return localconstants.DefaultWaitRetry
}

type PrintablePipeline struct {
	Items any
}

func (PrintablePipeline) Transform(r flowpipeapiclient.FlowpipeAPIResource) (any, error) {
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

func (p PrintablePipeline) GetItems(sanitizer *sanitize.Sanitizer) any {
	items, ok := p.Items.([]FpPipeline)
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

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]FpPipeline)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("Unable to cast to []ListPipelineResponseItem")
	}

	var tableRows []TableRow
	for _, item := range lp {
		var description string
		if item.Description != nil {
			description = *item.Description
		}

		cells := []any{
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
