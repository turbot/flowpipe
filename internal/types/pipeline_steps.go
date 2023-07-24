package types

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/flowpipe/pipeparser/terraform/addrs"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
)

const (
	HttpMethodGet    = "get"
	HttpMethodPost   = "post"
	HttpMethodPut    = "put"
	HttpMethodDelete = "delete"
	HttpMethodPatch  = "patch"
)

var ValidHttpMethods = []string{
	HttpMethodGet,
	HttpMethodPost,
	HttpMethodPut,
	HttpMethodDelete,
	HttpMethodPatch,
}

type StepForEach struct {
	Index             int         `json:"index" binding:"required"`
	ForEachOutput     *StepOutput `json:"for_each_output,omitempty"`
	ForEachTotalCount int         `json:"for_each_total_count" binding:"required"`
}

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

func NewPipelineStep(stepType, stepName string) IPipelineStep {
	var step IPipelineStep
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		s := &PipelineStepHttp{}
		step = s
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
	case schema.BlockTypePipelineStepSleep:
		s := &PipelineStepSleep{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	case schema.BlockTypePipelineStepEmail:
		s := &PipelineStepEmail{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	case schema.BlockTypePipelineStepEcho:
		s := &PipelineStepEcho{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	default:
		return nil
	}

	step.SetName(stepName)
	step.SetType(stepType)

	return step
}

// A common interface that all pipeline steps must implement
type IPipelineStep interface {
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
	GetForEach() hcl.Expression
	GetError() *PipelineStepError
	SetAttributes(hcl.Attributes, *pipeparser.ParseContext) hcl.Diagnostics
}

// A common base struct that all pipeline steps must embed
type PipelineStepBase struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Name        string   `json:"name"`
	Type        string   `json:"step_type"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Resolved    bool     `json:"resolved,omitempty"`

	// This cant' be serialised
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	ForEach              hcl.Expression            `json:"-"`
}

func (p *PipelineStepBase) GetForEach() hcl.Expression {
	return p.ForEach
}

func (p *PipelineStepBase) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	p.UnresolvedAttributes[name] = expr
}

func (p *PipelineStepBase) GetUnresolvedAttributes() map[string]hcl.Expression {
	return p.UnresolvedAttributes
}

func (p *PipelineStepBase) SetName(name string) {
	p.Name = name
}

func (p *PipelineStepBase) GetName() string {
	return p.Name
}

func (p *PipelineStepBase) SetType(stepType string) {
	p.Type = stepType
}

func (p *PipelineStepBase) GetType() string {
	return p.Type
}

func (p *PipelineStepBase) GetDependsOn() []string {
	return p.DependsOn
}

func (p *PipelineStepBase) IsResolved() bool {
	return len(p.UnresolvedAttributes) == 0
}

func (p *PipelineStepBase) SetResolved(resolved bool) {
	p.Resolved = resolved
}

func (p *PipelineStepBase) GetFullyQualifiedName() string {
	return p.Type + "." + p.Name
}

func (p *PipelineStepBase) AppendDependsOn(dependsOn ...string) {
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

func (p *PipelineStepBase) SetBaseAttributes(hclAttributes hcl.Attributes) hcl.Diagnostics {
	var diags hcl.Diagnostics
	var hclDependsOn []hcl.Traversal
	if attr, exists := hclAttributes[schema.AttributeTypeDependsOn]; exists {
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
			// TODO: won't work with indices, i.e. text = "foo ${step.echo[1].text}"
			// TODO: ^^ only when we have for_each working properly
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  constants.BadDependsOn,
				Detail:   "Invalid depends_on format " + strings.Join(parts, "."),
			})
			continue
		}

		dependsOn = append(dependsOn, parts[1]+"."+parts[2])
	}

	if attr, exists := hclAttributes[schema.AttributeTypeForEach]; exists {
		p.ForEach = attr.Expr

		traversals := attr.Expr.Variables()

		for _, t := range traversals {
			parts := TraversalAsStringSlice(t)
			if len(parts) >= 3 {
				if helpers.StringSliceContains(ValidDependsOnTypes, parts[0]) {
					if len(parts) >= 3 {
						dependsOn = append(dependsOn, parts[1]+"."+parts[2])
					}
				}
			}
		}
	}

	p.DependsOn = append(p.DependsOn, dependsOn...)

	if attr, exists := hclAttributes[schema.AttributeTypeTitle]; exists {
		title, diag := hclhelpers.AttributeToString(attr, nil, false)
		if diag != nil && diag.Severity == hcl.DiagError {
			diags = append(diags, diag)
		} else {
			p.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		description, diag := hclhelpers.AttributeToString(attr, nil, false)
		if diag != nil && diag.Severity == hcl.DiagError {
			diags = append(diags, diag)
		} else {
			p.Description = description
		}
	}

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

var ValidBaseStepAttributes = []string{
	schema.AttributeTypeTitle,
	schema.AttributeTypeDescription,
	schema.AttributeTypeDependsOn,
	schema.AttributeTypeForEach,
}

var ValidDependsOnTypes = []string{
	schema.BlockTypePipelineStep,
}

func (p *PipelineStepBase) IsBaseAttribute(name string) bool {
	return helpers.StringSliceContains(ValidBaseStepAttributes, name)
}

type PipelineStepHttp struct {
	PipelineStepBase

	Url              *string                `json:"url" binding:"required"`
	RequestTimeoutMs *int64                 `json:"request_timeout_ms,omitempty"`
	Method           *string                `json:"method,omitempty"`
	Insecure         *bool                  `json:"insecure,omitempty"`
	RequestBody      *string                `json:"request_body,omitempty"`
	RequestHeaders   map[string]interface{} `json:"request_headers,omitempty"`
}

func (p *PipelineStepHttp) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var urlInput string
	if p.UnresolvedAttributes[schema.AttributeTypeUrl] == nil {
		if p.Url == nil {
			return nil, fperr.InternalWithMessage("Url must be supplied")
		}
		urlInput = *p.Url
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeUrl], evalContext, &urlInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	inputs := map[string]interface{}{
		schema.AttributeTypeUrl: urlInput,
	}

	if p.Method != nil {
		inputs[schema.AttributeTypeMethod] = *p.Method
	}

	if p.RequestTimeoutMs != nil {
		inputs[schema.AttributeTypeRequestTimeoutMs] = *p.RequestTimeoutMs
	}

	if p.Insecure != nil {
		inputs[schema.AttributeTypeInsecure] = *p.Insecure
	}

	if p.RequestBody != nil {
		inputs[schema.AttributeTypeRequestBody] = *p.RequestBody
	}

	if p.RequestHeaders != nil {
		inputs[schema.AttributeTypeRequestHeaders] = p.RequestHeaders
	}

	return inputs, nil
}

func (p *PipelineStepHttp) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineStepHttp) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeUrl:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					urlString, diag := hclhelpers.AttributeToString(attr, parseContext.EvalCtx, false)
					if diag != nil && diag.Severity == hcl.DiagError {
						diags = append(diags, diag)
						continue
					}
					p.Url = urlString
				}
			}
		case schema.AttributeTypeRequestTimeoutMs:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					requestTimeoutMs, diag := hclhelpers.AttributeToInt(attr, parseContext.EvalCtx, false)
					if diag != nil && diag.Severity == hcl.DiagError {
						diags = append(diags, diag)
						continue
					}
					p.RequestTimeoutMs = requestTimeoutMs
				}
			}
		case schema.AttributeTypeMethod:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					method, diag := hclhelpers.AttributeToString(attr, parseContext.EvalCtx, false)
					if diag != nil && diag.Severity == hcl.DiagError {
						diags = append(diags, diag)
						continue
					}

					if method != nil {
						if !helpers.StringSliceContains(ValidHttpMethods, strings.ToLower(*method)) {
							diags = append(diags, &hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  "Invalid HTTP method: " + *method,
								Subject:  &attr.Range,
							})
							continue
						}
						p.Method = method
					}
				}
			}
		case schema.AttributeTypeInsecure:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					insecure, diag := hclhelpers.AttributeToBool(attr, parseContext.EvalCtx, false)
					if diag != nil && diag.Severity == hcl.DiagError {
						diags = append(diags, diag)
						continue
					}
					p.Insecure = insecure
				}
			}
		case schema.AttributeTypeRequestBody:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					requestBody, diag := hclhelpers.AttributeToString(attr, parseContext.EvalCtx, false)
					if diag != nil && diag.Severity == hcl.DiagError {
						diags = append(diags, diag)
						continue
					}
					p.RequestBody = requestBody
				}
			}
		case schema.AttributeTypeRequestHeaders:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					mapAttrib, diag := hclhelpers.AttributeToMap(attr, parseContext.EvalCtx, false)
					if diag != nil && diag.Severity == hcl.DiagError {
						diags = append(diags, diag)
						continue
					}

					p.RequestHeaders = mapAttrib
				}
			}
		default:
			if !p.IsBaseAttribute(name) {
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

type PipelineStepSleep struct {
	PipelineStepBase
	Duration string `json:"duration"`
}

func (p *PipelineStepSleep) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var durationInput string

	if p.UnresolvedAttributes[schema.AttributeTypeDuration] == nil {
		durationInput = p.Duration
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeDuration], evalContext, &durationInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	return map[string]interface{}{
		schema.AttributeTypeDuration: durationInput,
	}, nil
}

func (p *PipelineStepSleep) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineStepSleep) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDuration:
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
			if !p.IsBaseAttribute(name) {
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

type PipelineStepEmail struct {
	PipelineStepBase
	To string `json:"to"`
}

func (p *PipelineStepEmail) GetError() *PipelineStepError {
	return nil
}

func (p *PipelineStepEmail) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	return map[string]interface{}{
		"to": p.To,
	}, nil
}

func (p *PipelineStepEmail) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeTo:
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

type PipelineStepEcho struct {
	PipelineStepBase
	Text     string   `json:"text"`
	ListText []string `json:"list_text"`
}

func (p *PipelineStepEcho) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var textInput string
	var listTextInput []string

	if p.UnresolvedAttributes[schema.AttributeTypeText] == nil {
		textInput = p.Text
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeText], evalContext, &textInput)
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
		schema.AttributeTypeText: textInput,
		"list_text":              listTextInput,
	}, nil
}

func (p *PipelineStepEcho) GetError() *PipelineStepError {
	return nil
}

func dependsOnFromExpressions(name string, expr hcl.Expression, p IPipelineStep) {
	if len(expr.Variables()) == 0 {
		return
	}
	traversals := expr.Variables()
	for _, traversal := range traversals {
		parts := TraversalAsStringSlice(traversal)
		if len(parts) > 0 {
			if parts[0] == schema.BlockTypePipelineStep {
				dependsOn := parts[1] + "." + parts[2]
				p.AppendDependsOn(dependsOn)
			}
		}
	}
	p.AddUnresolvedAttribute(name, expr)
}

func (p *PipelineStepEcho) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeText:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					val, err := expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse " + schema.AttributeTypeText + " attribute",
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
			if !p.IsBaseAttribute(name) {
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
