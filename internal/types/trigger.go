package types

import (
	"fmt"
	"github.com/turbot/pipe-fittings/schema"
	"golang.org/x/exp/maps"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/sanitize"
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

func (t FpTrigger) String(_ *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	output := ""
	statusText := au.Green("Enabled").String()
	if !t.Enabled {
		statusText = au.Red("Disabled").String()
	}
	keyWidth := 10
	if t.Description != nil {
		keyWidth = 13
	}

	if t.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:").Bold(), *t.Title)
	}
	if t.Description != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:").Bold(), *t.Description)
	}
	output += fmt.Sprintf("%-*s%s %s\n", keyWidth, au.Blue("Name:").Bold(), t.Name, statusText)
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Type:").Bold(), t.Type)

	switch t.Type {
	case schema.TriggerTypeHttp:
		if t.Url != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Url:").Bold(), *t.Url)
		}
		for _, pipeline := range t.Pipelines {
			output += fmt.Sprintf("%-*s%s %s\n", keyWidth, au.Blue("Pipeline:").Bold(), au.BrightBlack(strings.ToUpper(pipeline.CaptureGroup)), pipeline.Pipeline)
		}
	case schema.TriggerTypeQuery:
		if t.Schedule != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Schedule:").Bold(), *t.Schedule)
		}
		if t.Query != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Query:").Bold(), *t.Query)
		}
		for _, pipeline := range t.Pipelines {
			output += fmt.Sprintf("%-*s%s %s\n", keyWidth, au.Blue("Pipeline:").Bold(), au.BrightBlack(strings.ToUpper(pipeline.CaptureGroup)), pipeline.Pipeline)
		}
	case schema.TriggerTypeSchedule:
		if t.Schedule != nil {
			output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Schedule:").Bold(), *t.Schedule)
		}
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Pipeline:").Bold(), t.Pipelines[0].Pipeline)
	}

	if len(t.Tags) > 0 {
		output += fmt.Sprintf("%s\n", au.Blue("Tags:").Bold())
		for k, v := range t.Tags {
			output += fmt.Sprintf("- %s %s\n", au.Blue(k+":"), v)
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

func (p PrintableTrigger) GetTable() (Table, error) {
	var tableRows []TableRow
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
			pipelineText = maps.Keys(distinct)[0]
		} else {
			pipelineText = fmt.Sprintf("%d pipelines", len(distinct))
		}

		cells := []any{
			item.getTypeAndName(),
			item.Enabled,
			pipelineText,
			description,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return NewTable(tableRows, p.getColumns()), nil
}

func (PrintableTrigger) getColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
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
