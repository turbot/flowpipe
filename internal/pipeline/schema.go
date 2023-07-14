package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/configschema"
)

var PipelineBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeDescription,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       configschema.BlockTypePipeline,
			LabelNames: []string{configschema.LabelName},
		},
		{
			Type:       configschema.BlockTypePipelineStep,
			LabelNames: []string{configschema.LabelType, configschema.LabelName},
		},
		{
			Type:       configschema.BlockTypePipelineOutput,
			LabelNames: []string{configschema.LabelName},
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
			Name:     configschema.AttributeTypeUrl,
			Required: true,
		},
		{
			Name: configschema.AttributeTypeDependsOn,
		},
	},
}

var PipelineStepSleepBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeDuration,
			Required: true,
		},
		{
			Name: configschema.AttributeTypeDependsOn,
		},
	},
}
var PipelineStepEmailBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     configschema.AttributeTypeTo,
			Required: true,
		},
		{
			Name: configschema.AttributeTypeDependsOn,
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
			Name: configschema.AttributeTypeDependsOn,
		},
	},
}
