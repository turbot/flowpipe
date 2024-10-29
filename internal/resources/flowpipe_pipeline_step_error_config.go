package resources

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
)

type ErrorConfig struct {
	// circular link to its "parent"
	PipelineStepBase *PipelineStepBase `json:"-"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	If                   *bool                     `json:"if"`
	Ignore               *bool                     `json:"ignore,omitempty" hcl:"ignore,optional" cty:"ignore"`
}

func NewErrorConfig(p *PipelineStepBase) *ErrorConfig {
	return &ErrorConfig{
		PipelineStepBase:     p,
		UnresolvedAttributes: make(map[string]hcl.Expression),
	}
}

func (e *ErrorConfig) Equals(other *ErrorConfig) bool {
	if e == nil || other == nil {
		return false
	}

	if e == nil && other != nil || e != nil && other == nil {
		return false
	}

	for key, expr := range e.UnresolvedAttributes {
		otherExpr, ok := other.UnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	// reverse
	for key := range other.UnresolvedAttributes {
		if _, ok := e.UnresolvedAttributes[key]; !ok {
			return false
		}
	}

	return utils.BoolPtrEqual(e.Ignore, other.Ignore)

}

func (e *ErrorConfig) AppendDependsOn(dependsOn ...string) {
	e.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (e *ErrorConfig) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (e *ErrorConfig) AppendConnectionDependsOn(...string) {
	// not implemented
}

func (e *ErrorConfig) GetPipeline() *Pipeline {
	return e.PipelineStepBase.GetPipeline()
}

func (e *ErrorConfig) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	e.UnresolvedAttributes[name] = expr
}

func (e *ErrorConfig) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeIf:
			e.AddUnresolvedAttribute(name, attr.Expr)
		case schema.AttributeTypeIgnore:
			stepDiags := setBoolAttribute(attr, evalContext, e, "Ignore", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   fmt.Sprintf("Unsupported attribute '%s' in error block", attr.Name),
				Subject:  &attr.NameRange,
			})
		}
	}

	return diags
}

func (e *ErrorConfig) Validate() bool {
	return true
}
