package fpconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/zclconf/go-cty/cty/gocty"
)

func decodeStep(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (types.IPipelineStep, hcl.Diagnostics) {
	stepType := block.Labels[0]
	stepName := block.Labels[1]

	// TODO: collect all diags?

	step := types.NewPipelineStep(stepType, stepName)
	if step == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid pipeline step type " + stepType,
			Subject:  &block.DefRange,
		}}
	}

	pipelineStepBlockSchema := GetPipelineStepBlockSchema(stepType)
	if pipelineStepBlockSchema == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Pipeline step block schema not found for step " + stepType,
			Subject:  &block.DefRange,
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

	if errorBlocks := stepOptions.Blocks.ByType()[schema.BlockTypeError]; len(errorBlocks) > 0 {
		if len(errorBlocks) > 1 {
			return nil, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Multiple error blocks found for step " + stepName,
				Subject:  &block.DefRange,
			}}
		}
		errorBlock := errorBlocks[0]

		attributes, diags := errorBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return nil, diags
		}

		ignore := false
		retries := 0

		if attr, exists := attributes[schema.AttributeTypeIgnore]; exists {
			val, diags := attr.Expr.Value(nil)
			if len(diags) > 0 {
				return nil, diags
			}

			var target bool
			if err := gocty.FromCtyValue(val, &target); err != nil {
				return nil, hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Error decoding ignore attribute",
					Detail:   err.Error(),
					Subject:  &block.DefRange,
				}}
			}
			ignore = target
		}

		if attr, exists := attributes[schema.AttributeTypeRetries]; exists {
			val, diags := attr.Expr.Value(nil)
			if len(diags) > 0 {
				return nil, diags
			}

			var target int
			if err := gocty.FromCtyValue(val, &target); err != nil {
				return nil, hcl.Diagnostics{&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Error decoding retries attribute",
					Detail:   err.Error(),
					Subject:  &block.DefRange,
				}}
			}
			retries = target

		}

		errorConfig := &types.ErrorConfig{
			Ignore:  ignore,
			Retries: retries,
		}

		step.SetErrorConfig(errorConfig)
	} else {
		errorConfig := &types.ErrorConfig{
			Ignore:  false,
			Retries: 0,
		}
		step.SetErrorConfig(errorConfig)
	}

	return step, hcl.Diagnostics{}
}

func decodeOutput(block *hcl.Block, parseCtx *FlowpipeConfigParseContext) (*types.PipelineOutput, hcl.Diagnostics) {

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
