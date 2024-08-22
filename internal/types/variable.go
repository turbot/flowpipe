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

func NewPrintableVariable(resp *ListVariableResponse) *PrintableVariable {
	result := &PrintableVariable{}

	for _, item := range resp.Items {
		result.Items = append(result.Items, *item)
	}

	return result
}

func NewPrintableVariableFromSingle(input *FpVariable) *PrintableVariable {
	return &PrintableVariable{
		Items: []FpVariable{*input},
	}
}

type PrintableVariable struct {
	Items []FpVariable
}

func (p PrintableVariable) GetItems() []FpVariable {
	return p.Items
}

func (p PrintableVariable) GetTable() (*printers.Table, error) {
	var tableRows []printers.TableRow
	for _, item := range p.Items {

		var description string
		if item.Description != nil {
			description = *item.Description
		}

		cells := []any{
			item.ModName,
			item.ResourceName,
			item.Type,
			description,
			item.ValueDefault,
			item.Value,
		}

		tableRows = append(tableRows, printers.TableRow{Cells: cells})
	}

	return printers.NewTable().WithData(tableRows, p.getColumns()), nil
}

func (PrintableVariable) getColumns() (columns []string) {
	return []string{"MOD NAME", "NAME", "TYPE", "DESCRIPTION", "DEFAULT", "VALUE"}
}

// This type is used by the API to return a list of variables.
type ListVariableResponse struct {
	Items     []*FpVariable `json:"items"`
	NextToken *string       `json:"next_token,omitempty"`
}

type FpVariable struct {
	ModName         string      `json:"mod_name"`
	Type            string      `json:"type"`
	TypeString      string      `json:"type_string"`
	QualifiedName   string      `json:"qualified_name"`
	ResourceName    string      `json:"resource_name"`
	Description     *string     `json:"description,omitempty"`
	ValueDefault    interface{} `json:"value_default,omitempty" `
	Value           interface{} `json:"value,omitempty"`
	FileName        string      `json:"file_name,omitempty"`
	StartLineNumber int         `json:"start_line_number,omitempty"`
	EndLineNumber   int         `json:"end_line_number,omitempty"`
}

func (p FpVariable) String(sanitizer *sanitize.Sanitizer, opts sanitize.RenderOptions) string {
	au := aurora.NewAurora(opts.ColorEnabled)
	output := ""
	keyWidth := 9
	if p.Description != nil && len(*p.Description) > 0 {
		keyWidth = 13
	}
	// deliberately shadow the receiver with a sanitized version of the struct
	var err error
	if p, err = sanitize.SanitizeStruct(sanitizer, p); err != nil {
		return ""
	}

	output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Name:"), p.ResourceName)

	if p.Description != nil && len(*p.Description) > 0 {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Description:"), *p.Description)
	}

	if p.Type != "" {
		output += fmt.Sprintf("%-*s%s\n", keyWidth, au.Blue("Type:"), p.Type)
	}

	if p.ValueDefault != nil {
		output += fmt.Sprintf("%-*s%v\n", keyWidth, au.Blue("Default:"), p.ValueDefault)
	}

	if p.Value != nil {
		output += fmt.Sprintf("%-*s%v\n", keyWidth, au.Blue("Value:"), p.Value)
	}

	return output
}

func FpVariableFromApi(apiVariable flowpipeapiclient.FpVariable) *FpVariable {
	var res = FpVariable{
		ModName:       *apiVariable.ModName,
		Type:          *apiVariable.Type,
		TypeString:    *apiVariable.TypeString,
		QualifiedName: *apiVariable.QualifiedName,
		ResourceName:  *apiVariable.ResourceName,
		Description:   apiVariable.Description,
	}

	if !helpers.IsNil(apiVariable.ValueDefault) {
		res.ValueDefault = *apiVariable.ValueDefault
	}

	if !helpers.IsNil(apiVariable.Value) {
		res.Value = *apiVariable.Value
	}

	return &res
}

func FpVariableFromModVariable(variable *modconfig.Variable) *FpVariable {
	return &FpVariable{
		ModName:         variable.ModName,
		Type:            variable.TypeString,
		TypeString:      variable.TypeString,
		QualifiedName:   variable.Name(),
		ResourceName:    variable.ResourceName,
		Description:     variable.Description,
		ValueDefault:    variable.DefaultGo,
		Value:           variable.ValueGo,
		StartLineNumber: variable.ValueSourceStartLineNumber,
		EndLineNumber:   variable.ValueSourceEndLineNumber,
		FileName:        variable.ValueSourceFileName,
	}
}

func ListVariableResponseFromAPI(apiResp *flowpipeapiclient.ListVariableResponse) *ListVariableResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListVariableResponse{
		NextToken: apiResp.NextToken,
		Items:     make([]*FpVariable, len(apiResp.Items)),
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = FpVariableFromApi(apiItem)
	}
	return res
}
