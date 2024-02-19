package types

import (
	"github.com/turbot/pipe-fittings/printers"
)

// This type is used by the API to return a list of integrations.
type ListIntegrationResponse struct {
	Items     []FpIntegration `json:"items"`
	NextToken *string         `json:"next_token,omitempty"`
}

type FpIntegration struct {
	Name            string  `json:"name"`
	Type            string  `json:"type"`
	Description     *string `json:"description,omitempty"`
	Title           *string `json:"title,omitempty"`
	Documentation   *string `json:"documentation,omitempty"`
	FileName        string  `json:"file_name,omitempty"`
	StartLineNumber int     `json:"start_line_number,omitempty"`
	EndLineNumber   int     `json:"end_line_number,omitempty"`
}

func NewPrintableIntegration(resp *ListIntegrationResponse) *PrintableIntegration {
	return &PrintableIntegration{
		Items: resp.Items,
	}
}

type PrintableIntegration struct {
	Items []FpIntegration
}

func (p PrintableIntegration) GetItems() []FpIntegration {
	return p.Items
}

func (p PrintableIntegration) GetTable() (*printers.Table, error) {
	var tableRows []printers.TableRow
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
		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable().WithData(tableRows, p.getColumns()), nil
}

func (PrintableIntegration) getColumns() (columns []string) {
	return []string{"NAME", "TYPE", "DESCRIPTION"}
}
