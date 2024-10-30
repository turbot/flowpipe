package resources

import (
	"log/slog"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type LoopDefn interface {
	GetType() string
	UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error)
	SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics
	Equals(LoopDefn) bool
	AppendDependsOn(...string)
	AppendCredentialDependsOn(...string)
	AppendConnectionDependsOn(...string)
	AddUnresolvedAttribute(string, hcl.Expression)
	GetUnresolvedAttributes() map[string]hcl.Expression
	ResolveUntil(evalContext *hcl.EvalContext) (bool, hcl.Diagnostics)
}

func GetLoopDefn(stepType string, p *PipelineStepBase, hclRange *hcl.Range) LoopDefn {
	loopStep := LoopStep{
		PipelineStepBase:     p,
		UnresolvedAttributes: map[string]hcl.Expression{},
		Range:                hclRange,
	}

	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		return &LoopHttpStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepSleep:
		return &LoopSleepStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepQuery:
		return &LoopQueryStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepPipeline:
		return &LoopPipelineStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepTransform:
		return &LoopTransformStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepContainer:
		return &LoopContainerStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepInput:
		return &LoopInputStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepMessage:
		return &LoopMessageStep{
			LoopStep: loopStep,
		}
	case schema.BlockTypePipelineStepFunction:
		return &LoopFunctionStep{
			LoopStep: loopStep,
		}
	}

	return nil
}

type LoopStep struct {
	// circular link to its "parent"
	PipelineStepBase     *PipelineStepBase `json:"-"`
	Range                *hcl.Range        `json:"-"`
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	Until                *bool
}

func (l *LoopStep) ResolveUntil(evalContext *hcl.EvalContext) (bool, hcl.Diagnostics) {

	diags := hcl.Diagnostics{}
	if l.Until != nil {
		return *l.Until, diags
	}

	if l.UnresolvedAttributes[schema.AttributeTypeUntil] != nil {
		until, diags := l.UnresolvedAttributes[schema.AttributeTypeUntil].Value(evalContext)
		if diags.HasErrors() {
			slog.Error("Error resolving until", "diags", diags)
			return false, diags
		}

		if until == cty.NilVal {
			return false, diags
		}

		return until.True(), diags
	}

	return false, diags
}

func (l *LoopStep) GetUnresolvedAttributes() map[string]hcl.Expression {
	return l.UnresolvedAttributes
}

func (l *LoopStep) AppendDependsOn(dependsOn ...string) {
	l.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (*LoopStep) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (*LoopStep) AppendConnectionDependsOn(...string) {
	// not implemented
}

func (l *LoopStep) GetPipeline() *Pipeline {
	return l.PipelineStepBase.GetPipeline()
}

func (l *LoopStep) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	l.UnresolvedAttributes[name] = expr
}

func (l *LoopStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	if attr, ok := hclAttributes[schema.AttributeTypeUntil]; ok {
		stepDiags := setBoolAttributeWithResultReference(attr, evalContext, l, "Until", true, true)
		if stepDiags.HasErrors() {
			diags = append(diags, stepDiags...)
		}
	}

	if l.Until == nil && l.UnresolvedAttributes[schema.AttributeTypeUntil] == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing required attribute",
			Detail:   "The argument 'until' is required, but no definition was found",
			Subject:  l.Range,
		})
	}

	return diags
}

func (l LoopStep) Equals(other LoopStep) bool {

	// Compare UnresolvedAttributes (map comparison)
	otherUnresolvedAttributes := other.GetUnresolvedAttributes()
	if len(l.UnresolvedAttributes) != len(otherUnresolvedAttributes) {
		return false
	}

	for key, expr := range l.UnresolvedAttributes {
		otherExpr, ok := otherUnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	// and reverse
	for key := range otherUnresolvedAttributes {
		if _, ok := l.UnresolvedAttributes[key]; !ok {
			return false
		}
	}

	return utils.BoolPtrEqual(l.Until, other.Until)
}

type LoopSleepStep struct {
	LoopStep

	UnresolvedAttributes hcl.Attributes `json:"-"`
	Duration             *string        `json:"duration,omitempty"`
}

func (l *LoopSleepStep) Equals(other LoopDefn) bool {

	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || !helpers.IsNil(l) && other == nil {
		return false
	}

	otherLoopSleepStep, ok := other.(*LoopSleepStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopSleepStep.LoopStep) {
		return false
	}

	return utils.BoolPtrEqual(l.Until, otherLoopSleepStep.Until) &&
		utils.PtrEqual(l.Duration, otherLoopSleepStep.Duration)
}

func (l *LoopSleepStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	result, diags := simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), input, evalContext, schema.AttributeTypeDuration, l.Duration)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("sleep", diags)
	}

	return result, nil
}

func (*LoopSleepStep) GetType() string {
	return schema.BlockTypePipelineStepSleep
}

func (s *LoopSleepStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := s.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDuration:
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, s, "Duration", true, true)
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
