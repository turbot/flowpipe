package modconfig

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/perr"
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
	Index      int                  `json:"index" binding:"required"`
	Output     *Output              `json:"output,omitempty"`
	TotalCount int                  `json:"total_count" binding:"required"`
	Each       json.SimpleJSONValue `json:"each"`
}

// Input to the step or pipeline execution
type Input map[string]interface{}

// Output is the output from a step execution.
type Output struct {
	Status string      `json:"status,omitempty"`
	Data   OutputData  `json:"data,omitempty"`
	Errors []StepError `json:"errors,omitempty"`
}

type OutputData map[string]interface{}

func (o *Output) Get(key string) interface{} {
	if o == nil {
		return nil
	}
	return o.Data[key]
}

func (o *Output) Set(key string, value interface{}) {
	if o == nil {
		return
	}
	o.Data[key] = value
}

func (o *Output) HasErrors() bool {
	if o == nil {
		return false
	}

	return o.Errors != nil && len(o.Errors) > 0
}

func (o *Output) AsCtyMap() (map[string]cty.Value, error) {
	if o == nil {
		return map[string]cty.Value{}, nil
	}

	variables := make(map[string]cty.Value)

	for key, value := range o.Data {
		if value == nil {
			continue
		}

		// Check if the value is a Go native data type
		switch v := value.(type) {
		case string, int, float32, float64, int8, int16, int32, int64, bool, []string, []int, []float32, []float64, []int8, []int16, []int32, []int64, []bool:
			ctyType, err := gocty.ImpliedType(v)
			if err != nil {
				return nil, err
			}

			variables[key], err = gocty.ToCtyValue(v, ctyType)
			if err != nil {
				return nil, err
			}
		case []interface{}, map[string]interface{}:
			val, err := hclhelpers.ConvertMapOrSliceToCtyValue(v)
			if err != nil {
				return nil, err
			}
			variables[key] = val
		}

	}

	if o.Errors != nil {
		errList := []cty.Value{}
		for _, stepErr := range o.Errors {
			ctyMap := map[string]cty.Value{}
			var err error
			ctyMap["message"], err = gocty.ToCtyValue(stepErr.Message, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["error_code"], err = gocty.ToCtyValue(stepErr.ErrorCode, cty.Number)
			if err != nil {
				return nil, err
			}
			ctyMap["pipeline_execution_id"], err = gocty.ToCtyValue(stepErr.PipelineExecutionID, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["step_execution_id"], err = gocty.ToCtyValue(stepErr.StepExecutionID, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["pipeline"], err = gocty.ToCtyValue(stepErr.Pipeline, cty.String)
			if err != nil {
				return nil, err
			}
			ctyMap["step"], err = gocty.ToCtyValue(stepErr.Step, cty.String)
			if err != nil {
				return nil, err
			}
			errList = append(errList, cty.ObjectVal(ctyMap))
		}
		variables["errors"] = cty.ListVal(errList)
	}
	return variables, nil
}

type StepError struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	Pipeline            string `json:"pipeline"`
	Step                string `json:"step"`
	Message             string `json:"message"`
	ErrorCode           int    `json:"error_code"`
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
	case schema.BlockTypePipelineStepPipeline:
		s := &PipelineStepPipeline{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s
	case schema.BlockTypePipelineStepFunction:
		s := &PipelineStepFunction{}
		s.UnresolvedAttributes = make(map[string]hcl.Expression)
		step = s

	case schema.BlockTypePipelineStepContainer:
		s := &PipelineStepContainer{}
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
	SetPipelineName(string)
	GetPipelineName() string
	IsResolved() bool
	AddUnresolvedAttribute(string, hcl.Expression)
	GetUnresolvedAttributes() map[string]hcl.Expression
	GetInputs(*hcl.EvalContext) (map[string]interface{}, error)
	GetDependsOn() []string
	AppendDependsOn(...string)
	GetForEach() hcl.Expression
	SetAttributes(hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
	SetErrorConfig(*ErrorConfig)
	GetErrorConfig() *ErrorConfig
	SetOutputConfig(map[string]*PipelineOutput)
	GetOutputConfig() map[string]*PipelineOutput
	Equals(other IPipelineStep) bool
}

type ErrorConfig struct {
	Ignore  bool `json:"ignore"`
	Retries int  `json:"retries"`
}

func (ec *ErrorConfig) Equals(other *ErrorConfig) bool {
	if ec == nil || other == nil {
		return false
	}

	// Compare Ignore
	if ec.Ignore != other.Ignore {
		return false
	}

	// Compare Retries
	if ec.Retries != other.Retries {
		return false
	}

	return true
}

// A common base struct that all pipeline steps must embed
type PipelineStepBase struct {
	Title        *string                    `json:"title,omitempty"`
	Description  *string                    `json:"description,omitempty"`
	Name         string                     `json:"name"`
	Type         string                     `json:"step_type"`
	PipelineName string                     `json:"pipeline_name,omitempty"`
	DependsOn    []string                   `json:"depends_on,omitempty"`
	Resolved     bool                       `json:"resolved,omitempty"`
	ErrorConfig  *ErrorConfig               `json:"-"`
	OutputConfig map[string]*PipelineOutput `json:"-"`

	// This cant' be serialised
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	ForEach              hcl.Expression            `json:"-"`
}

func (p *PipelineStepBase) Equals(otherBase *PipelineStepBase) bool {
	if p == nil || otherBase == nil {
		return false
	}

	// Compare Title
	if !reflect.DeepEqual(p.Title, otherBase.Title) {
		return false
	}

	// Compare Description
	if !reflect.DeepEqual(p.Description, otherBase.Description) {
		return false
	}

	// Compare Name
	if p.Name != otherBase.Name {
		return false
	}

	// Compare Type
	if p.Type != otherBase.Type {
		return false
	}

	// Compare DependsOn slices
	if len(p.DependsOn) != len(otherBase.DependsOn) {
		return false
	}
	for i, dep := range p.DependsOn {
		if dep != otherBase.DependsOn[i] {
			return false
		}
	}

	// Compare Resolved
	if p.Resolved != otherBase.Resolved {
		return false
	}

	// Compare ErrorConfig (if not nil)
	if (p.ErrorConfig == nil && otherBase.ErrorConfig != nil) || (p.ErrorConfig != nil && otherBase.ErrorConfig == nil) {
		return false
	}
	if p.ErrorConfig != nil && otherBase.ErrorConfig != nil && !p.ErrorConfig.Equals(otherBase.ErrorConfig) {
		return false
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(p.UnresolvedAttributes) != len(otherBase.UnresolvedAttributes) {
		return false
	}
	for key, expr := range p.UnresolvedAttributes {
		otherExpr, ok := otherBase.UnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}

		// haven't found a good way to test check equality for two hcl expressions
	}

	// Compare ForEach (if not nil)
	if (p.ForEach == nil && otherBase.ForEach != nil) || (p.ForEach != nil && otherBase.ForEach == nil) {
		return false
	}
	if p.ForEach != nil && otherBase.ForEach != nil && !hclhelpers.ExpressionsEqual(p.ForEach, otherBase.ForEach) {
		return false
	}

	return true
}

func (p *PipelineStepBase) SetPipelineName(pipelineName string) {
	p.PipelineName = pipelineName
}

func (p *PipelineStepBase) GetPipelineName() string {
	return p.PipelineName
}

func (p *PipelineStepBase) SetErrorConfig(errorConfig *ErrorConfig) {
	p.ErrorConfig = errorConfig
}

func (p *PipelineStepBase) GetErrorConfig() *ErrorConfig {
	return p.ErrorConfig
}

func (p *PipelineStepBase) SetOutputConfig(output map[string]*PipelineOutput) {
	p.OutputConfig = output
}

func (p *PipelineStepBase) GetOutputConfig() map[string]*PipelineOutput {
	return p.OutputConfig
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
				Subject:  traversal.SourceRange().Ptr(),
			})
		}
		parts := hclhelpers.TraversalAsStringSlice(traversal)
		if len(parts) < 3 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  constants.BadDependsOn,
				Detail:   "Invalid depends_on format " + strings.Join(parts, "."),
				Subject:  traversal.SourceRange().Ptr(),
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
		title, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			p.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		description, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			p.Description = description
		}
	}

	// if attribute is always unresolved, or at least we treat it to be unresolved. Most of the
	// usage will be testing the value that can only be had during the pipeline execution
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
	return slices.Contains[[]string, string](ValidBaseStepAttributes, name)
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

func (p *PipelineStepHttp) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepHttp)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Compare Url field
	if reflect.DeepEqual(p.Url, other.Url) {
		return false
	}

	// Compare RequestTimeoutMs field
	if reflect.DeepEqual(p.RequestTimeoutMs, other.RequestTimeoutMs) {
		return false
	}

	// Compare Method field
	if reflect.DeepEqual(p.Method, other.Method) {
		return false
	}

	// Compare Insecure field
	if reflect.DeepEqual(p.Insecure, other.Insecure) {
		return false
	}

	// Compare RequestBody field
	if reflect.DeepEqual(p.RequestBody, other.RequestBody) {
		return false
	}

	// Compare RequestHeaders field using deep equality
	if !reflect.DeepEqual(p.RequestHeaders, other.RequestHeaders) {
		return false
	}

	// All fields are equal
	return true

}

func (p *PipelineStepHttp) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var urlInput string
	if p.UnresolvedAttributes[schema.AttributeTypeUrl] == nil {
		if p.Url == nil {
			return nil, perr.InternalWithMessage("Url must be supplied")
		}
		urlInput = *p.Url
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeUrl], evalContext, &urlInput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
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
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
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
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
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
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
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
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
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
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
		inputs[schema.AttributeTypeRequestHeaders] = requestHeaders
	}

	return inputs, nil
}

func (p *PipelineStepHttp) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeUrl:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				urlString, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeUrl + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Url = &urlString
			}
		case schema.AttributeTypeRequestTimeoutMs:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				int64Val, stepDiags := hclhelpers.CtyToInt64(val)
				if stepDiags.HasErrors() {
					diags = append(diags, stepDiags...)
					continue
				}
				p.RequestTimeoutMs = int64Val
			}

		case schema.AttributeTypeMethod:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				method, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeMethod + " attribute to string",
						Subject:  &attr.Range,
					})
				}

				if method != "" {
					if !helpers.StringSliceContains(ValidHttpMethods, strings.ToLower(method)) {
						diags = append(diags, &hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  "Invalid HTTP method: " + method,
							Subject:  &attr.Range,
						})
						continue
					}
					p.Method = &method
				}
			}
		case schema.AttributeTypeInsecure:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				if val.Type() != cty.Bool {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid value for insecure attribute",
						Subject:  &attr.Range,
					})
					continue
				}
				insecure := val.True()
				p.Insecure = &insecure
			}

		case schema.AttributeTypeRequestBody:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				requestBody, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeRequestBody + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.RequestBody = &requestBody
			}

		case schema.AttributeTypeRequestHeaders:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)

			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.RequestHeaders, err = hclhelpers.CtyToGoMapInterface(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse request_headers attribute",
						Subject:  &attr.Range,
					})
					continue
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

func (p *PipelineStepSleep) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepSleep)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	return p.Duration == other.Duration
}

func (p *PipelineStepSleep) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var durationInput string

	if p.UnresolvedAttributes[schema.AttributeTypeDuration] == nil {
		durationInput = p.Duration
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeDuration], evalContext, &durationInput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	return map[string]interface{}{
		schema.AttributeTypeDuration: durationInput,
	}, nil
}

func (p *PipelineStepSleep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeDuration:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				duration, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeDuration + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Duration = duration
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
	To               []string `json:"to"`
	From             *string  `json:"from"`
	SenderCredential *string  `json:"sender_credential"`
	Host             *string  `json:"host"`
	Port             *int64   `json:"port"`
	SenderName       *string  `json:"sender_name"`
	Cc               []string `json:"cc"`
	Bcc              []string `json:"bcc"`
	Body             *string  `json:"body"`
	ContentType      *string  `json:"content_type"`
	Subject          *string  `json:"subject"`
}

func (p *PipelineStepEmail) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepEmail)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Use reflect.DeepEqual to compare slices and pointers
	return reflect.DeepEqual(p.To, other.To) &&
		reflect.DeepEqual(p.From, other.From) &&
		reflect.DeepEqual(p.SenderCredential, other.SenderCredential) &&
		reflect.DeepEqual(p.Host, other.Host) &&
		reflect.DeepEqual(p.Port, other.Port) &&
		reflect.DeepEqual(p.SenderName, other.SenderName) &&
		reflect.DeepEqual(p.Cc, other.Cc) &&
		reflect.DeepEqual(p.Bcc, other.Bcc) &&
		reflect.DeepEqual(p.Body, other.Body) &&
		reflect.DeepEqual(p.ContentType, other.ContentType) &&
		reflect.DeepEqual(p.Subject, other.Subject)

}

func (p *PipelineStepEmail) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var to []string
	if p.UnresolvedAttributes[schema.AttributeTypeTo] == nil {
		to = p.To
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeTo], evalContext, &to)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var from *string
	if p.UnresolvedAttributes[schema.AttributeTypeFrom] == nil {
		from = p.From
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeFrom], evalContext, &from)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var senderCredential *string
	if p.UnresolvedAttributes[schema.AttributeTypeSenderCredential] == nil {
		senderCredential = p.SenderCredential
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSenderCredential], evalContext, &senderCredential)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var host *string
	if p.UnresolvedAttributes[schema.AttributeTypeHost] == nil {
		host = p.Host
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeHost], evalContext, &host)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var port *int64
	if p.UnresolvedAttributes[schema.AttributeTypePort] == nil {
		port = p.Port
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypePort], evalContext, &port)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var senderName *string
	if p.UnresolvedAttributes[schema.AttributeTypeSenderName] == nil {
		senderName = p.SenderName
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSenderName], evalContext, &senderName)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var body *string
	if p.UnresolvedAttributes[schema.AttributeTypeBody] == nil {
		body = p.Body
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeBody], evalContext, &body)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var subject *string
	if p.UnresolvedAttributes[schema.AttributeTypeSubject] == nil {
		subject = p.Subject
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSubject], evalContext, &subject)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var contentType *string
	if p.UnresolvedAttributes[schema.AttributeTypeContentType] == nil {
		contentType = p.ContentType
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeContentType], evalContext, &contentType)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var cc []string
	if p.UnresolvedAttributes[schema.AttributeTypeCc] == nil {
		cc = p.Cc
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCc], evalContext, &cc)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var bcc []string
	if p.UnresolvedAttributes[schema.AttributeTypeBcc] == nil {
		bcc = p.Bcc
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeBcc], evalContext, &bcc)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	results := map[string]interface{}{}

	if to != nil {
		results[schema.AttributeTypeTo] = to
	}

	if from != nil {
		results[schema.AttributeTypeFrom] = *from
	}

	if senderCredential != nil {
		results[schema.AttributeTypeSenderCredential] = *senderCredential
	}

	if host != nil {
		results[schema.AttributeTypeHost] = *host
	}

	if port != nil {
		results[schema.AttributeTypePort] = *port
	}

	if senderName != nil {
		results[schema.AttributeTypeSenderName] = *senderName
	}

	if cc != nil {
		results[schema.AttributeTypeCc] = cc
	}

	if bcc != nil {
		results[schema.AttributeTypeBcc] = bcc
	}

	if body != nil {
		results[schema.AttributeTypeBody] = *body
	}

	if contentType != nil {
		results[schema.AttributeTypeContentType] = *contentType
	}

	if subject != nil {
		results[schema.AttributeTypeSubject] = *subject
	}

	return results, nil
}

func (p *PipelineStepEmail) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeTo:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				emailRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeTo + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.To = emailRecipients
			}

		case schema.AttributeTypeFrom:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				from, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeFrom + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.From = &from
			}

		case schema.AttributeTypeSenderCredential:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				senderCredential, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSenderCredential + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SenderCredential = &senderCredential
			}

		case schema.AttributeTypeHost:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				host, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeHost + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Host = &host
			}

		case schema.AttributeTypePort:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				port, ctyDiags := hclhelpers.CtyToInt64(val)
				if ctyDiags.HasErrors() {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to convert port into integer",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Port = port
			}

		case schema.AttributeTypeSenderName:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				senderName, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSenderName + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.SenderName = &senderName
			}

		case schema.AttributeTypeCc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				ccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeCc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Cc = ccRecipients
			}

		case schema.AttributeTypeBcc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				bccRecipients, ctyErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBcc + " attribute to string slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.Bcc = bccRecipients
			}

		case schema.AttributeTypeBody:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				body, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeBody + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Body = &body
			}

		case schema.AttributeTypeContentType:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				contentType, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeContentType + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.ContentType = &contentType
			}

		case schema.AttributeTypeSubject:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				subject, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSubject + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Subject = &subject
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Email Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

type PipelineStepEcho struct {
	PipelineStepBase
	Text    string               `json:"text"`
	Numeric float64              `json:"numeric"`
	Json    json.SimpleJSONValue `json:"json"`
}

func (p *PipelineStepEcho) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepEcho)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	if p.Text != other.Text {
		return false
	}

	// TODO: json test?
	return true
}

func (p *PipelineStepEcho) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var textInput string

	if p.UnresolvedAttributes[schema.AttributeTypeText] == nil {
		textInput = p.Text
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeText], evalContext, &textInput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
	}

	var numericInput float64
	if p.UnresolvedAttributes[schema.AttributeTypeNumeric] == nil {
		numericInput = p.Numeric
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeNumeric], evalContext, &numericInput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
	}

	var jsonInput json.SimpleJSONValue
	if p.UnresolvedAttributes[schema.AttributeTypeJson] == nil {
		jsonInput = p.Json
	} else {
		var ctyOutput cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeJson], evalContext, &ctyOutput)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
		jsonInput = json.SimpleJSONValue{Value: ctyOutput}
	}

	return map[string]interface{}{
		schema.AttributeTypeText:    textInput,
		schema.AttributeTypeJson:    jsonInput,
		schema.AttributeTypeNumeric: numericInput,
	}, nil
}

func dependsOnFromExpressions(attr *hcl.Attribute, evalContext *hcl.EvalContext, p IPipelineStep) (cty.Value, hcl.Diagnostics) {
	expr := attr.Expr
	// resolve it first if we can
	val, stepDiags := expr.Value(evalContext)
	if stepDiags != nil && stepDiags.HasErrors() {
		resolvedDiags := 0
		for _, e := range stepDiags {
			if e.Severity == hcl.DiagError {
				if e.Detail == `There is no variable named "step".` {
					traversals := expr.Variables()
					dependsOnAdded := false
					for _, traversal := range traversals {
						parts := hclhelpers.TraversalAsStringSlice(traversal)
						if len(parts) > 0 {
							// When the expression/traversal is referencing an index, the index is also included in the parts
							// for example: []string len: 5, cap: 5, ["step","sleep","sleep_1","0","duration"]
							if parts[0] == schema.BlockTypePipelineStep {
								dependsOn := parts[1] + "." + parts[2]
								p.AppendDependsOn(dependsOn)
								dependsOnAdded = true
							}
						}
					}
					if dependsOnAdded {
						resolvedDiags++
					}
				} else if e.Detail == `There is no variable named "each".` || e.Detail == `There is no variable named "param".` {
					resolvedDiags++
				} else {
					return cty.NilVal, stepDiags
				}
			}
		}

		// check if all diags have been resolved
		if resolvedDiags == len(stepDiags) {

			// * Don't forget to add this, if you change the logic ensure that the code flow still
			// * calls AddUnresolvedAttribute
			p.AddUnresolvedAttribute(attr.Name, expr)
			return cty.NilVal, hcl.Diagnostics{}
		} else {
			// There's an error here
			return cty.NilVal, stepDiags
		}
	}

	return val, hcl.Diagnostics{}
}

func (p *PipelineStepEcho) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeText:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				text, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeText + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Text = text
			}
		case schema.AttributeTypeJson:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}
			if val != cty.NilVal {
				p.Json = json.SimpleJSONValue{Value: val}
			}

		case schema.AttributeTypeNumeric:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val != cty.NilVal {
				p.Numeric, _ = val.AsBigFloat().Float64()
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

func (p *PipelineStepQuery) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepQuery)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	if len(p.Args) != len(other.Args) {
		return false
	}
	for i := range p.Args {
		if p.Args[i] != other.Args[i] {
			return false
		}
	}

	return reflect.DeepEqual(p.ConnnectionString, other.ConnnectionString) &&
		reflect.DeepEqual(p.Sql, other.Sql)
}

func (p *PipelineStepQuery) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var sql *string
	if p.UnresolvedAttributes[schema.AttributeTypeSql] == nil {
		if p.Sql == nil {
			return nil, perr.BadRequestWithMessage("sql must be supplied")
		}
		sql = p.Sql
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSql], evalContext, &sql)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	var connectionString *string
	if p.UnresolvedAttributes[schema.AttributeTypeConnectionString] == nil {
		if p.ConnnectionString == nil {
			return nil, perr.BadRequestWithMessage("connection string must be supplied")
		}
		connectionString = p.ConnnectionString
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeConnectionString], evalContext, &connectionString)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
	}

	results := map[string]interface{}{}

	if sql != nil {
		results[schema.AttributeTypeSql] = *sql
	}

	if connectionString != nil {
		results[schema.AttributeTypeConnectionString] = *connectionString
	}

	if p.UnresolvedAttributes[schema.AttributeTypeArgs] != nil {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeArgs], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}

		mapValue, err := hclhelpers.CtyToGoMapInterface(args)
		if err != nil {
			return nil, err
		}
		results[schema.AttributeTypeArgs] = mapValue

	} else if p.Args != nil {
		results[schema.AttributeTypeArgs] = p.Args
	}

	return results, nil
}

func (p *PipelineStepQuery) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSql:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				sql, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSql + " attribute to string",
						Subject:  &attr.Range,
					})
				}
				p.Sql = &sql
			}
		case schema.AttributeTypeConnectionString:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				connectionString := val.AsString()
				p.ConnnectionString = &connectionString
			}
		case schema.AttributeTypeArgs:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVals, err2 := hclhelpers.CtyToGoInterfaceSlice(val)
				if err2 != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeArgs + "' attribute to Go values",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Args = goVals
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Echo Step '" + attr.Name + "'",
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

type PipelineStepPipeline struct {
	PipelineStepBase

	Pipeline cty.Value `json:"-"`
	Args     Input     `json:"args"`
}

func (p *PipelineStepPipeline) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepPipeline)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	// Check if the maps have the same number of elements
	if len(p.Args) != len(other.Args) {
		return false
	}

	// Iterate through the first map
	for key, value1 := range p.Args {
		// Check if the key exists in the second map
		value2, ok := other.Args[key]
		if !ok {
			return false
		}

		// Use reflect.DeepEqual to compare the values
		if !reflect.DeepEqual(value1, value2) {
			return false
		}
	}

	// TODO: more here, can't just compare the name
	return p.Pipeline.AsValueMap()[schema.LabelName] == other.Pipeline.AsValueMap()[schema.LabelName]

}

func (p *PipelineStepPipeline) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var pipeline string
	if p.UnresolvedAttributes[schema.AttributeTypePipeline] == nil {
		if p.Pipeline == cty.NilVal {
			return nil, perr.InternalWithMessage("Pipeline must be supplied")
		}
		valueMap := p.Pipeline.AsValueMap()
		pipelineNameCty := valueMap[schema.LabelName]
		pipeline = pipelineNameCty.AsString()

	} else {
		var pipelineCty cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypePipeline], evalContext, &pipelineCty)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
		valueMap := pipelineCty.AsValueMap()
		pipelineNameCty := valueMap[schema.LabelName]
		pipeline = pipelineNameCty.AsString()
	}

	results := map[string]interface{}{}

	results[schema.AttributeTypePipeline] = pipeline

	if p.UnresolvedAttributes[schema.AttributeTypeArgs] != nil {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeArgs], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError(schema.BlockTypePipelineStep, diags)
		}

		mapValue, err := hclhelpers.CtyToGoMapInterface(args)
		if err != nil {
			return nil, err
		}
		results[schema.AttributeTypeArgs] = mapValue

	} else if p.Args != nil {
		results[schema.AttributeTypeArgs] = p.Args
	}

	return results, nil
}

func (p *PipelineStepPipeline) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypePipeline:
			expr := attr.Expr
			if attr.Expr != nil {
				val, err := expr.Value(evalContext)
				if err != nil {
					// For Step's Pipeline reference, all it needs is the pipeline. It can't possibly use the output of a pipeline
					// so if the Pipeline is not parsed (yet) then the error message is:
					// Summary: "Unknown variable"
					// Detail: "There is no variable named \"pipeline\"."
					//
					// Do not unpack the error and create a new "Diagnostic", leave the original error message in
					// and let the "Mod processing" determine if there's an unresolved block
					//
					// There's no "depends_on" from the step to the pipeline, the Flowpipe ES engine does not require it
					diags = append(diags, err...)

					return diags
				}
				p.Pipeline = val
			}
		case schema.AttributeTypeArgs:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				goVals, err2 := hclhelpers.CtyToGoMapInterface(val)
				if err2 != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeArgs + " attribute to Go values",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Args = goVals
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

type PipelineStepFunction struct {
	PipelineStepBase

	Function cty.Value `json:"-"`

	Runtime string `json:"runtime" cty:"runtime"`
	Src     string `json:"src" cty:"src"`
	Handler string `json:"handler" cty:"handler"`

	Event map[string]interface{} `json:"event"`
	Env   map[string]string      `json:"env"`
}

func (p *PipelineStepFunction) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepFunction)
	if !ok {
		return false
	}

	// TODO: more here, can't just compare the name
	return p.Function.AsValueMap()[schema.LabelName] == other.Function.AsValueMap()[schema.LabelName]
}

func (p *PipelineStepFunction) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {

	var env map[string]string
	if p.UnresolvedAttributes[schema.AttributeTypeEnv] == nil {
		env = p.Env
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEnv], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}

		var err error
		env, err = hclhelpers.CtyToGoMapString(args)
		if err != nil {
			return nil, err
		}
	}

	var event map[string]interface{}
	if p.UnresolvedAttributes[schema.AttributeTypeEvent] == nil {
		event = p.Event
	} else {
		val, diags := p.UnresolvedAttributes[schema.AttributeTypeEvent].Value(evalContext)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}

		var err error
		event, err = hclhelpers.CtyToGoMapInterface(val)
		if err != nil {
			return nil, err
		}
	}

	var src string
	if p.UnresolvedAttributes[schema.AttributeTypeSrc] == nil {
		src = p.Src
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSrc], evalContext, &src)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
	}

	var runtime string
	if p.UnresolvedAttributes[schema.AttributeTypeRuntime] == nil {
		runtime = p.Runtime
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSrc], evalContext, &runtime)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
	}

	var handler string
	if p.UnresolvedAttributes[schema.AttributeTypeHandler] == nil {
		handler = p.Handler
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeSrc], evalContext, &handler)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
	}

	return map[string]interface{}{
		schema.LabelName:            p.PipelineName + "." + p.GetFullyQualifiedName(),
		schema.AttributeTypeSrc:     src,
		schema.AttributeTypeRuntime: runtime,
		schema.AttributeTypeHandler: handler,
		schema.AttributeTypeEvent:   event,
		schema.AttributeTypeEnv:     env,
	}, nil
}

func (p *PipelineStepFunction) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSrc:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Src = val.AsString()
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
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Function Step: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

type PipelineStepContainer struct {
	PipelineStepBase

	Image string            `json:"image"`
	Cmd   []string          `json:"cmd"`
	Env   map[string]string `json:"env"`
}

func (p *PipelineStepContainer) Equals(iOther IPipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && iOther == nil {
		return true
	}

	other, ok := iOther.(*PipelineStepContainer)
	if !ok {
		return false
	}

	return p.Image == other.Image && reflect.DeepEqual(p.Cmd, other.Cmd) && reflect.DeepEqual(p.Env, other.Env)
}

func (p *PipelineStepContainer) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	var image string
	if p.UnresolvedAttributes[schema.AttributeTypeImage] == nil {
		image = p.Image
	} else {
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeImage], evalContext, &image)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}
	}

	var cmd []string
	if p.UnresolvedAttributes[schema.AttributeTypeCmd] == nil {
		cmd = p.Cmd
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeCmd], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}

		var err error
		cmd, err = hclhelpers.CtyToGoStringSlice(args, args.Type())
		if err != nil {
			return nil, err
		}
	}

	var env map[string]string
	if p.UnresolvedAttributes[schema.AttributeTypeEnv] == nil {
		env = p.Env
	} else {
		var args cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeEnv], evalContext, &args)
		if diags.HasErrors() {
			return nil, error_helpers.HclDiagsToError("step", diags)
		}

		var err error
		env, err = hclhelpers.CtyToGoMapString(args)
		if err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		schema.LabelName:          p.Name,
		schema.AttributeTypeImage: image,
		schema.AttributeTypeCmd:   cmd,
		schema.AttributeTypeEnv:   env,
	}, nil
}

func (p *PipelineStepContainer) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeImage:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				p.Image = val.AsString()
			}
		case schema.AttributeTypeCmd:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				cmds, moreErr := hclhelpers.CtyToGoStringSlice(val, val.Type())
				if moreErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse '" + schema.AttributeTypeCmd + "' attribute to string slice",
						Subject:  &attr.Range,
					})
					continue
				}
				p.Cmd = cmds
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
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Function Step: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}
