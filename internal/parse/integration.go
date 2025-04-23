package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

var integrationBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{},
}

func DecodeIntegration(configPath string, block *hcl.Block) (resources.Integration, hcl.Diagnostics) {
	if len(block.Labels) != 2 {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid integration block - expected 2 labels, found %d", len(block.Labels)),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	integrationType := block.Labels[0]

	integration := resources.NewIntegrationFromBlock(block)
	if integration == nil {
		diags := hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid integration type '%s'", integrationType),
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}
	hclImplBody, r, diags := block.Body.PartialContent(integrationBlockSchema)
	if len(diags) > 0 {
		return nil, diags
	}

	body := r.(*hclsyntax.Body)

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(configPath),
		Variables: make(map[string]cty.Value),
	}

	diags = parse.DecodeHclBody(body, evalCtx, nil, integration)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = resources.HclImplFromAttributes(integration.GetHclResourceImpl(), hclImplBody.Attributes, evalCtx)
	if len(diags) > 0 {
		return nil, diags
	}

	moreDiags := integration.Validate()
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	integration.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)

	return integration, diags
}
