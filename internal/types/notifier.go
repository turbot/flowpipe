package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/printers"
)

type ListNotifierResponse struct {
	Items     []FpNotifier `json:"items"`
	NextToken *string      `json:"next_token,omitempty"`
}

type FpNotifier struct {
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	Title           *string `json:"title,omitempty"`
	Documentation   *string `json:"documentation,omitempty"`
	FileName        string  `json:"file_name,omitempty"`
	StartLineNumber int     `json:"start_line_number,omitempty"`
	EndLineNumber   int     `json:"end_line_number,omitempty"`
}

func FpNotifierFromModNotifier(notifier modconfig.Notifier) (*FpNotifier, error) {
	resp := &FpNotifier{
		Name:        notifier.Name(),
		Description: notifier.GetHclResourceImpl().Description,
	}

	resp.FileName = notifier.GetNotifierImpl().FileName
	resp.StartLineNumber = notifier.GetNotifierImpl().StartLineNumber
	resp.EndLineNumber = notifier.GetNotifierImpl().EndLineNumber

	return resp, nil
}

func FpNotifierFromAPI(apiResp flowpipeapiclient.FpNotifier) FpNotifier {
	var res = FpNotifier{
		Name:        *apiResp.Name,
		Description: apiResp.Description,
	}

	return res
}

func ListNotifierResponseFromAPI(apiResp *flowpipeapiclient.ListNotifierResponse) *ListNotifierResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListNotifierResponse{
		NextToken: apiResp.NextToken,
		Items:     make([]FpNotifier, len(apiResp.Items)),
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = FpNotifierFromAPI(apiItem)
	}
	return res
}

func NewPrintableNotifier(resp *ListNotifierResponse) *PrintableNotifier {
	return &PrintableNotifier{
		Items: resp.Items,
	}
}

func NewPrintableNotifierFromSingle(input *FpNotifier) *PrintableNotifier {
	return &PrintableNotifier{
		Items: []FpNotifier{*input},
	}
}

type PrintableNotifier struct {
	Items []FpNotifier
}

func (p PrintableNotifier) GetItems() []FpNotifier {
	return p.Items
}

func (p PrintableNotifier) GetTable() (*printers.Table, error) {
	var tableRows []printers.TableRow
	for _, item := range p.Items {

		var description string
		if item.Description != nil {
			description = *item.Description
		}

		cells := []any{
			item.Name,
			description,
		}

		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable().WithData(tableRows, p.getColumns()), nil
}

func (PrintableNotifier) getColumns() (columns []string) {
	return []string{"NAME", "DESCRIPTION"}
}
