package types

import (
	"fmt"

	"github.com/logrusorgru/aurora"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
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
	Url             *string `json:"url,omitempty"`
}

func (f FpIntegration) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var output string
	var statusText string
	// left := au.BrightBlack("[")
	// right := au.BrightBlack("]")
	keyWidth := 10
	if f.Description != nil {
		keyWidth = 13
	}

	output += fmt.Sprintf("%-*s%s %s\n", keyWidth, au.Blue("Name:"), f.Name, statusText)
	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Type:"), f.Type)
	if f.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:"), *f.Title)
	}
	if f.Description != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:"), *f.Description)
	}
	if f.Url != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("URL:"), *f.Url)
	}

	return output
}

func ListIntegrationResponseFromAPI(apiResp *flowpipeapiclient.ListIntegrationResponse) *ListIntegrationResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListIntegrationResponse{
		NextToken: apiResp.NextToken,
		Items:     make([]FpIntegration, len(apiResp.Items)),
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = FpIntegrationFromAPI(apiItem)
	}
	return res
}

func FpIntegrationFromAPI(apiIntegration flowpipeapiclient.FpIntegration) FpIntegration {
	res := FpIntegration{
		Name:          typehelpers.SafeString(apiIntegration.Name),
		Type:          typehelpers.SafeString(apiIntegration.Type),
		Description:   apiIntegration.Description,
		Title:         apiIntegration.Title,
		Documentation: apiIntegration.Documentation,
	}
	return res
}

func NewPrintableIntegration(resp *ListIntegrationResponse) *PrintableIntegration {
	return &PrintableIntegration{
		Items: resp.Items,
	}
}

func NewPrintableIntegrationFromSingle(input *FpIntegration) *PrintableIntegration {
	return &PrintableIntegration{
		Items: []FpIntegration{*input},
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
