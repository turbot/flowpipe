package types

import (
	"fmt"
	"strings"

	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"golang.org/x/exp/maps"

	"github.com/logrusorgru/aurora"
	typehelpers "github.com/turbot/go-kit/types"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

type FpTrigger struct {
	Name            string              `json:"name"`
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
	return res
}

type PrintableTrigger struct {
	Items []FpTrigger
}

func (p PrintableTrigger) GetItems() []FpTrigger {
	return p.Items
}

func NewPrintableTrigger(resp *ListTriggerResponse) *PrintableTrigger {
	return &PrintableTrigger{
		Items: resp.Items,
	}
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
			item.getTypeAndName(),
			item.Enabled,
			pipelineText,
			description,
		}
		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable().WithData(tableRows, p.getColumns()), nil
}

func (PrintableTrigger) getColumns() (columns []printers.TableColumnDefinition) {
	return []printers.TableColumnDefinition{
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the trigger",
		},
		{
			Name:        "ENABLED",
			Type:        "boolean",
			Description: "If true, trigger is enabled",
		},
		{
			Name:        "PIPELINE",
			Type:        "string",
			Description: "Pipeline associated with trigger",
		},
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Trigger description",
		},
	}
}
