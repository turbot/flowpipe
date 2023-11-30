package types

// The definition of a single Flowpipe Variable
type Variable struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type PrintableVariable struct {
	Items []Variable
}

func (p PrintableVariable) GetItems() []Variable {
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
	// 	cells := []any{
	// 		*item.Type,
	// 		*item.Name,
	// 		*item.Parallel,
	// 	}
	// 	tableRows = append(tableRows, TableRow{Cells: cells})
	// }

	// return Table{
	// 	Items:    tableRows,
	// 	Columns: p.getColumns(),
	// }, nil
}

//func (PrintableVariable) getColumns() (columns []TableColumnDefinition) {
//	return []TableColumnDefinition{
//		{
//			Name:        "TYPE",
//			Type:        "string",
//			Description: "The type of the variable",
//		},
//		{
//			Name:        "NAME",
//			Type:        "string",
//			Description: "The name of the variable",
//		}}
//}

// This type is used by the API to return a list of variables.
type ListVariableResponse struct {
	Items     []Variable `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}
