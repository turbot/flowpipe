package resources

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepFunction struct {
	PipelineStepBase

	Runtime string `json:"runtime" cty:"runtime"`
	Source  string `json:"source" cty:"source"`
	Handler string `json:"handler" cty:"handler"`

	Event map[string]interface{} `json:"event"`
	Env   map[string]string      `json:"env"`
}

func (p *PipelineStepFunction) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || p != nil && helpers.IsNil(iOther) {
		return false
	}

	other, ok := iOther.(*PipelineStepFunction)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return p.Name == other.Name &&
		p.Runtime == other.Runtime &&
		p.Handler == other.Handler &&
		p.Source == other.Source
}

func (p *PipelineStepFunction) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	res, _, err := p.GetInputs2(evalContext)
	return res, err
}
func (p *PipelineStepFunction) GetInputs2(evalContext *hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {

	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, nil, err
	}

	var allConnectionDependencies []ConnectionDependency

	var env map[string]string
	if p.UnresolvedAttributes[schema.AttributeTypeEnv] == nil {
		env = p.Env
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEnv], evalContext, &args)
		if diags.HasErrors() {
			return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}

		var err error
		env, err = hclhelpers.CtyToGoMapString(args)
		if err != nil {
			return nil, nil, perr.BadRequestWithMessage(p.Name + ": unable to parse env attribute to map[string]string: " + err.Error())
		}
	}
	results[schema.AttributeTypeEnv] = env

	// event
	eventValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeEvent, p.Event)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeEvent] = eventValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// src
	srcValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeSource, p.Source)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeSource] = srcValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// runtime
	runtimeValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeRuntime, p.Runtime)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeRuntime] = runtimeValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// handler
	handlerValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeHandler, p.Handler)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeHandler] = handlerValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	results[schema.LabelName] = p.PipelineName + "." + p.GetFullyQualifiedName()

	return results, allConnectionDependencies, nil
}

func (p *PipelineStepFunction) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSource:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Source = val.AsString()
			}

		case schema.AttributeTypeHandler:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Handler = val.AsString()
			}

		case schema.AttributeTypeRuntime:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val != cty.NilVal {
				p.Runtime = val.AsString()
			}

		case schema.AttributeTypeEnv:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				env, moreErr := hclhelpers.CtyToGoMapString(val)
				if moreErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeEnv + "' attribute to string map",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Env = env
			}
		case schema.AttributeTypeEvent:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val != cty.NilVal {
				events, err := hclhelpers.CtyToGoMapInterface(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeEvent + "' attribute to string map",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Event = events
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Function Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepFunction) Validate() hcl.Diagnostics {
	// validate the base attributes
	diags := p.ValidateBaseAttributes()
	return diags
}
