package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

// The definition of a single Flowpipe Variable
type Variable struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type PrintableVariable struct {
	Items interface{}
}

func (PrintableVariable) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	return nil, nil
	// apiResourceType := r.GetResourceType()
	// if apiResourceType != "ListVariableResponse" {
	// 	return nil, fperr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
	// }

	// lp, ok := r.(*flowpipeapiclient.ListVariableResponse)
	// if !ok {
	// 	return nil, fperr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListVariableResponse")
	// }

	// return lp.Items, nil
}

func (p PrintableVariable) GetItems() interface{} {
	return p.Items
}

func (p PrintableVariable) GetTable() (Table, error) {
	return Table{}, nil
	// lp, ok := p.Items.([]flowpipeapiclient.Variable)

	// if !ok {
	// 	return Table{}, fperr.BadRequestWithMessage("unable to cast to []flowpipeapiclient.Variable")
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

func (PrintableVariable) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the variable",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the variable",
		}}
}

// This type is used by the API to return a list of variables.
type ListVariableResponse struct {
	Items     []Variable `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}
