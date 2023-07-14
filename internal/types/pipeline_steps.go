package types

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/addrs"
	"github.com/turbot/flowpipe/pipeparser/configschema"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
)

type Input map[string]interface{}

// StepOutput is the output from a pipeline.
type StepOutput map[string]interface{}

func (o *StepOutput) Get(key string) interface{} {
	if o == nil {
		return nil
	}
	return (*o)[key]
}

func (o *StepOutput) AsHclVariables() (cty.Value, error) {
	if o == nil {
		return cty.ObjectVal(map[string]cty.Value{}), nil
	}

	variables := make(map[string]cty.Value)
	for key, value := range *o {
		// Check if the value is a Go native data type
		switch v := value.(type) {
		case string:
			variables[key] = cty.StringVal(v)
		case int:
			variables[key] = cty.NumberIntVal(int64(v))
		case float64:
			variables[key] = cty.NumberFloatVal(v)
		case bool:
			variables[key] = cty.BoolVal(v)
		case []string:
			stringValues, ok := value.([]string)
			if !ok {
				// should never happen
				return cty.NilVal, fperr.InternalWithMessage("Failed to cast to []string. Should never happen")
			}

			var vals []cty.Value
			for _, v := range stringValues {
				vals = append(vals, cty.StringVal(v))
			}
			variables[key] = cty.ListVal(vals)
		case []int:
			intValues, ok := value.([]int)
			if !ok {
				// should never happen
				return cty.NilVal, fperr.InternalWithMessage("Failed to cast to []int. Should never happen")
			}
			var vals []cty.Value
			for _, v := range intValues {
				vals = append(vals, cty.NumberIntVal(int64(v)))
			}
			variables[key] = cty.ListVal(vals)
		case []float64:
			floatValues, ok := value.([]float64)
			if !ok {
				// should never happen
				return cty.NilVal, fperr.InternalWithMessage("Failed to cast to []float64. Should never happen")
			}
			var vals []cty.Value
			for _, v := range floatValues {
				vals = append(vals, cty.NumberFloatVal(v))
			}
			variables[key] = cty.ListVal(vals)
		case []bool:
			boolValues, ok := value.([]bool)
			if !ok {
				// should never happen
				return cty.NilVal, fperr.InternalWithMessage("Failed to cast to []bool. Should never happen")
			}
			var vals []cty.Value
			for _, v := range boolValues {
				vals = append(vals, cty.BoolVal(v))
			}
			variables[key] = cty.ListVal(vals)

			// TODO: warning?
			// default:
			// 	return cty.NilVal, fperr.InternalWithMessage("unsupported type for variable: " + key)
		}

	}
	return cty.ObjectVal(variables), nil
}

type StepError struct {
	// TODO: not sure about this
	Detail fperr.ErrorModel `json:"detail"`
}

type NextStep struct {
	StepName string `json:"step_name"`
	DelayMs  int    `json:"delay_ms,omitempty"`
}

type PipelineStepError struct {
	Ignore  bool `yaml:"ignore" json:"ignore"`
	Retries int  `yaml:"retries" json:"retries"`
}

func NewPipelineStep(stepType, stepName string) IPipelineHclStep {
	var step IPipelineHclStep
	switch stepType {
	case configschema.BlockTypePipelineStepHttp:
		s := &PipelineHclStepHttp{}
		step = s
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
	case configschema.BlockTypePipelineStepSleep:
		s := &PipelineHclStepSleep{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	case configschema.BlockTypePipelineStepEmail:
		s := &PipelineHclStepEmail{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	case configschema.BlockTypePipelineStepEcho:
		s := &PipelineHclStepEcho{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	default:
		return nil
	}

	step.SetName(stepName)
	step.SetType(stepType)

	return step
}

type IPipelineHclStep interface {
	GetFullyQualifiedName() string
	GetName() string
	SetName(string)
	GetType() string
	SetType(string)
	IsResolved() bool
	AddUnresolvedAttribute(string, hcl.Expression)
	GetUnresolvedAttributes() map[string]hcl.Expression
	GetInputs(*hcl.EvalContext) (map[string]interface{}, error)
	GetDependsOn() []string
	AppendDependsOn(...string)
	GetFor() string
	GetError() *PipelineStepError
	SetAttributes(hcl.Attributes, *pipeparser.ParseContext) hcl.Diagnostics
}

type PipelineHclStepBase struct {
	Name      string   `json:"name"`
	Type      string   `json:"step_type"`
	DependsOn []string `json:"depends_on,omitempty"`
	Resolved  bool     `json:"resolved,omitempty"`

	// This cant' be serialised
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
}

func (p *PipelineHclStepBase) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	p.UnresolvedAttributes[name] = expr
}

func (p *PipelineHclStepBase) GetUnresolvedAttributes() map[string]hcl.Expression {
	return p.UnresolvedAttributes
}

func (p *PipelineHclStepBase) SetName(name string) {
	p.Name = name
}

func (p *PipelineHclStepBase) GetName() string {
	return p.Name
}

func (p *PipelineHclStepBase) SetType(stepType string) {
	p.Type = stepType
}

func (p *PipelineHclStepBase) GetType() string {
	return p.Type
}

func (p *PipelineHclStepBase) GetDependsOn() []string {
	return p.DependsOn
}

func (p *PipelineHclStepBase) IsResolved() bool {
	return len(p.UnresolvedAttributes) == 0
}

func (p *PipelineHclStepBase) SetResolved(resolved bool) {
	p.Resolved = resolved
}

func (p *PipelineHclStepBase) GetFullyQualifiedName() string {
	return p.Type + "." + p.Name
}

func (p *PipelineHclStepBase) AppendDependsOn(dependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range p.DependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range dependsOn {
		if !existingDeps[dep] {
			p.DependsOn = append(p.DependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

// Direct copy from Terraform source code
func decodeDependsOn(attr *hcl.Attribute) ([]hcl.Traversal, hcl.Diagnostics) {
	var ret []hcl.Traversal
	exprs, diags := hcl.ExprList(attr.Expr)

	for _, expr := range exprs {
		// expr, shimDiags := shimTraversalInString(expr, false)
		// diags = append(diags, shimDiags...)

		// TODO: should we support legacy "expression in string" syntax here?
		// TODO: terraform supports it by calling shimTraversalInString

		traversal, travDiags := hcl.AbsTraversalForExpr(expr)
		diags = append(diags, travDiags...)
		if len(traversal) != 0 {
			ret = append(ret, traversal)
		}
	}

	return ret, diags
}

func (p *PipelineHclStepBase) SetBaseAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics
	var hclDependsOn []hcl.Traversal
	if attr, exists := hclAttributes[configschema.AttributeTypeDependsOn]; exists {
		deps, depsDiags := decodeDependsOn(attr)
		diags = append(diags, depsDiags...)
		hclDependsOn = append(hclDependsOn, deps...)
	}

	if len(diags) > 0 {
		return diags
	}

	var dependsOn []string
	for _, traversal := range hclDependsOn {
		_, addrDiags := addrs.ParseRef(traversal)
		if addrDiags.HasErrors() {
			// We ignore this here, because this isn't a suitable place to return
			// errors. This situation should be caught and rejected during
			// validation.
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  constants.BadDependsOn,
				Detail:   fmt.Sprintf("The depends_on argument must be a reference to another step, but the given value %q is not a valid reference.", traversal),
			})
		}
		parts := TraversalAsStringSlice(traversal)
		if len(parts) != 3 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  constants.BadDependsOn,
				Detail:   "Invalid depends_on format " + strings.Join(parts, "."),
			})
			continue
		}

		dependsOn = append(dependsOn, parts[1]+"."+parts[2])
	}

	p.DependsOn = append(p.DependsOn, dependsOn...)
	return diags
}

// TraversalAsStringSlice converts a traversal to a path string
// (if an absolute traversal is passed - convert to relative)
func TraversalAsStringSlice(traversal hcl.Traversal) []string {
	var parts = make([]string, len(traversal))
	offset := 0

	if !traversal.IsRelative() {
		s := traversal.SimpleSplit()
		parts[0] = s.Abs.RootName()
		offset++
		traversal = s.Rel
	}
	for i, r := range traversal {
		switch t := r.(type) {
		case hcl.TraverseAttr:
			parts[i+offset] = t.Name
		case hcl.TraverseIndex:
			idx, err := hclhelpers.CtyToString(t.Key)
			if err != nil {
				// we do not expect this to fail
				continue
			}
			parts[i+offset] = idx
		}
	}
	return parts
}

var ValidResourceItemTypes = []string{
	configschema.AttributeTypeDependsOn,
}

func (p *PipelineHclStepBase) IsBaseAttributes(name string) bool {
	return helpers.StringSliceContains(ValidResourceItemTypes, name)
}

type PipelineHclStepHttp struct {
	PipelineHclStepBase
	Url string `json:"url"`
}

func (p *PipelineHclStepHttp) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var urlInput string

	if p.UnresolvedAttributes[configschema.AttributeTypeUrl] == nil {
		urlInput = p.Url
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[configschema.AttributeTypeUrl], evalContext, &urlInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(configschema.BlockTypePipelineStep, diags)
		}
	}

	return map[string]interface{}{
		configschema.AttributeTypeUrl: urlInput,
	}, nil
}

func (p *PipelineHclStepHttp) GetFor() string {
	return ""
}

func (p *PipelineHclStepHttp) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineHclStepHttp) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeUrl:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					val, err := attr.Expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse url attribute",
							Subject:  &attr.Range,
						})
						continue
					}
					p.Url = val.AsString()
				}
			}
		default:
			if !p.IsBaseAttributes(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for HTTP Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

type PipelineHclStepSleep struct {
	PipelineHclStepBase
	Duration string `json:"duration"`
}

func (p *PipelineHclStepSleep) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var durationInput string

	if p.UnresolvedAttributes[configschema.AttributeTypeDuration] == nil {
		durationInput = p.Duration
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[configschema.AttributeTypeDuration], evalContext, &durationInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(configschema.BlockTypePipelineStep, diags)
		}
	}

	return map[string]interface{}{
		configschema.AttributeTypeDuration: durationInput,
	}, nil
}

func (p *PipelineHclStepSleep) GetFor() string {
	return ""
}

func (p *PipelineHclStepSleep) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineHclStepSleep) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeDuration:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					val, err := expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse duration attribute",
							Subject:  &attr.Range,
						})
						continue
					}
					p.Duration = val.AsString()
				}
			}
		default:
			if !p.IsBaseAttributes(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Sleep Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

type PipelineHclStepEmail struct {
	PipelineHclStepBase
	To string `json:"to"`
}

func (p *PipelineHclStepEmail) GetFor() string {
	return ""
}

func (p *PipelineHclStepEmail) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineHclStepEmail) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"to": p.To,
	}, nil
}

func (p *PipelineHclStepEmail) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeTo:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {

					val, err := attr.Expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse to attribute",
							Subject:  &attr.Range,
						})
						continue
					}
					p.To = val.AsString()
				}
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Sleep Step: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}
	return diags
}

type PipelineHclStepEcho struct {
	PipelineHclStepBase
	Text     string   `json:"text"`
	ListText []string `json:"list_text"`
}

func (p *PipelineHclStepEcho) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var textInput string
	var listTextInput []string

	if p.UnresolvedAttributes[configschema.AttributeTypeText] == nil {
		textInput = p.Text
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[configschema.AttributeTypeText], evalContext, &textInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError("step", diags)
		}
	}

	if p.UnresolvedAttributes["list_text"] == nil {
		listTextInput = p.ListText
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes["list_text"], evalContext, &listTextInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError("step", diags)
		}
	}

	return map[string]interface{}{
		configschema.AttributeTypeText: textInput,
		"list_text":                    listTextInput,
	}, nil
}

func (p *PipelineHclStepEcho) GetFor() string {
	return ""
}

func (p *PipelineHclStepEcho) GetError() *PipelineStepError {
	return nil
}

func dependsOnFromExpressions(name string, expr hcl.Expression, p IPipelineHclStep) {
	if len(expr.Variables()) == 0 {
		return
	}
	traversals := expr.Variables()
	for _, traversal := range traversals {
		parts := TraversalAsStringSlice(traversal)
		if len(parts) > 0 {
			if parts[0] == configschema.BlockTypePipelineStep {
				dependsOn := parts[1] + "." + parts[2]
				p.AppendDependsOn(dependsOn)
			}
		}
	}
	p.AddUnresolvedAttribute(name, expr)
}

func (p *PipelineHclStepEcho) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case configschema.AttributeTypeText:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					val, err := expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse " + configschema.AttributeTypeText + " attribute",
							Subject:  &attr.Range,
						})
						continue
					}

					p.Text = val.AsString()
				}
			}
		case "list_text":
			if attr.Expr != nil {
				expr := attr.Expr
				val, err := expr.Value(parseContext.EvalCtx)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + "liext_text" + " attribute",
						Subject:  &attr.Range,
					})
					continue
				}

				valueSlice := val.AsValueSlice()
				var stringSlice []string
				for _, v := range valueSlice {
					stringSlice = append(stringSlice, v.AsString())
				}
				p.ListText = stringSlice
			}
		default:
			if !p.IsBaseAttributes(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Echo Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}
