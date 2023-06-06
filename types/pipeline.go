package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/fperr"
)

type Input map[string]interface{}

// Output is the output from a pipeline.
type Output map[string]interface{}

func (o *Output) Get(key string) interface{} {
	return (*o)[key]
}

// The definition of a single Flowpipe Pipeline
type Pipeline struct {
	Type     string                   `json:"type"`
	Name     string                   `json:"name"`
	Steps    map[string]*PipelineStep `json:"steps"`
	Parallel bool                     `json:"parallel"`
	Args     Input                    `json:"args"`
	Output   string                   `json:"output,omitempty"`
}

type PrintablePipeline struct {
	Items interface{}
}

func (PrintablePipeline) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	apiResourceType := r.GetResourceType()
	if apiResourceType != "ListPipelineResponse" {
		return nil, fperr.BadRequestWithMessage("Invalid resource type: " + apiResourceType)
	}

	lp, ok := r.(*flowpipeapiclient.ListPipelineResponse)
	if !ok {
		return nil, fperr.BadRequestWithMessage("Unable to cast to flowpipeapiclient.ListPipelineResponse")
	}

	return lp.Items, nil
}

func (p PrintablePipeline) GetItems() interface{} {
	return p.Items
}

func (p PrintablePipeline) GetTable() (Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.Pipeline)

	if !ok {
		return Table{}, fperr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Pipeline")
	}

	var tableRows []TableRow
	for _, item := range lp {
		cells := []interface{}{
			*item.Type,
			*item.Name,
			*item.Parallel,
		}
		tableRows = append(tableRows, TableRow{Cells: cells})
	}

	return Table{
		Rows:    tableRows,
		Columns: p.GetColumns(),
	}, nil
}

func (PrintablePipeline) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the pipeline",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the pipeline",
		},
		{
			Name:        "PARALLEL",
			Type:        "boolean",
			Description: "Whether the pipeline is parallel",
		},
	}
}

type PipelineStep struct {
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Input     string   `json:"input_template"`
	DependsOn []string `json:"depends_on"`
	For       string   `json:"for,omitempty"`
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []Pipeline `json:"items"`
	NextToken *string    `json:"next_token,omitempty"`
}

type RunPipelineResponse struct {
	ExecutionID string `json:"execution_id"`
}
