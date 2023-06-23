package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/fperr"
)

type Input map[string]interface{}

// StepOutput is the output from a pipeline.
type StepOutput map[string]interface{}

func (o *StepOutput) Get(key string) interface{} {
	if o == nil {
		return nil
	}
	return (*o)[key]
}

type StepError struct {
	// TODO: not sure about this
	Detail fperr.ErrorModel `json:"detail"`
}

type NextStep struct {
	StepName string `json:"step_name"`
	DelayMs  int    `json:"delay_ms,omitempty"`
}

// The definition of a single Flowpipe Pipeline
type Pipeline struct {
	Type     string                   `yaml:"type" json:"type"`
	Name     string                   `yaml:"name" json:"name"`
	Steps    map[string]*PipelineStep `yaml:"steps" json:"steps"`
	Parallel bool                     `yaml:"parallel" json:"parallel"`
	Args     Input                    `yaml:"args" json:"args"`
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
	Type      string            `yaml:"type" json:"type"`
	Name      string            `yaml:"name" json:"name"`
	Input     string            `yaml:"input" json:"input_template"`
	DependsOn []string          `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	For       string            `yaml:"for,omitempty" json:"for,omitempty"`
	Error     PipelineStepError `yaml:"error" json:"error"`
}

type PipelineStepError struct {
	Ignore  bool `yaml:"ignore" json:"ignore"`
	Retries int  `yaml:"retries" json:"retries"`
}

// This type is used by the API to return a list of pipelines.
type ListPipelineResponse struct {
	Items     []PipelineHcl `json:"items"`
	NextToken *string       `json:"next_token,omitempty"`
}

type RunPipelineResponse struct {
	ExecutionID           string `json:"execution_id"`
	PipelineExecutionID   string `json:"pipeline_execution_id"`
	ParentStepExecutionID string `json:"parent_step_execution_id"`
}

type CmdPipeline struct {
	Command string `json:"command" binding:"required,oneof=run"`
}
