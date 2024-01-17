package types

import (
	"fmt"
	"github.com/turbot/pipe-fittings/schema"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/turbot/flowpipe/internal/sanitize"
	typehelpers "github.com/turbot/go-kit/types"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

type FpTrigger struct {
	Name            string              `json:"name"`
	Type            string              `json:"type"`
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
}

type FpTriggerPipeline struct {
	CaptureGroup string `json:"capture_group"`
	Pipeline     string `json:"pipeline"`
}

func (t FpTrigger) String(_ *sanitize.Sanitizer, opts RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	output := ""
	keyWidth := 10
	if t.Description != nil {
		keyWidth = 13
	}

	if t.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:").Bold(), *t.Title)
	}
	output += fmt.Sprintf("%-*s%s", keyWidth, au.Blue("Name:").Bold(), t.Name)

	switch t.Type {
	case schema.TriggerTypeHttp, schema.TriggerTypeQuery:
		for _, pipeline := range t.Pipelines {
			output += fmt.Sprintf("\n%-*s%s %s", keyWidth, au.Blue("Pipeline:").Bold(), au.BrightBlack(strings.ToUpper(pipeline.CaptureGroup)), pipeline.Pipeline)
		}
	case schema.TriggerTypeSchedule:
		output += fmt.Sprintf("\n%-*s%s", keyWidth, au.Blue("Pipeline:").Bold(), t.Pipelines[0].Pipeline)
	}
	output += fmt.Sprintf("\n%-*s%s", keyWidth, au.Blue("Type:").Bold(), t.Type)
	if t.Url != nil {
		output += fmt.Sprintf("\n%-*s%s", keyWidth, au.Blue("Url:").Bold(), *t.Url)
	}
	if t.Schedule != nil {
		output += fmt.Sprintf("\n%-*s%s", keyWidth, au.Blue("Schedule:").Bold(), *t.Schedule)
	}
	if len(t.Tags) > 0 {
		output += fmt.Sprintf("\n%s\n", au.Blue("Tags:").Bold())
		isFirstTag := true
		for k, v := range t.Tags {
			if isFirstTag {
				output += "  " + k + " = " + v
				isFirstTag = false
			} else {
				output += ", " + k + " = " + v
			}
		}
	}
	if t.Description != nil {
		output += fmt.Sprintf("\n\n%s\n", au.Blue("Description:").Bold())
		output += *t.Description
	}

	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	return output
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
		Description:   apiTrigger.Description,
		Pipelines:     pls,
		Url:           apiTrigger.Url,
		Title:         apiTrigger.Title,
		Documentation: apiTrigger.Documentation,
		Schedule:      apiTrigger.Schedule,
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

		var url string
		if item.Url != nil {
			url = *item.Url
		}

		var schedule string
		if item.Schedule != nil {
			schedule = *item.Schedule
		}

		cells := []any{
			item.Name,
			item.Type,
			description,
			url,
			schedule,
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
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the trigger",
		},
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Trigger description",
		},
		{
			Name:        "URL",
			Type:        "string",
			Description: "HTTP Trigger URL",
		},
		{
			Name:        "SCHEDULE",
			Type:        "string",
			Description: "Schedule or Interval",
		},
	}
}
