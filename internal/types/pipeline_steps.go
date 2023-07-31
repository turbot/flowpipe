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
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/terraform-components/addrs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
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

type OutputVariables map[string]interface{}

// StepOutput is the output from a pipeline.
type StepOutput struct {
	Status          string          `json:"status"`
	OutputVariables OutputVariables `json:"output_variables,omitempty"`
	Errors          *StepErrors     `json:"errors,omitempty"`
}

func (o *StepOutput) Get(key string) interface{} {
	if o == nil {
		return nil
	}
	return o.OutputVariables[key]
}

func (o *StepOutput) Set(key string, value interface{}) {
	if o == nil {
		return
	}
	o.OutputVariables[key] = value
}

func (o *StepOutput) HasErrors() bool {
	if o == nil {
		return false
	}

	return o.Errors != nil && len(*o.Errors) > 0
}

func (o *StepOutput) AsHclVariables() (cty.Value, error) {
	if o == nil {
		return cty.ObjectVal(map[string]cty.Value{}), nil
	}

	variables := make(map[string]cty.Value)

	for key, value := range o.OutputVariables {
		if value == nil {
			continue
		}

		// Check if the value is a Go native data type
		switch v := value.(type) {
		case string, int, float32, float64, int8, int16, int32, int64, bool, []string, []int, []float32, []float64, []int8, []int16, []int32, []int64, []bool:
			ctyType, err := gocty.ImpliedType(v)
			if err != nil {
				return cty.NilVal, err
			}

			variables[key], err = gocty.ToCtyValue(v, ctyType)
			if err != nil {
				return cty.NilVal, err
			}
		case []interface{}, map[string]interface{}:
			val, err := hclhelpers.ConvertMapOrSliceToCtyValue(v)
			if err != nil {
				return cty.NilVal, err
			}
			variables[key] = val
		}

	}

	if o.Errors != nil {
		errList := []cty.Value{}
		for _, stepErr := range *o.Errors {
			ctyMap := map[string]cty.Value{}
			var err error
			ctyMap["message"], err = gocty.ToCtyValue(stepErr.Message, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			ctyMap["error_code"], err = gocty.ToCtyValue(stepErr.ErrorCode, cty.Number)
			if err != nil {
				return cty.NilVal, err
			}
			ctyMap["pipeline_execution_id"], err = gocty.ToCtyValue(stepErr.PipelineExecutionID, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			ctyMap["step_execution_id"], err = gocty.ToCtyValue(stepErr.StepExecutionID, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			ctyMap["pipeline"], err = gocty.ToCtyValue(stepErr.Pipeline, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			ctyMap["step"], err = gocty.ToCtyValue(stepErr.Step, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			errList = append(errList, cty.ObjectVal(ctyMap))
		}
		variables["errors"] = cty.ListVal(errList)
	}
	return cty.ObjectVal(variables), nil
}

type StepError struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	Pipeline            string `json:"pipeline"`
	Step                string `json:"step"`
	Message             string `json:"message"`
	ErrorCode           int    `json:"error_code"`
}

type StepErrors []StepError

func (s *StepErrors) Add(err StepError) {
	*s = append(*s, err)
}

type NextStepAction string

const (
	// Default Next Step action which is just to start them, note that
	// the step may yet be "skipped" if the IF clause is preventing the step
	// to actually start, but at the very least we can "start" the step.
	NextStepActionStart NextStepAction = "start"

	// This happens if the step can't be started because one of it's dependency as failed
	NextStepActionInaccessible NextStepAction = "inaccessible"

	NextStepActionSkip NextStepAction = "skip"
)

type NextStep struct {
	StepName string         `json:"step_name"`
	DelayMs  int            `json:"delay_ms,omitempty"`
	Action   NextStepAction `json:"action"`
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
	case schema.BlockTypePipelineStepQuery:
		s := &PipelineStepQuery{}
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
	SetAttributes(hcl.Attributes, *pipeparser.ParseContext) hcl.Diagnostics
	SetErrorConfig(*ErrorConfig)
	GetErrorConfig() *ErrorConfig
}

type ErrorConfig struct {
	Ignore  bool `json:"ignore"`
	Retries int  `json:"retries"`
}

// A common base struct that all pipeline steps must embed
type PipelineStepBase struct {
	Title       *string      `json:"title,omitempty"`
	Description *string      `json:"description,omitempty"`
	Name        string       `json:"name"`
	Type        string       `json:"step_type"`
	DependsOn   []string     `json:"depends_on,omitempty"`
	Resolved    bool         `json:"resolved,omitempty"`
	ErrorConfig *ErrorConfig `json:"error_configs,omitempty"`

	// This cant' be serialised
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	ForEach              hcl.Expression            `json:"-"`
}

func (p *PipelineStepBase) SetErrorConfig(errorConfig *ErrorConfig) {
	p.ErrorConfig = errorConfig
}

func (p *PipelineStepBase) GetErrorConfig() *ErrorConfig {
	return p.ErrorConfig
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
		parts := hclhelpers.TraversalAsStringSlice(traversal)
		if len(parts) < 3 {
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

		do, dgs := hclhelpers.ExpressionToDepends(attr.Expr, ValidDependsOnTypes)
		diags = append(diags, dgs...)
		dependsOn = append(dependsOn, do...)
	}

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

	if attr, exists := hclAttributes[schema.AttributeTypeIf]; exists {
		// If is always treated as an unresolved attribute
		p.AddUnresolvedAttribute(schema.AttributeTypeIf, attr.Expr)

		do, dgs := hclhelpers.ExpressionToDepends(attr.Expr, ValidDependsOnTypes)
		diags = append(diags, dgs...)
		dependsOn = append(dependsOn, do...)
	}

	p.DependsOn = append(p.DependsOn, dependsOn...)

	return diags
}

var ValidBaseStepAttributes = []string{
	schema.AttributeTypeTitle,
	schema.AttributeTypeDescription,
	schema.AttributeTypeDependsOn,
	schema.AttributeTypeForEach,
	schema.AttributeTypeIf,
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

	if p.UnresolvedAttributes[schema.AttributeTypeMethod] == nil {
		if p.Method != nil {
			inputs[schema.AttributeTypeMethod] = *p.Method
		}
	} else {
		var method string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeMethod], evalContext, &method)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
		inputs[schema.AttributeTypeMethod] = strings.ToLower(method)
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestTimeoutMs] == nil {
		if p.RequestTimeoutMs != nil {
			inputs[schema.AttributeTypeRequestTimeoutMs] = *p.RequestTimeoutMs
		}
	} else {
		var timeoutMs int64
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestTimeoutMs], evalContext, &timeoutMs)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
		inputs[schema.AttributeTypeRequestTimeoutMs] = timeoutMs
	}

	if p.UnresolvedAttributes[schema.AttributeTypeInsecure] == nil {
		if p.Insecure != nil {
			inputs[schema.AttributeTypeInsecure] = *p.Insecure
		}
	} else {
		var insecure bool
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeInsecure], evalContext, &insecure)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
		inputs[schema.AttributeTypeInsecure] = insecure
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestBody] == nil {
		if p.RequestBody != nil {
			inputs[schema.AttributeTypeRequestBody] = *p.RequestBody
		}
	} else {
		var requestBody string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestBody], evalContext, &requestBody)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
		inputs[schema.AttributeTypeRequestBody] = requestBody
	}

	if p.UnresolvedAttributes[schema.AttributeTypeRequestHeaders] == nil {
		if p.RequestHeaders != nil {
			inputs[schema.AttributeTypeRequestHeaders] = p.RequestHeaders
		}
	} else {
		var requestHeaders map[string]string
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeRequestHeaders], evalContext, &requestHeaders)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
		inputs[schema.AttributeTypeRequestHeaders] = requestHeaders
	}

	return inputs, nil
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
	Text    string               `json:"text"`
	Json    json.SimpleJSONValue `json:"json"`
	Dynamic cty.Value            `json:"dynamic"`
}

func (p *PipelineStepEcho) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var textInput string

	if p.UnresolvedAttributes[schema.AttributeTypeText] == nil {
		textInput = p.Text
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeText], evalContext, &textInput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError("step", diags)
		}
	}

	var jsonInput json.SimpleJSONValue
	if p.UnresolvedAttributes[schema.AttributeTypeJson] == nil {
		jsonInput = p.Json
	} else {
		var ctyOutput cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeJson], evalContext, &ctyOutput)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError("step", diags)
		}
		jsonInput = json.SimpleJSONValue{Value: ctyOutput}
	}

	return map[string]interface{}{
		schema.AttributeTypeText: textInput,
		schema.AttributeTypeJson: jsonInput,
	}, nil
}

func dependsOnFromExpressions(name string, expr hcl.Expression, p IPipelineStep) {
	if len(expr.Variables()) == 0 {
		return
	}
	traversals := expr.Variables()
	for _, traversal := range traversals {
		parts := hclhelpers.TraversalAsStringSlice(traversal)
		if len(parts) > 0 {
			// When the expression/traversal is referencing an index, the index is also included in the parts
			// for example: []string len: 5, cap: 5, ["step","sleep","sleep_1","0","duration"]
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
		case schema.AttributeTypeJson:
			expr := attr.Expr
			if len(expr.Variables()) > 0 {
				dependsOnFromExpressions(name, expr, p)
			} else {
				val, err := expr.Value(parseContext.EvalCtx)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeJson + " attribute",
						Subject:  &attr.Range,
					})
					continue
				}

				p.Json = json.SimpleJSONValue{Value: val}
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

type PipelineStepQuery struct {
	PipelineStepBase
	ConnnectionString *string       `json:"connection_string"`
	Sql               *string       `json:"sql"`
	Args              []interface{} `json:"args"`
}

func (p *PipelineStepQuery) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var sql *string
	if p.UnresolvedAttributes[schema.AttributeTypeSql] == nil {
		if p.Sql == nil {
			return nil, fperr.InternalWithMessage("Url must be supplied")
		}
		sql = p.Sql
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeUrl], evalContext, &sql)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var connectionString *string
	if p.UnresolvedAttributes[schema.AttributeTypeConnectionString] == nil {
		if p.ConnnectionString == nil {
			return nil, fperr.InternalWithMessage("Url must be supplied")
		}
		connectionString = p.ConnnectionString
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeConnectionString], evalContext, &connectionString)
		if diags.HasErrors() {
			return nil, pipeparser.DiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	results := map[string]interface{}{}

	if sql != nil {
		results[schema.AttributeTypeSql] = *sql
	}

	if connectionString != nil {
		results[schema.AttributeTypeConnectionString] = *connectionString
	}

	if p.Args != nil {
		results[schema.AttributeTypeArgs] = p.Args
	}

	return results, nil
}

func (p *PipelineStepQuery) SetAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSql:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					val, err := expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse " + schema.AttributeTypeSql + " attribute",
							Subject:  &attr.Range,
						})
						continue
					}

					sql := val.AsString()
					p.Sql = &sql
				}
			}
		case schema.AttributeTypeConnectionString:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					val, err := expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse " + schema.AttributeTypeConnectionString + " attribute",
							Subject:  &attr.Range,
						})
						continue
					}

					connectionString := val.AsString()
					p.ConnnectionString = &connectionString
				}
			}
		case schema.AttributeTypeArgs:
			if attr.Expr != nil {
				expr := attr.Expr
				if len(expr.Variables()) > 0 {
					dependsOnFromExpressions(name, expr, p)
				} else {
					vals, err := expr.Value(parseContext.EvalCtx)
					if err != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse " + schema.AttributeTypeSql + " attribute",
							Subject:  &attr.Range,
						})
						continue
					}
					goVals, err2 := hclhelpers.CtyToGoInterfaceSlice(vals)
					if err2 != nil {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Unable to parse " + schema.AttributeTypeSql + " attribute to Go values",
							Subject:  &attr.Range,
						})
						continue
					}
					p.Args = goVals
				}
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
