package resources

import (
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

func NewThrowConfig(p *PipelineStepBase) *ThrowConfig {
	return &ThrowConfig{
		PipelineStepBase:     p,
		UnresolvedAttributes: make(map[string]hcl.Expression),
	}
}

type ThrowConfig struct {
	// Circular reference to its parent
	PipelineStepBase     *PipelineStepBase         `json:"-"`
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`

	If      *bool
	Message *string
}

func (t *ThrowConfig) AppendDependsOn(dependsOn ...string) {
	t.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (t *ThrowConfig) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (t *ThrowConfig) AppendConnectionDependsOn(...string) {
	// not implemented
}

func (t *ThrowConfig) GetPipeline() *Pipeline {
	return t.PipelineStepBase.GetPipeline()
}

func (t *ThrowConfig) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	t.UnresolvedAttributes[name] = expr
}

func (t *ThrowConfig) SetAttributes(throwBlock *hcl.Block, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeIf:
			t.AddUnresolvedAttribute(name, attr.Expr)
		case schema.AttributeTypeMessage:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, t, true)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				valString, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMessage + " argument to string",
						Subject:  &attr.Range,
					})
					continue
				}

				t.Message = &valString
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid argument",
				Detail:   "Unsupported argument '" + name + "' in throw block",
				Subject:  &attr.Range,
			})
		}
	}

	if t.UnresolvedAttributes[schema.AttributeTypeIf] == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing required argument",
			Detail:   "The argument 'if' is required, but no definition was found.",
			Subject:  &throwBlock.DefRange,
		})
	}

	return diags
}

func (t *ThrowConfig) Equals(other *ThrowConfig) bool {
	if t == nil && other == nil {
		return true
	}

	if t == nil && other != nil || t != nil && other == nil {
		return false
	}

	for k, v := range t.UnresolvedAttributes {
		if other.UnresolvedAttributes[k] == nil || !hclhelpers.ExpressionsEqual(v, other.UnresolvedAttributes[k]) {
			return false
		}
	}

	// reverse
	for k := range other.UnresolvedAttributes {
		if _, ok := t.UnresolvedAttributes[k]; !ok {
			return false
		}
	}

	return utils.BoolPtrEqual(t.If, other.If) &&
		utils.PtrEqual(t.Message, other.Message)
}

func (t *ThrowConfig) Resolve(evalContext *hcl.EvalContext) (*ThrowConfig, hcl.Diagnostics) {
	// make a copy, don't point to the same memory
	newThrowConfig := &ThrowConfig{}
	diags := hcl.Diagnostics{}

	// resolve IF
	if t.If != nil {
		// this should never happened
		newThrowConfig.If = utils.ToPointer(*t.If)
	} else if t.UnresolvedAttributes[schema.AttributeTypeIf] != nil {
		attr := t.UnresolvedAttributes[schema.AttributeTypeIf]
		val, moreDiags := attr.Value(evalContext)
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
		}

		if val != cty.NilVal {
			valBool, err := hclhelpers.CtyToGo(val)
			if err != nil || helpers.IsNil(valBool) || reflect.TypeOf(valBool).Kind() != reflect.Bool {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + schema.AttributeTypeIf + " attribute to bool",
					Detail:   "Unable to parse " + schema.AttributeTypeIf + " attribute to bool",
					Subject:  attr.Range().Ptr(),
				})
			} else {
				valBoolVal := valBool.(bool)
				newThrowConfig.If = &valBoolVal
			}
		}
	}

	if newThrowConfig.If != nil && !*newThrowConfig.If {
		return newThrowConfig, diags
	}

	if t.Message != nil {
		newThrowConfig.Message = utils.ToPointer(*t.Message)
	} else if t.UnresolvedAttributes[schema.AttributeTypeMessage] != nil {
		attr := t.UnresolvedAttributes[schema.AttributeTypeMessage]
		val, moreDiags := attr.Value(evalContext)
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
		}

		if val != cty.NilVal {
			valString, err := hclhelpers.CtyToString(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + schema.AttributeTypeMessage + " attribute to string",
					Detail:   "Unable to parse " + schema.AttributeTypeMessage + " attribute to string",
					Subject:  attr.Range().Ptr(),
				})
			} else {
				newThrowConfig.Message = &valString
			}
		}
	}

	return newThrowConfig, diags
}
