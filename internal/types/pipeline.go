package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/turbot/go-kit/helpers"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/sanitize"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	localconstants "github.com/turbot/flowpipe/internal/constants"
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
	Name            string                     `json:"name"`
	Description     *string                    `json:"description,omitempty"`
	Mod             string                     `json:"mod"`
	Title           *string                    `json:"title,omitempty"`
	Documentation   *string                    `json:"documentation,omitempty"`
	FileName        string                     `json:"file_name,omitempty"`
	StartLineNumber int                        `json:"start_line_number,omitempty"`
	EndLineNumber   int                        `json:"end_line_number,omitempty"`
	Tags            map[string]string          `json:"tags,omitempty"`
	Steps           []modconfig.PipelineStep   `json:"steps,omitempty"`
	OutputConfig    []modconfig.PipelineOutput `json:"outputs,omitempty"`
	Params          []FpPipelineParam          `json:"params,omitempty"`
}

func (p FpPipeline) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	output := ""
	keyWidth := 9
	if p.Description != nil {
		keyWidth = 13
	}
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	if p.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:").Bold(), *p.Title)
	}
	if p.Description != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:").Bold(), *p.Description)
	}

	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Name:").Bold(), p.Name)

	if len(p.Tags) > 0 {
		output += fmt.Sprintf("%s\n", au.Blue("Tags:").Bold())
		for k, v := range p.Tags {
			output += fmt.Sprintf("- %s %s\n", au.Blue(k+":"), v)
		}
	}

	var pArg string
	if len(p.Params) > 0 {
		output += fmt.Sprintf("%s\n", au.Blue("Params:").Bold())
		for _, param := range p.Params {
			output += param.String(sanitizer, opts)

			// below is to build param string for usage
			if !helpers.IsNil(param.Default) || (param.Optional != nil && *param.Optional) {
				continue
			}
			pArg += " --arg " + param.Name + "=<value>"
		}
	}

	if len(p.OutputConfig) > 0 {
		output += fmt.Sprintf("%s\n", au.Blue("Outputs:").Bold())
		outputKeyWidth := 13
		for _, o := range p.OutputConfig {
			output += fmt.Sprintf("- %-*s%s\n", outputKeyWidth, au.Blue("Name:"), o.Name)
			output += fmt.Sprintf("  %-*s%s\n", outputKeyWidth, au.Blue("Description:"), o.Description)
		}
	}

	output += fmt.Sprintf("%-*sflowpipe pipeline run %s%s\n", keyWidth, au.Blue("Usage:").Bold(), p.Name, pArg)
	return output
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

		var paramDefault any
		if !param.Default.IsNull() {
			paramDefaultGoVal, err := hclhelpers.CtyToGo(param.Default)
			if err != nil {
				return nil, perr.BadRequestWithMessage("unable to convert param default to go value: " + param.Name)
			}
			paramDefault = map[string]any{param.Name: paramDefaultGoVal}
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

	resp.FileName = pipeline.FileName
	resp.StartLineNumber = pipeline.StartLineNumber
	resp.EndLineNumber = pipeline.EndLineNumber

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

	for _, p := range apiResp.Params {
		res.Params = append(res.Params, pipelineParamFromApiResponse(p))
	}

	for _, o := range apiResp.Outputs {
		res.OutputConfig = append(res.OutputConfig, pipelineOutputFromApiResponse(o))
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

func pipelineParamFromApiResponse(paramApiResponse flowpipeapiclient.FpPipelineParam) FpPipelineParam {
	param := FpPipelineParam{
		Name:        *paramApiResponse.Name,
		Description: paramApiResponse.Description,
		Default:     paramApiResponse.Default,
		Optional:    paramApiResponse.Optional,
		Type:        *paramApiResponse.Type,
	}
	return param
}

func pipelineOutputFromApiResponse(outputApiResponse flowpipeapiclient.ModconfigPipelineOutput) modconfig.PipelineOutput {
	output := modconfig.PipelineOutput{
		DependsOn: outputApiResponse.DependsOn,
	}

	if outputApiResponse.Name != nil {
		output.Name = *outputApiResponse.Name
	}

	if outputApiResponse.Description != nil {
		output.Description = *outputApiResponse.Description
	}

	return output
}

type FpPipelineParam struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Optional    *bool   `json:"optional,omitempty"`
	Default     any     `json:"default,omitempty"`
	Type        string  `json:"type"`
}

func (p FpPipelineParam) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	keyWidth := 10
	if p.Description != nil && len(*p.Description) > 0 {
		keyWidth = 13
	}
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	o := ""
	if p.Optional != nil && *p.Optional {
		o = au.Sprintf(" %s", au.Yellow("Optional").Italic())
	}

	output := fmt.Sprintf("- %-*s%s%s\n", keyWidth, au.Blue("Name:"), p.Name, o)
	output += fmt.Sprintf("  %-*s%s\n", keyWidth, au.Blue("Type:"), p.Type)
	if p.Description != nil && len(*p.Description) > 0 {
		output += fmt.Sprintf("  %-*s%s\n", keyWidth, au.Blue("Description:"), *p.Description)
	}
	if defaults, hasDefaults := p.Default.(map[string]any); hasDefaults {
		if v, ok := defaults[p.Name]; ok {
			var valueString string
			if isSimpleType(v) {
				valueString = formatSimpleValue(v, au)
			} else {
				s, err := opts.JsonFormatter.Marshal(v)
				if err != nil {
					valueString = au.Sprintf(au.Red("error parsing value"))
				} else {
					valueString = string(s)
				}
			}
			output += fmt.Sprintf("  %-*s%s\n", keyWidth, au.Blue("Default:"), valueString)
		} else {
			output += fmt.Sprintf("  %-*s%s\n", keyWidth, au.Blue("Default:"), "")
		}
	} else {
		output += fmt.Sprintf("  %-*s%s\n", keyWidth, au.Blue("Default:"), "")
	}

	if strings.HasSuffix(output, "\n\n") {
		output = strings.TrimSuffix(output, "\n")
	}
	return output
}

type PipelineExecutionResponse map[string]interface{}

type CmdPipeline struct {
	Command       string                 `json:"command" binding:"required,oneof=run"`
	Args          map[string]interface{} `json:"args,omitempty"`
	ArgsString    map[string]string      `json:"args_string,omitempty"`
	ExecutionMode *string                `json:"execution_mode,omitempty" binding:"omitempty,oneof=synchronous asynchronous"`
	WaitRetry     *int                   `json:"wait_retry,omitempty" binding:"omitempty"`
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
	Items []FpPipeline
}

func NewPrintablePipeline(resp *ListPipelineResponse) *PrintablePipeline {
	return &PrintablePipeline{
		Items: resp.Items,
	}
}

func NewPrintablePipelineFromSingle(input *FpPipeline) *PrintablePipeline {
	return &PrintablePipeline{
		Items: []FpPipeline{*input},
	}
}

func (p PrintablePipeline) GetItems() []FpPipeline {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	var tableRows []TableRow
	for _, item := range p.Items {
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

	return NewTable(tableRows, p.getColumns()), nil
}

func (PrintablePipeline) getColumns() (columns []TableColumnDefinition) {
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

type FpPipelineExecution struct {
	ExecutionId         string                `json:"execution_id"`
	PipelineExecutionId string                `json:"pipeline_execution_id"`
	Status              string                `json:"status"`
	PipelineName        *string               `json:"pipeline,omitempty"`
	CreatedAt           *time.Time            `json:"created_at,omitempty"`
	Outputs             map[string]any        `json:"outputs,omitempty"`
	Errors              []modconfig.StepError `json:"errors,omitempty"`
}

func (p FpPipelineExecution) String(sanitizer *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	out := fmt.Sprintf("%s%s\n", au.Blue("Execution ID: ").Bold(), p.ExecutionId)
	if !helpers.IsNil(p.PipelineName) {
		out += fmt.Sprintf("%s%s\n", au.Blue("Pipeline:     ").Bold(), *p.PipelineName)
	}
	out += fmt.Sprintf("%s%s\n", au.Blue("Pipeline ID:  ").Bold(), p.PipelineExecutionId)
	out += fmt.Sprintf("%s%s\n", au.Blue("Status:       ").Bold(), p.Status)
	if !helpers.IsNil(p.CreatedAt) {
		out += fmt.Sprintf("%s%s\n", au.Blue("Created At:   ").Bold(), p.CreatedAt.Format(time.RFC3339))
	}

	if !helpers.IsNil(p.Outputs) && len(p.Outputs) > 0 {
		out += fmt.Sprintf("%s\n", au.Blue("Outputs:").Bold())
		out += sortAndParseMap(p.Outputs, "", " ", au, opts)
	}

	if !helpers.IsNil(p.Errors) && len(p.Errors) > 0 {
		out += fmt.Sprintf("%s\n", au.Blue("Errors:").Bold())
		for _, e := range p.Errors {
			out += fmt.Sprintf("  %s %s\n", au.Red(e.Error.Title+":"), au.Red(e.Error.Detail))
		}
	}

	return out
}

func FpPipelineExecutionFromAPIResponse(apiResp map[string]any) (*FpPipelineExecution, error) {
	if apiResp != nil && apiResp["flowpipe"] != nil {
		contents := apiResp["flowpipe"].(map[string]any)

		executionId, _ := contents["execution_id"].(string)
		pipelineId, _ := contents["pipeline_execution_id"].(string)
		pipeline, hasName := contents["pipeline"].(string)

		exec := FpPipelineExecution{
			ExecutionId:         executionId,
			PipelineExecutionId: pipelineId,
			Status:              "queued",
		}

		if hasName {
			exec.PipelineName = &pipeline
		}

		return &exec, nil
	}

	return nil, perr.Internal(fmt.Errorf("unexpected API response"))
}

type PrintablePipelineExecution struct {
	Items []FpPipelineExecution
}

func (p PrintablePipelineExecution) GetItems() []FpPipelineExecution {
	return p.Items
}

func (p PrintablePipelineExecution) GetTable() (Table, error) {
	var tableRows []TableRow
	for _, item := range p.Items {
		cells := []any{
			item.ExecutionId,
			item.PipelineName,
			item.Status,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}
	return NewTable(tableRows, p.getColumns()), nil
}

func (PrintablePipelineExecution) getColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "EXECUTION ID",
			Type:        "string",
			Description: "Execution ID",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "Pipeline name",
		},
		{
			Name:        "STATUS",
			Type:        "string",
			Description: "Pipeline execution status",
		},
	}
}
