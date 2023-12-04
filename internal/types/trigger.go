package types

import (
	typehelpers "github.com/turbot/go-kit/types"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

// TODO kai review omitempty
type FpTrigger struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Description   *string           `json:"description,omitempty"`
	Pipeline      string            `json:"pipeline"`
	Url           *string           `json:"url,omitempty"`
	Title         *string           `json:"title,omitempty"`
	Documentation *string           `json:"documentation,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	Schedule      *string           `json:"schedule,omitempty"`
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
	res := FpTrigger{
		Name:          typehelpers.SafeString(apiTrigger.Name),
		Type:          typehelpers.SafeString(apiTrigger.Type),
		Description:   apiTrigger.Description,
		Pipeline:      typehelpers.SafeString(apiTrigger.Pipeline),
		Url:           apiTrigger.Url,
		Title:         apiTrigger.Title,
		Documentation: apiTrigger.Documentation,
		// Schedule:      apiTrigger.Schedule,
		Tags: make(map[string]string),
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
			item.Pipeline,
			item.Type,
			item.Name,
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
			Name:        "PIPELINE",
			Type:        "string",
			Description: "The name of the pipeline",
		},
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
