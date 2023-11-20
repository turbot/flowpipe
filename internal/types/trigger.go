package types

import (
	"fmt"
	typehelpers "github.com/turbot/go-kit/types"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/pipe-fittings/perr"
)

type FpTrigger struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Description   *string           `json:"description,omitempty"`
	Pipeline      string            `json:"pipeline"`
	Url           *string           `json:"url,omitempty"`
	Title         *string           `json:"title"`
	Documentation *string           `json:"documentation"`
	Tags          map[string]string `json:"tags"`
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
	Items interface{}
}

func (PrintableTrigger) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListTriggerResponse" {
		return nil, fmt.Errorf("invalid resource type: %s", apiResourceType)
	}

	lp, ok := r.(*flowpipeapiclient.ListTriggerResponse)
	if !ok {
		return nil, fmt.Errorf("unable to cast to flowpipeapiclient.ListTriggerResponse")
	}

	return lp.Items, nil
}

func (p PrintableTrigger) GetItems() interface{} {
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.FpTrigger)

	if !ok {
		return Table{}, perr.BadRequestWithMessage("unable to cast to []flowpipeapiclient.FpTrigger")
	}

	var tableRows []TableRow
	for _, item := range lp {

		var description string
		if item.Description != nil {
			description = *item.Description
		}
		cells := []interface{}{
			*item.Pipeline,
			*item.Type,
			*item.Name,
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
