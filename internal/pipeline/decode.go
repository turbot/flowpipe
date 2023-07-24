package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

func decodeStep(block *hcl.Block, parseCtx *PipelineParseContext) (types.IPipelineStep, hcl.Diagnostics) {
	stepType := block.Labels[0]
	stepName := block.Labels[1]

	step := types.NewPipelineStep(stepType, stepName)
	if step == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid pipeline step type " + stepType,
		}}
	}

	pipelineStepBlockSchema := GetPipelineStepBlockSchema(stepType)
	if pipelineStepBlockSchema == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Pipeline step block schema not found for step " + stepType,
		}}
	}

	stepOptions, rest, diags := block.Body.PartialContent(pipelineStepBlockSchema)

	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, step)
	if len(diags) > 0 {
		return nil, diags
	}

	diags = step.SetAttributes(stepOptions.Attributes, &parseCtx.ParseContext)
	if len(diags) > 0 {
		return nil, diags
	}

	return step, hcl.Diagnostics{}
}

func decodeOutput(block *hcl.Block, parseCtx *PipelineParseContext) (*types.PipelineOutput, hcl.Diagnostics) {

	o := &types.PipelineOutput{
		Name: block.Labels[0],
	}

	outputOptions, rest, diags := block.Body.PartialContent(PipelineOutputBlockSchema)

	if diags.HasErrors() {
		return o, diags
	}

	diags = gohcl.DecodeBody(rest, parseCtx.EvalCtx, outputOptions)
	if len(diags) > 0 {
		return nil, diags
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeSensitive]; exists {
		valDiags := gohcl.DecodeExpression(attr.Expr, nil, &o.Sensitive)
		diags = append(diags, valDiags...)
	}

	if attr, exists := outputOptions.Attributes[schema.AttributeTypeValue]; exists {
		expr := attr.Expr
		if len(expr.Variables()) > 0 {
			traversals := expr.Variables()
			for _, traversal := range traversals {
				parts := hclhelpers.TraversalAsStringSlice(traversal)
				if len(parts) > 0 {
					if parts[0] == schema.BlockTypePipelineStep {
						dependsOn := parts[1] + "." + parts[2]
						o.AppendDependsOn(dependsOn)
					}
				}
			}
		}
		o.UnresolvedValue = attr.Expr

	} else {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing value attribute",
			Subject:  &block.DefRange,
		})
	}

	return o, diags
}
