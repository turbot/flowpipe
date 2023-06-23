package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type PrintableTrigger struct {
	Items interface{}
}

func (PrintableTrigger) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	return nil, nil
	// apiResourceType := r.GetResourceType()
	// if apiResourceType != "ListTriggerResponse" {
	// 	return nil, fperr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
	// }

	// lp, ok := r.(*flowpipeapiclient.ListTriggerResponse)
	// if !ok {
	// 	return nil, fperr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListTriggerResponse")
	// }

	// return lp.Items, nil
}

func (p PrintableTrigger) GetItems() interface{} {
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	return Table{}, nil
	// lp, ok := p.Items.([]flowpipeapiclient.Trigger)

	// if !ok {
	// 	return Table{}, fperr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Trigger")
	// }

	// var tableRows []TableRow
	// for _, item := range lp {
	// 	cells := []interface{}{
	// 		*item.Type,
	// 		*item.Name,
	// 		*item.Parallel,
	// 	}
	// 	tableRows = append(tableRows, TableRow{Cells: cells})
	// }

	// return Table{
	// 	Rows:    tableRows,
	// 	Columns: p.GetColumns(),
	// }, nil
}

func (PrintableTrigger) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
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
	}
}

// This type is used by the API to return a list of triggers.
type ListTriggerResponse struct {
	Items     []Trigger `json:"items"`
	NextToken *string   `json:"next_token,omitempty"`
}
