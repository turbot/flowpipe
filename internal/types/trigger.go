package types

import (
	"fmt"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

type FpTrigger struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description *string                `json:"description,omitempty"`
	Args        map[string]interface{} `json:"args,omitempty"`
	Pipeline    string                 `json:"pipeline"`
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
	if apiResourceType != "ListTriggerResponse" {
		return nil, fmt.Errorf("invalid resource type: %s", apiResourceType)
	}

	lp, ok := r.(*flowpipeapiclient.ListTriggerResponse)
	if !ok {
		return nil, fmt.Errorf("unable to cast to flowpipeapiclient.ListTriggerResponse")
	}

	return lp.Items, nil
}

func (p PrintableTrigger) GetItems() interface{} {
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.FpTrigger)

	if !ok {
		return Table{}, pcerr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.FpTrigger")
	}

	var tableRows []TableRow
	for _, item := range lp {

		description := ""
		if item.Description != nil {
			description = *item.Description
		}
		cells := []interface{}{
			*item.Name,
			description,
			*item.Pipeline,
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
		// {
		// 	Name:        "TYPE",
		// 	Type:        "string",
		// 	Description: "The type of the trigger",
		// },
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
			Name:        "PIPELINE",
			Type:        "string",
			Description: "The name of the pipeline",
		},
	}
}
