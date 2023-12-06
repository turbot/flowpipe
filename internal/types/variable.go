package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	typehelpers "github.com/turbot/go-kit/types"
)

// The definition of a single Flowpipe Variable
type Variable struct {
	Type        string  `json:"type"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Default     any     `json:"default"`
	Value       any     `json:"value"`
}

type PrintableVariable struct {
	Items []Variable
}

func NewPrintableVariable(resp *ListVariableResponse) *PrintableVariable {
	return &PrintableVariable{
		Items: resp.Items,
	}
}

func (p PrintableVariable) GetItems() []Variable {
	return p.Items
}

func (p PrintableVariable) GetTable() (Table, error) {
	var tableRows []TableRow
	for _, item := range p.Items {
		var description string
		if item.Description != nil {
			description = *item.Description
		}

		cells := []any{
			item.Name,
			item.Type,
			description,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return NewTable(tableRows, p.getColumns()), nil
}

func (PrintableVariable) getColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the variable",
		},
		{
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the variable",
		},
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Variable description",
		},
	}
}

// This type is used by the API to return a list of variables.
type ListVariableResponse struct {
	Items     []Variable `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}

func ListVariableResponseFromAPIResponse(apiResp *flowpipeapiclient.ListVariableResponse) (*ListVariableResponse, error) {
	if apiResp == nil {
		return nil, nil
	}

	var res = &ListVariableResponse{
		Items:     make([]Variable, len(apiResp.Items)),
		NextToken: apiResp.NextToken,
	}

	for i, apiItem := range apiResp.Items {
		item, err := VariableFromAPIResponse(apiItem)
		if err != nil {
			return nil, err
		}
		res.Items[i] = *item
	}
	return res, nil
}

func VariableFromAPIResponse(apiResp flowpipeapiclient.Variable) (*Variable, error) {
	res := &Variable{
		Name:        typehelpers.SafeString(apiResp.Name),
		Type:        typehelpers.SafeString(apiResp.Type),
		Description: apiResp.Description,
		Default:     apiResp.Default,
		Value:       apiResp.Value,
	}

	return res, nil
}
