package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

var PipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       schema.BlockTypeParam,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipeline,
			LabelNames: []string{schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipelineStep,
			LabelNames: []string{schema.LabelType, schema.LabelName},
		},
		{
			Type:       schema.BlockTypePipelineOutput,
			LabelNames: []string{schema.LabelName},
		},
	},
}

var PipelineOutputBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name:     "value",
			Required: true,
		},
		{
			Name: "depends_on",
		},
		{
			Name: "sensitive",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "precondition"},
		{Type: "postcondition"},
	},
}

var PipelineStepHttpBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeUrl,
			Required: true,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
	},
}

var PipelineStepSleepBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDuration,
			Required: true,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
	},
}
var PipelineStepEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeTo,
			Required: true,
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
	},
}

var PipelineStepEchoBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "text",
		},
		{
			Name: "list_text",
		},
		{
			Name: "for_each",
		},
		{
			Name: schema.AttributeTypeDependsOn,
		},
	},
}
