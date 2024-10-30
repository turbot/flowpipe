package resources

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/iancoleman/strcase"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type LoopQueryStep struct {
	LoopStep

	Database *string        `json:"database,omitempty" hcl:"database,optional" cty:"database"`
	Sql      *string        `json:"sql,omitempty" hcl:"sql,optional" cty:"sql"`
	Args     *[]interface{} `json:"args,omitempty" hcl:"args,optional" cty:"args"`
}

func (l *LoopQueryStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopQueryStep, ok := other.(*LoopQueryStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopQueryStep.LoopStep) {
		return false
	}

	if l.Args == nil && otherLoopQueryStep.Args != nil || l.Args != nil && otherLoopQueryStep.Args == nil {
		return false
	}

	if l.Args != nil {
		if !reflect.DeepEqual(*l.Args, *otherLoopQueryStep.Args) {
			return false
		}
	}

	return utils.PtrEqual(l.Database, otherLoopQueryStep.Database) &&
		utils.PtrEqual(l.Sql, otherLoopQueryStep.Sql)
}

func (l *LoopQueryStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	result, diags := simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), input, evalContext, schema.AttributeTypeDatabase, l.Database)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("query", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeSql, l.Sql)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("query", diags)
	}

	result, diags = interfaceSliceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeArgs, l.Args)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("query", diags)
	}

	return result, nil
}

func (*LoopQueryStep) GetType() string {
	return schema.BlockTypePipelineStepQuery
}

func (l *LoopQueryStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := l.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDatabase, schema.AttributeTypeSql:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeArgs:
			fieldName := strcase.ToCamel(name)
			stepDiags := setInterfaceSliceAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

		case schema.AttributeTypeUntil:
			// already handled in SetAttributes
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Invalid attribute '" + name + "' in the step loop block",
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}
