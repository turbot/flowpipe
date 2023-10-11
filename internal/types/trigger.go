package types

import (
	"encoding/json"
	"fmt"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type FpTrigger struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Description   *string           `json:"description,omitempty"`
	Pipeline      string            `json:"pipeline"`
	Url           *string           `json:"url,omitempty"`
	Title         *string           `json:"title"`
	Documentation *string           `json:"documentation"`
	Tags          map[string]string `json:"tags"`
}

// This type is used by the API to return a list of triggers.
type ListTriggerResponse struct {
	Items     []FpTrigger `json:"items"`
	NextToken *string     `json:"next_token,omitempty"`
}

type PrintableTrigger struct {
	Items interface{}
}

func (PrintableTrigger) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType == "ListTriggerResponse" {
		lp, ok := r.(*flowpipeapiclient.ListTriggerResponse)
		if !ok {
			return nil, fmt.Errorf("unable to cast to flowpipeapiclient.ListTriggerResponse")
		}
		return lp.Items, nil
	} else if apiResourceType == "FpTrigger" {
		lp, ok := r.(*flowpipeapiclient.FpTrigger)
		if !ok {
			return nil, fmt.Errorf("unable to cast to flowpipeapiclient.FpTrigger")
		}
		return []flowpipeapiclient.FpTrigger{*lp}, nil
	} else {
		return nil, fmt.Errorf("invalid resource type: %s", apiResourceType)
	}
}

func (p PrintableTrigger) GetItems() interface{} {
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.FpTrigger)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("unable to cast to []flowpipeapiclient.FpTrigger")
	}

	var tableRows []TableRow
	for _, item := range lp {

		description, documentation, title, tags := "", "", "", ""
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
			*item.Pipeline,
			*item.Type,
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

func (PrintableTrigger) GetColumns() (columns []TableColumnDefinition) {
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
			Name:        "TITLE",
			Type:        "string",
			Description: "The title of the trigger",
		},
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Trigger description",
		},
		{
			Name:        "DOCUMENTATION",
			Type:        "string",
			Description: "Trigger documentation",
		},
		{
			Name:        "TAGS",
			Type:        "string",
			Description: "Trigger tags",
		},
	}
}
