package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/logrusorgru/aurora"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/go-kit/helpers"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"golang.org/x/exp/maps"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

type FpTrigger struct {
	Name            string              `json:"name"`
	Mod             string              `json:"mod"`
	Type            string              `json:"type"`
	Enabled         bool                `json:"enabled"`
	Description     *string             `json:"description,omitempty"`
	Pipelines       []FpTriggerPipeline `json:"pipelines,omitempty"`
	Url             *string             `json:"url,omitempty"`
	Title           *string             `json:"title,omitempty"`
	FileName        string              `json:"file_name,omitempty"`
	StartLineNumber int                 `json:"start_line_number,omitempty"`
	EndLineNumber   int                 `json:"end_line_number,omitempty"`
	Documentation   *string             `json:"documentation,omitempty"`
	Tags            map[string]string   `json:"tags,omitempty"`
	Schedule        *string             `json:"schedule,omitempty"`
	Query           *string             `json:"query,omitempty"`
	RootMod         string              `json:"root_mod"`
	Params          []FpPipelineParam   `json:"params,omitempty"`
}

type FpTriggerPipeline struct {
	CaptureGroup string `json:"capture_group"`
	Pipeline     string `json:"pipeline"`
}

func (t FpTrigger) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var output string
	var statusText string
	left := au.BrightBlack("[")
	right := au.BrightBlack("]")
	keyWidth := 10
	if t.Description != nil {
		keyWidth = 13
	}

	if !t.Enabled {
		statusText = fmt.Sprintf("%s%s%s", left, au.Red("disabled"), right)
	}
	output += fmt.Sprintf("%-*s%s %s\n", keyWidth, au.Blue("Name:"), t.getTypeAndName(), statusText)
	if t.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:"), *t.Title)
	}
	if t.Description != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:"), *t.Description)
	}
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Type:"), t.Type)

	switch t.Type {
	case schema.TriggerTypeHttp:
		if t.Url != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("URL:"), *t.Url)
		}
		output += fmt.Sprintf("%s\n", au.Blue("Pipeline:"))
		for _, pipeline := range t.Pipelines {
			output += fmt.Sprintf("  %s %s\n", au.Blue(utils.ToTitleCase(pipeline.CaptureGroup)+":"), t.getPipelineDisplay(pipeline.Pipeline))
		}
		// TODO: Add usage section
	case schema.TriggerTypeQuery:
		if t.Schedule != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Schedule:"), *t.Schedule)
		}
		if t.Query != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Query:"), *t.Query)
		}
		output += fmt.Sprintf("%s\n", au.Blue("Pipeline:"))
		for _, pipeline := range t.Pipelines {
			output += fmt.Sprintf("  %s %s\n", au.Blue(utils.ToTitleCase(pipeline.CaptureGroup)+":"), t.getPipelineDisplay(pipeline.Pipeline))
		}
	case schema.TriggerTypeSchedule:
		if t.Schedule != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Schedule:"), *t.Schedule)
		}
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Pipeline:"), t.getPipelineDisplay(t.Pipelines[0].Pipeline))
	}

	if len(t.Tags) > 0 {
		output += fmt.Sprintf("%s\n", au.Blue("Tags:"))
		for k, v := range t.Tags {
			output += fmt.Sprintf("  %s %s\n", au.Cyan(k+":"), v)
		}
	}

	if strings.HasSuffix(output, "\n\n") {
		output = strings.TrimSuffix(output, "\n")
	}
	return output
}

func (t FpTrigger) getTypeAndName() string {
	shortName := strings.Split(t.Name, ".")[len(strings.Split(t.Name, "."))-1]
	return fmt.Sprintf("%s.%s", t.Type, shortName)
}

func (t FpTrigger) TriggerDisplayName() string {
	if t.Mod == t.RootMod {
		// Find the first dot

		// Split the string into components and keep only the last two parts
		parts := strings.SplitN(t.Name, ".", 3)

		// Rejoin the last part
		result := parts[2]
		return result
	}
	return t.Name
}

func (t FpTrigger) getPipelineDisplay(pipeline string) string {
	rootMod := strings.Split(t.Name, ".")[0]
	if strings.Split(pipeline, ".")[0] == rootMod {
		return strings.Split(pipeline, ".")[len(strings.Split(pipeline, "."))-1]
	}
	return pipeline
}

// This type is used by the API to return a list of triggers.
type ListTriggerResponse struct {
	Items     []FpTrigger `json:"items"`
	NextToken *string     `json:"next_token,omitempty"`
}

func (o ListTriggerResponse) GetResourceType() string {
	return "ListTriggerResponse"
}

func ListTriggerResponseFromAPI(apiResp *flowpipeapiclient.ListTriggerResponse) *ListTriggerResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListTriggerResponse{
		NextToken: apiResp.NextToken,
		Items:     make([]FpTrigger, len(apiResp.Items)),
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = FpTriggerFromAPI(apiItem)
	}
	return res
}

func FpTriggerFromAPI(apiTrigger flowpipeapiclient.FpTrigger) FpTrigger {
	var pls []FpTriggerPipeline
	for _, pl := range apiTrigger.Pipelines {
		pls = append(pls, FpTriggerPipeline{
			CaptureGroup: *pl.CaptureGroup,
			Pipeline:     *pl.Pipeline,
		})
	}
	res := FpTrigger{
		Name:          typehelpers.SafeString(apiTrigger.Name),
		Mod:           typehelpers.SafeString(apiTrigger.Mod),
		RootMod:       typehelpers.SafeString(apiTrigger.RootMod),
		Type:          typehelpers.SafeString(apiTrigger.Type),
		Enabled:       *apiTrigger.Enabled,
		Description:   apiTrigger.Description,
		Pipelines:     pls,
		Url:           apiTrigger.Url,
		Title:         apiTrigger.Title,
		Documentation: apiTrigger.Documentation,
		Schedule:      apiTrigger.Schedule,
		Query:         apiTrigger.Query,
		Tags:          make(map[string]string),
	}
	if apiTrigger.Tags != nil {
		res.Tags = *apiTrigger.Tags
	}
	for _, p := range apiTrigger.Params {
		res.Params = append(res.Params, pipelineParamFromApiResponse(p))
	}
	return res
}

type PrintableTrigger struct {
	Items []FpTrigger
}

func (p PrintableTrigger) GetItems() []FpTrigger {
	return p.Items
}

func NewPrintableTrigger(resp *ListTriggerResponse) *PrintableTrigger {
	result := &PrintableTrigger{
		Items: []FpTrigger{},
	}

	if resp.Items != nil {
		result.Items = resp.Items
	}

	return result
}

func NewPrintableTriggerFromSingle(input *FpTrigger) *PrintableTrigger {
	return &PrintableTrigger{
		Items: []FpTrigger{*input},
	}
}

func (p PrintableTrigger) GetTable() (*printers.Table, error) {
	var tableRows []printers.TableRow
	for _, item := range p.Items {

		var description string
		if item.Description != nil {
			description = *item.Description
		}

		distinct := make(map[string]bool)
		for _, i := range item.Pipelines {
			distinct[i.Pipeline] = true
		}

		var pipelineText string
		if len(distinct) == 1 {
			pipelineText = item.getPipelineDisplay(maps.Keys(distinct)[0])
		} else {
			pipelineText = fmt.Sprintf("%d pipelines", len(distinct))
		}

		cells := []any{
			item.Mod[4:],
			item.TriggerDisplayName(),
			item.Enabled,
			pipelineText,
			description,
		}
		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable().WithData(tableRows, p.getColumns()), nil
}

func (PrintableTrigger) getColumns() (columns []string) {
	return []string{"MOD", "NAME", "ENABLED", "PIPELINE", "DESCRIPTION"}
}

type TriggerExecutionResponse struct {
	Results    map[string]interface{}          `json:"results,omitempty"`
	Errors     []perr.ErrorModel               `json:"errors,omitempty"`
	Flowpipe   FlowpipeTriggerResponseMetadata `json:"flowpipe"`
	LastStatus string                          `json:"last_status,omitempty"`
}

type FlowpipeTriggerResponseMetadata struct {
	ProcessID  string     `json:"process_id,omitempty"`
	Name       string     `json:"name,omitempty"`
	Type       string     `json:"type,omitempty"`
	IsStale    *bool      `json:"is_stale,omitempty"`
	LastLoaded *time.Time `json:"last_loaded,omitempty"`
}

type CmdTrigger struct {
	Command string `json:"command" binding:"required,oneof=run reset"`

	// Sepcify execution id, if not specified, a new execution id will be created
	ExecutionID   string                 `json:"execution_id,omitempty"`
	Args          map[string]interface{} `json:"args,omitempty"`
	ArgsString    map[string]string      `json:"args_string,omitempty"`
	ExecutionMode *string                `json:"execution_mode,omitempty" binding:"omitempty,oneof=synchronous asynchronous"`
	WaitRetry     *int                   `json:"wait_retry,omitempty"`
}

func (c *CmdTrigger) GetExecutionMode() string {
	return utils.Deref(c.ExecutionMode, localconstants.DefaultExecutionMode)
}

func (c *CmdTrigger) GetWaitRetry() int {
	return utils.Deref(c.WaitRetry, localconstants.DefaultWaitRetry)
}

func FpTriggerFromModTrigger(t resources.Trigger, rootMod string) (*FpTrigger, error) {
	tt := resources.GetTriggerTypeFromTriggerConfig(t.Config)

	fpTrigger := FpTrigger{
		Name:            t.FullName,
		Mod:             t.GetMod().Name(),
		Type:            tt,
		Description:     t.Description,
		Title:           t.Title,
		Tags:            t.Tags,
		Documentation:   t.Documentation,
		FileName:        t.FileName,
		StartLineNumber: t.StartLineNumber,
		EndLineNumber:   t.EndLineNumber,
		Enabled:         helpers.IsNil(t.Enabled) || *t.Enabled,
		RootMod:         rootMod,
	}

	var pipelineParams []FpPipelineParam
	for i, param := range t.Params {

		var paramDefault any
		if !param.Default.IsNull() {
			paramDefaultGoVal, err := hclhelpers.CtyToGo(param.Default)
			if err != nil {
				return nil, perr.BadRequestWithMessage("unable to convert param default to go value: " + param.Name)
			}
			paramDefault = paramDefaultGoVal
		}

		pipelineParams = append(pipelineParams, FpPipelineParam{
			Name:        param.Name,
			Description: utils.ToStringPointer(param.Description),
			Tags:        param.Tags,
			Optional:    &t.Params[i].Optional,
			Type:        param.Type,
			TypeString:  param.TypeString,
			Default:     paramDefault,
		})

		fpTrigger.Params = pipelineParams
	}

	switch tt {
	case schema.TriggerTypeHttp:
		cfg := t.Config.(*resources.TriggerHttp)
		fpTrigger.Url = &cfg.Url
		for _, method := range cfg.Methods {
			pipelineInfo := method.Pipeline.AsValueMap()
			pipelineName := pipelineInfo["name"].AsString()
			fpTrigger.Pipelines = append(fpTrigger.Pipelines, FpTriggerPipeline{
				CaptureGroup: method.Type,
				Pipeline:     pipelineName,
			})
		}
	case schema.TriggerTypeQuery:
		cfg := t.Config.(*resources.TriggerQuery)
		fpTrigger.Schedule = &cfg.Schedule
		fpTrigger.Query = &cfg.Sql
		for _, capture := range cfg.Captures {
			pipelineInfo := capture.Pipeline.AsValueMap()
			pipelineName := pipelineInfo["name"].AsString()
			fpTrigger.Pipelines = append(fpTrigger.Pipelines, FpTriggerPipeline{
				CaptureGroup: capture.Type,
				Pipeline:     pipelineName,
			})
		}
	case schema.TriggerTypeSchedule:
		cfg := t.Config.(*resources.TriggerSchedule)
		fpTrigger.Schedule = &cfg.Schedule
		pipelineInfo := t.GetPipeline().AsValueMap()
		pipelineName := pipelineInfo["name"].AsString()
		fpTrigger.Pipelines = append(fpTrigger.Pipelines, FpTriggerPipeline{
			CaptureGroup: "default",
			Pipeline:     pipelineName,
		})
	}

	return &fpTrigger, nil
}
