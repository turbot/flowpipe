package types

import (
	"fmt"
	"github.com/turbot/flowpipe/internal/sanitize"
	typehelpers "github.com/turbot/go-kit/types"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/pipe-fittings/perr"
)

// TODO kai review omitempty
type FpTrigger struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Description   *string           `json:"description,omitempty"`
	Pipeline      string            `json:"pipeline"`
	Url           *string           `json:"url,omitempty"`
	Title         *string           `json:"title,omitempty"`
	Documentation *string           `json:"documentation,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// This type is used by the API to return a list of triggers.
type ListTriggerResponse struct {
	Items     []FpTrigger `json:"items"`
	NextToken *string     `json:"next_token,omitempty"`
}

func (o ListTriggerResponse) GetResourceType() string {
	return "ListTriggerResponse"
}

func ListTriggerResponseFromAPI(apiResp *flowpipeapiclient.ListTriggerResponse) *ListTriggerResponse {
	if apiResp == nil {
		return nil
	}

	var res = &ListTriggerResponse{
		NextToken: apiResp.NextToken,
		Items:     make([]FpTrigger, len(apiResp.Items)),
	}
	for i, apiItem := range apiResp.Items {
		res.Items[i] = FpTriggerFromAPI(apiItem)
	}
	return res
}

func FpTriggerFromAPI(apiTrigger flowpipeapiclient.FpTrigger) FpTrigger {
	res := FpTrigger{
		Name:          typehelpers.SafeString(apiTrigger.Name),
		Type:          typehelpers.SafeString(apiTrigger.Type),
		Description:   apiTrigger.Description,
		Pipeline:      typehelpers.SafeString(apiTrigger.Pipeline),
		Url:           apiTrigger.Url,
		Title:         apiTrigger.Title,
		Documentation: apiTrigger.Documentation,
		Tags:          make(map[string]string),
	}
	if apiTrigger.Tags != nil {
		res.Tags = *apiTrigger.Tags
	}
	return res
}

type PrintableTrigger struct {
	Items any
}

func (PrintableTrigger) Transform(r flowpipeapiclient.FlowpipeAPIResource) (any, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListTriggerResponse" {
		return nil, fmt.Errorf("invalid resource type: %s", apiResourceType)
	}

	lp, ok := r.(*ListTriggerResponse)
	if !ok {
		return nil, fmt.Errorf("unable to cast to flowpipeapiclient.ListTriggerResponse")
	}

	return lp.Items, nil
}

func (p PrintableTrigger) GetItems(sanitizer *sanitize.Sanitizer) any {
	items, ok := p.Items.([]FpTrigger)
	if !ok {
		// not expected
		return []any{}
	}

	sanitizedItems := make([]any, len(items))
	for i, item := range items {
		sanitizedItems[i] = sanitizer.SanitizeStruct(item)
	}
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	lp, ok := p.Items.([]FpTrigger)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("unable to cast to []FpTrigger")
	}

	var tableRows []TableRow
	for _, item := range lp {

		var description string
		if item.Description != nil {
			description = *item.Description
		}
		cells := []any{
			item.Pipeline,
			item.Type,
			item.Name,
			description,
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
		{
			Name:        "PIPELINE",
			Type:        "string",
			Description: "The name of the pipeline",
		},
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
		{
			Name:        "DESCRIPTION",
			Type:        "string",
			Description: "Trigger description",
		},
	}
}
