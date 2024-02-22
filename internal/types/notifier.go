package types

import (
	"fmt"

	"github.com/logrusorgru/aurora"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/sanitize"
)

type ListNotifierResponse struct {
	Items     []FpNotifier `json:"items"`
	NextToken *string      `json:"next_token,omitempty"`
}

type FpNotifier struct {
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	Title           *string    `json:"title,omitempty"`
	Notifies        []FpNotify `json:"notifies,omitempty"`
	FileName        string     `json:"file_name,omitempty"`
	StartLineNumber int        `json:"start_line_number,omitempty"`
	EndLineNumber   int        `json:"end_line_number,omitempty"`
}

type FpNotify struct {
	Integration *string `json:"integration,omitempty"`

	Cc          []string `json:"cc,omitempty"`
	Bcc         []string `json:"bcc,omitempty"`
	Channel     *string  `json:"channel,omitempty"`
	Description *string  `json:"description,omitempty"`
	Subject     *string  `json:"subject,omitempty"`
	Title       *string  `json:"title,omitempty"`
	To          []string `json:"to,omitempty"`
}

func (n FpNotify) printItem(opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)

	// Integration can't be null, it's mandatory field in the schema
	output := fmt.Sprintf("%2s- %s %s\n", "", au.Cyan("Integration:"), *n.Integration)
	if n.Title != nil {
		output += fmt.Sprintf("%4s%s %s\n", "", au.Cyan("Title:"), *n.Title)
	}
	if n.Description != nil {
		output += fmt.Sprintf("%4s%s %s\n", "", au.Cyan("Description:"), *n.Description)
	}
	if n.Subject != nil {
		output += fmt.Sprintf("%4s%s %s\n", "", au.Cyan("Subject:"), *n.Subject)
	}
	if n.Channel != nil {
		output += fmt.Sprintf("%4s%s %s\n", "", au.Cyan("Channel:"), *n.Channel)
	}
	if len(n.To) > 0 {
		output += fmt.Sprintf("%4s%s\n", "", au.Cyan("To:"))
		output += printItems(n.To, 6)
	}
	if len(n.Cc) > 0 {
		output += fmt.Sprintf("%4s%s\n", "", au.Cyan("Cc:"))
		output += printItems(n.Cc, 6)
	}
	if len(n.Bcc) > 0 {
		output += fmt.Sprintf("%4s%s\n", "", au.Cyan("Bcc:"))
		output += printItems(n.Bcc, 6)
	}
	return output
}

func (p FpNotifier) String(_ *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	var output string
	var statusText string

	keyWidth := 10
	if p.Description != nil {
		keyWidth = 13
	}

	output += fmt.Sprintf("%-*s%s %s\n", keyWidth, au.Blue("Name:"), p.Name, statusText)
	if p.Title != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Title:"), *p.Title)
	}
	if p.Description != nil {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:"), *p.Description)
	}

	if len(p.Notifies) > 0 {
		output += fmt.Sprintf("%s\n", au.Blue("Notifies:"))

		for _, n := range p.Notifies {
			output += n.printItem(opts)
		}
	}

	return output
}

func printItems(items []string, baseLeftPad int) string {
	output := ""
	for _, item := range items {
		output += fmt.Sprintf("%*s- %s\n", baseLeftPad, "", item)
	}

	return output
}

func FpNotifierFromModNotifier(notifier modconfig.Notifier) (*FpNotifier, error) {
	resp := &FpNotifier{
		Name:        notifier.Name(),
		Description: notifier.GetHclResourceImpl().Description,
	}

	resp.FileName = notifier.GetNotifierImpl().FileName
	resp.StartLineNumber = notifier.GetNotifierImpl().StartLineNumber
	resp.EndLineNumber = notifier.GetNotifierImpl().EndLineNumber

	for _, notify := range notifier.GetNotifierImpl().Notifies {
		fpNotify := FpNotify{
			Cc:          notify.Cc,
			Bcc:         notify.Bcc,
			Channel:     notify.Channel,
			Description: notify.Description,
			Subject:     notify.Subject,
			Title:       notify.Title,
			To:          notify.To,
		}
		if !helpers.IsNil(notify.Integration) {
			fpNotify.Integration = &notify.Integration.GetHclResourceImpl().FullName
		} else {
			defaultIntegration := "webform.default"
			fpNotify.Integration = &defaultIntegration
		}

		resp.Notifies = append(resp.Notifies, fpNotify)
	}

	return resp, nil
}

func FpNotifierFromAPI(apiResp flowpipeapiclient.FpNotifier) FpNotifier {
	var res = FpNotifier{
		Name:        *apiResp.Name,
		Description: apiResp.Description,
	}

	for _, n := range apiResp.Notifies {
		var notify = FpNotify{
			Cc:          n.Cc,
			Bcc:         n.Bcc,
			Channel:     n.Channel,
			Description: n.Description,
			Subject:     n.Subject,
			Title:       n.Title,
			To:          n.To,
		}
		if n.Integration != nil {
			notify.Integration = n.Integration
		}
		res.Notifies = append(res.Notifies, notify)
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
