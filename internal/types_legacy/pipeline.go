package types_legacy

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
)

// The definition of a single Flowpipe Pipeline
type Pipeline struct {
	Type     string                   `yaml:"type" json:"type"`
	Name     string                   `yaml:"name" json:"name"`
	Steps    map[string]*PipelineStep `yaml:"steps" json:"steps"`
	Parallel bool                     `yaml:"parallel" json:"parallel"`
	Args     types.Input              `yaml:"args" json:"args"`
	Output   string                   `yaml:"output,omitempty" json:"output,omitempty"`
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

func (p PrintablePipeline) GetTable() (types.Table, error) {
	lp, ok := p.Items.([]flowpipeapiclient.Pipeline)

	if !ok {
		return types.Table{}, fperr.BadRequestWithMessage("Unable to cast to []flowpipeapiclient.Pipeline")
	}

	var tableRows []types.TableRow
	for _, item := range lp {
		cells := []interface{}{
			*item.Type,
			*item.Name,
			*item.Parallel,
		}
		tableRows = append(tableRows, types.TableRow{Cells: cells})
	}

	return types.Table{
		Rows:    tableRows,
		Columns: p.GetColumns(),
	}, nil
}

func (PrintablePipeline) GetColumns() (columns []types.TableColumnDefinition) {
	return []types.TableColumnDefinition{
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
	Type      string                  `yaml:"type" json:"type"`
	Name      string                  `yaml:"name" json:"name"`
	Input     string                  `yaml:"input" json:"input_template"`
	DependsOn []string                `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	For       string                  `yaml:"for,omitempty" json:"for,omitempty"`
	Error     types.PipelineStepError `yaml:"error" json:"error"`
}
