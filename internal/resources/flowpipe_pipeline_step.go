package resources

import (
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/terraform-components/addrs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

type StepForEach struct {
	ForEachStep bool                 `json:"for_each_step"`
	Key         string  `json:"key"  binding:"required"`
	Output      *Output `json:"output,omitempty"`
	TotalCount  int     `json:"total_count" binding:"required"`
	Each        json.SimpleJSONValue `json:"each" swaggerignore:"true"`
}

type StepLoop struct {
	Index         int    `json:"index" binding:"required"`
	Input         *Input `json:"input,omitempty"`
	LoopCompleted bool   `json:"loop_completed"`
}

type StepRetry struct {
	Count          int    `json:"count" binding:"required"`
	Input          *Input `json:"input,omitempty"`
	RetryCompleted bool   `json:"retry_completed"`
}

// Input to the step or pipeline execution
type Input map[string]interface{}

func (i *Input) AsCtyMap() (map[string]cty.Value, error) {
	if i == nil {
		return map[string]cty.Value{}, nil
	}

	variables := make(map[string]cty.Value)

	for key, value := range *i {
		if value == nil || key == "step_name" {
			continue
		}

		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(value)
		if err != nil {
			return nil, err
		}

		variables[key] = ctyVal
	}

	return variables, nil
}

// Output is the output from a step execution.
type Output struct {
	Status      string      `json:"status,omitempty"`
	FailureMode string      `json:"failure_mode,omitempty"`
	Data        OutputData  `json:"data,omitempty"`
	Errors      []StepError `json:"errors,omitempty"`
	// Flowpipe metadata, contains started_at, finished_at
	Flowpipe map[string]interface{} `json:"flowpipe,omitempty"`
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

	return len(o.Errors) > 0
}

func (o *Output) flowpipeMetadataCtyMap() (map[string]cty.Value, error) {
	if o == nil {
		return nil, nil
	}

	variables := make(map[string]cty.Value)

	var err error
	variables[schema.AttributeTypeStartedAt], err = hclhelpers.ConvertInterfaceToCtyValue(o.Flowpipe[schema.AttributeTypeStartedAt])
	if err != nil {
		return nil, err
	}

	variables[schema.AttributeTypeFinishedAt], err = hclhelpers.ConvertInterfaceToCtyValue(o.Flowpipe[schema.AttributeTypeFinishedAt])
	if err != nil {
		return nil, err
	}

	return variables, nil
}

func (o *Output) AsCtyMap() (map[string]cty.Value, error) {
	if o == nil {
		return map[string]cty.Value{}, nil
	}

	variables := make(map[string]cty.Value)

	// "native" primitive output (not a configured/declared output from the output block)
	for key, value := range o.Data {
		if value == nil {
			continue
		}

		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(value)
		if err != nil {
			return nil, err
		}

		variables[key] = ctyVal
	}

	// errors
	if o.Errors != nil {
		errList := []cty.Value{}
		for _, stepErr := range o.Errors {
			ctyMap := map[string]cty.Value{}
			var err error
			errorAttributes := map[string]cty.Type{
				"instance": cty.String,
				"detail":   cty.String,
				"type":     cty.String,
				"title":    cty.String,
				"status":   cty.Number,
			}

			errorObject := map[string]interface{}{
				"instance": stepErr.Error.Instance,
				"detail":   stepErr.Error.Detail,
				"type":     stepErr.Error.Type,
				"title":    stepErr.Error.Title,
				"status":   stepErr.Error.Status,
			}
			ctyMap["error"], err = gocty.ToCtyValue(errorObject, cty.Object(errorAttributes))
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

	// flowpipe metadata
	fpMetadata, err := o.flowpipeMetadataCtyMap()
	if err != nil {
		return nil, err
	}

	variables[schema.AttributeTypeFlowpipe] = cty.ObjectVal(fpMetadata)

	return variables, nil
}

type StepError struct {
	PipelineExecutionID string          `json:"pipeline_execution_id"`
	StepExecutionID     string          `json:"step_execution_id"`
	Pipeline            string          `json:"pipeline"`
	Step                string          `json:"step"`
	Error               perr.ErrorModel `json:"error"`
}

type NextStepAction string

const (
	// Default Next Step action which is just to start them, note that
	// the step may yet be "skipped" if the IF clause is preventing the step
	// to actually start, but at the very least we can "start" the step.
	NextStepActionStart NextStepAction = "start"

	// This happens if the step can't be started because one of it's dependency as failed
	//
	// Q: So why would step failure does not mean pipeline fail straight away?
	// A: We can't raise the pipeline fail command if there's "ignore error" directive on the step.
	//    If there are steps that depend on the failed step, these steps becomes "inaccessible", they can't start
	//    because the prerequisites have failed.
	//
	NextStepActionInaccessible NextStepAction = "inaccessible"

	NextStepActionSkip NextStepAction = "skip"
)

type NextStep struct {
	StepName       string         `json:"step_name"`
	Action         NextStepAction `json:"action"`
	StepForEach    *StepForEach   `json:"step_for_each,omitempty"`
	StepLoop       *StepLoop      `json:"step_loop,omitempty"`
	Input          Input          `json:"input"`
	MaxConcurrency *int           `json:"max_concurrency,omitempty"`
}

func NewPipelineStep(stepType, stepName string, pipeline *Pipeline) PipelineStep {
	var step PipelineStep
	switch stepType {
	case schema.BlockTypePipelineStepHttp:
		step = &PipelineStepHttp{}
	case schema.BlockTypePipelineStepSleep:
		step = &PipelineStepSleep{}
	case schema.BlockTypePipelineStepEmail:
		step = &PipelineStepEmail{}
	case schema.BlockTypePipelineStepTransform:
		step = &PipelineStepTransform{}
	case schema.BlockTypePipelineStepQuery:
		step = &PipelineStepQuery{}
	case schema.BlockTypePipelineStepPipeline:
		step = &PipelineStepPipeline{}
	case schema.BlockTypePipelineStepFunction:
		step = &PipelineStepFunction{}
	case schema.BlockTypePipelineStepContainer:
		step = &PipelineStepContainer{}
	case schema.BlockTypePipelineStepInput:
		step = &PipelineStepInput{}
	case schema.BlockTypePipelineStepMessage:
		step = &PipelineStepMessage{}
	default:
		return nil
	}

	step.Initialize()
	step.SetName(stepName)
	step.SetType(stepType)
	step.SetPipeline(pipeline)

	return step
}

// A common interface that all pipeline steps must implement
type PipelineStep interface {
	PipelineStepBaseInterface

	Initialize()
	GetFullyQualifiedName() string
	GetName() string
	SetName(string)
	GetType() string
	SetType(string)
	SetPipelineName(string)
	SetPipeline(*Pipeline)
	GetPipeline() *Pipeline
	GetPipelineName() string
	IsResolved() bool
	AddUnresolvedBody(string, hcl.Body)

	GetInputs(*hcl.EvalContext) (map[string]interface{}, error)
	GetInputs2(*hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error)

	GetDependsOn() []string
	GetCredentialDependsOn() []string
	GetConnectionDependsOn() []string
	GetForEach() hcl.Expression
	SetAttributes(hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
	GetUnresolvedAttributes() map[string]hcl.Expression
	SetBlockConfig(hcl.Blocks, *hcl.EvalContext) hcl.Diagnostics
	GetErrorConfig(*hcl.EvalContext, bool) (*ErrorConfig, hcl.Diagnostics)
	GetRetryConfig(*hcl.EvalContext, bool) (*RetryConfig, hcl.Diagnostics)
	GetLoopConfig() LoopDefn
	GetThrowConfig() []*ThrowConfig
	SetOutputConfig(map[string]*PipelineOutput)
	GetOutputConfig() map[string]*PipelineOutput
	Equals(other PipelineStep) bool
	Validate() hcl.Diagnostics
	SetFileReference(fileName string, startLineNumber int, endLineNumber int)
	SetRange(*hcl.Range)
	GetRange() *hcl.Range
	GetMaxConcurrency(*hcl.EvalContext) *int
}

type PipelineStepBaseInterface interface {
	AppendDependsOn(...string)
	AppendCredentialDependsOn(...string)
	AppendConnectionDependsOn(...string)
	AddUnresolvedAttribute(string, hcl.Expression)
	GetPipeline() *Pipeline
}

type ConnectionDependency struct {
	Source string
	Type   string
}

// A common base struct that all pipeline steps must embed
type PipelineStepBase struct {
	Title               *string        `json:"title,omitempty"`
	Description         *string        `json:"description,omitempty"`
	Name                string         `json:"name"`
	Type                string         `json:"step_type"`
	PipelineName        string         `json:"pipeline_name,omitempty"`
	Pipeline            *Pipeline      `json:"-"`
	Timeout             interface{}    `json:"timeout,omitempty"`
	DependsOn           []string       `json:"depends_on,omitempty"`
	CredentialDependsOn []string       `json:"credential_depends_on,omitempty"`
	ConnectionDependsOn []string       `json:"connection_depends_on,omitempty"`
	Resolved            bool           `json:"resolved,omitempty"`
	ErrorConfig         *ErrorConfig   `json:"-"`
	RetryConfig         *RetryConfig   `json:"retry,omitempty"`
	ThrowConfig         []*ThrowConfig `json:"throw,omitempty"`
	// TODO: we should serialise this, it's used in PipelineLoaded event to have a record the exact pipeline config loaded. There's no further need apart from record keeping, so it's OK to have it unserializeable for now.
	LoopConfig   LoopDefn                   `json:"-"`
	OutputConfig map[string]*PipelineOutput `json:"-"`
	FileName     string                     `json:"file_name"`
	StartLineNumber int                        `json:"start_line_number"`
	EndLineNumber   int                        `json:"end_line_number"`
	MaxConcurrency  *int                       `json:"max_concurrency,omitempty"`
	Range           *hcl.Range                 `json:"range"`

	// This cant' be serialised
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	UnresolvedBodies     map[string]hcl.Body       `json:"-"`
	ForEach              hcl.Expression            `json:"-"`
}

func (p *PipelineStepBase) Initialize() {
	p.UnresolvedAttributes = make(map[string]hcl.Expression)
	p.UnresolvedBodies = make(map[string]hcl.Body)
}

func (p *PipelineStepBase) GetLoopConfig() LoopDefn {
	return p.LoopConfig
}

func (p *PipelineStepBase) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	p.FileName = fileName
	p.StartLineNumber = startLineNumber
	p.EndLineNumber = endLineNumber
}

func (p *PipelineStepBase) SetRange(r *hcl.Range) {
	p.Range = r
}

func (p *PipelineStepBase) GetRange() *hcl.Range {
	return p.Range
}

func (p *PipelineStepBase) GetInputs2(*hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {
	return nil, nil, nil
}

func (p *PipelineStepBase) GetRetryConfig(evalContext *hcl.EvalContext, ifResolution bool) (*RetryConfig, hcl.Diagnostics) {

	if p.RetryConfig == nil {
		return nil, hcl.Diagnostics{}
	}

	if !ifResolution {
		return p.RetryConfig, hcl.Diagnostics{}
	}

	// do not modify the existing retry config, it should always be resolved at runtime
	newRetryConfig := &RetryConfig{}

	if p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeIf] != nil {
		ifValue, diags := p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeIf].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		// If the `if` attribute returns "false" then we return nil for the retry config, thus we won't be retrying it
		if !ifValue.True() {
			return nil, hcl.Diagnostics{}
		}
	}

	// resolved the rest of the unresolved attributes
	if p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeMaxAttempts] != nil {
		maxAttemptsValue, diags := p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeMaxAttempts].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		if maxAttemptsValue != cty.NilVal {
			maxAttemptsInt, diags := hclhelpers.CtyToInt64(maxAttemptsValue)
			if len(diags) > 0 {
				return nil, diags
			}

			newRetryConfig.MaxAttempts = maxAttemptsInt
		}
	} else {
		newRetryConfig.MaxAttempts = p.RetryConfig.MaxAttempts
	}

	if p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeStrategy] != nil {
		strategyValue, diags := p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeStrategy].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		if strategyValue != cty.NilVal {
			strategyStr, err := hclhelpers.CtyToString(strategyValue)
			if err != nil {
				return nil, hcl.Diagnostics{
					{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse strategy attribute as string",
					},
				}

			}
			newRetryConfig.Strategy = &strategyStr
		}
	} else {
		newRetryConfig.Strategy = p.RetryConfig.Strategy
	}

	if p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeMinInterval] != nil {
		minIntervalValue, diags := p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeMinInterval].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		if minIntervalValue != cty.NilVal {
			minIntervalInt, diags := hclhelpers.CtyToInt64(minIntervalValue)
			if len(diags) > 0 {
				return nil, diags
			}

			newRetryConfig.MinInterval = minIntervalInt
		}
	} else {
		newRetryConfig.MinInterval = p.RetryConfig.MinInterval
	}

	if p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeMaxInterval] != nil {
		maxIntervalValue, diags := p.RetryConfig.UnresolvedAttributes[schema.AttributeTypeMaxInterval].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		if maxIntervalValue != cty.NilVal {
			maxIntervalInt, diags := hclhelpers.CtyToInt64(maxIntervalValue)
			if len(diags) > 0 {
				return nil, diags
			}

			newRetryConfig.MaxInterval = maxIntervalInt
		}
	} else {
		newRetryConfig.MaxInterval = p.RetryConfig.MaxInterval
	}

	diags := newRetryConfig.Validate()
	if len(diags) > 0 {
		return nil, diags
	}

	return newRetryConfig, hcl.Diagnostics{}

}

// For Throw config we want the client to resolve individual element. This to avoid failing on the subsequent throw if an
// earlier throw is executed.
//
// For example: 3 throw configuration. If the first throw condition is met, then there's no reason we should evaluate the subsequent
// throw conditions, let alone failing their evaluation.
func (p *PipelineStepBase) GetThrowConfig() []*ThrowConfig {
	return p.ThrowConfig
}

func (p *PipelineStepBase) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	stepType := p.GetType()

	loopBlocks := blocks.ByType()[schema.BlockTypeLoop]
	if len(loopBlocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one loop block is allowed per step",
			Subject:  &blocks.ByType()[schema.BlockTypeLoop][0].DefRange,
		})
	}

	if len(loopBlocks) == 1 {
		loopBlock := loopBlocks[0]

		loopDefn := GetLoopDefn(stepType, p, &loopBlock.DefRange)
		if loopDefn == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Loop block is not supported for step type %s", stepType),
				Subject:  &loopBlock.DefRange,
			})
		} else {
			attribs, moreDiags := loopBlock.Body.JustAttributes()
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
			} else {
				moreDiags := loopDefn.SetAttributes(attribs, evalContext)
				if len(moreDiags) > 0 {
					diags = append(diags, moreDiags...)
				}
				p.LoopConfig = loopDefn
			}
		}
	}

	errorBlocks := blocks.ByType()[schema.BlockTypeError]
	if len(errorBlocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one error block is allowed per step",
			Subject:  &blocks.ByType()[schema.BlockTypeError][0].DefRange,
		})
	}

	if len(errorBlocks) == 1 {
		errorBlock := errorBlocks[0]
		errorDefn := NewErrorConfig(p)

		var errorBlockAttributes hcl.Attributes
		errorBlockAttributes, diags = errorBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		moreDiags := errorDefn.SetAttributes(errorBlockAttributes, evalContext)
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
		}
		p.ErrorConfig = errorDefn
	}

	retryBlocks := blocks.ByType()[schema.BlockTypeRetry]
	if len(retryBlocks) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Only one retry block is allowed per step",
			Subject:  &blocks.ByType()[schema.BlockTypeRetry][0].DefRange,
		})
	}

	if len(retryBlocks) == 1 {
		retryBlock := retryBlocks[0]
		retryConfig := NewRetryConfig(p)

		var retryBlockAttributes hcl.Attributes
		retryBlockAttributes, diags = retryBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		moreDiags := retryConfig.SetAttributes(retryBlockAttributes, evalContext)

		if len(moreDiags) > 0 {
			return moreDiags
		}

		p.RetryConfig = retryConfig
	}

	throwBlocks := blocks.ByType()[schema.BlockTypeThrow]

	for _, throwBlock := range throwBlocks {
		throwConfig := NewThrowConfig(p)

		attrs, moreDiags := throwBlock.Body.JustAttributes()
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
			continue
		}

		moreDiags = throwConfig.SetAttributes(throwBlock, attrs, evalContext)
		if len(moreDiags) > 0 {
			diags = append(diags, moreDiags...)
			continue
		}

		p.ThrowConfig = append(p.ThrowConfig, throwConfig)

	}

	return diags
}

func (p *PipelineStepBase) AddUnresolvedBody(name string, body hcl.Body) {
	p.UnresolvedBodies[name] = body
}

func (p *PipelineStepBase) Validate() hcl.Diagnostics {
	return hcl.Diagnostics{}
}

func (p *PipelineStepBase) Equals(other *PipelineStepBase) bool {
	if p == nil || other == nil {
		return false
	}

	if p == nil && other != nil || p != nil && other == nil {
		return false
	}

	// Compare ErrorConfig (if not nil)
	if (p.ErrorConfig == nil && other.ErrorConfig != nil) || (p.ErrorConfig != nil && other.ErrorConfig == nil) {
		return false
	}
	if p.ErrorConfig != nil && other.ErrorConfig != nil && !p.ErrorConfig.Equals(other.ErrorConfig) {
		return false
	}

	// Compare retry config
	if (p.RetryConfig == nil && other.RetryConfig != nil) || (p.RetryConfig != nil && other.RetryConfig == nil) {
		return false
	}
	if p.RetryConfig != nil && other.RetryConfig != nil && !p.RetryConfig.Equals(other.RetryConfig) {
		return false
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(p.UnresolvedAttributes) != len(other.UnresolvedAttributes) {
		return false
	}

	for key, expr := range p.UnresolvedAttributes {
		otherExpr, ok := other.UnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	// and reverse
	for key := range other.UnresolvedAttributes {
		if _, ok := p.UnresolvedAttributes[key]; !ok {
			return false
		}
	}

	if len(p.UnresolvedBodies) != len(other.UnresolvedBodies) {
		return false
	}

	for key, body := range p.UnresolvedBodies {
		otherBody, ok := other.UnresolvedBodies[key]
		if !ok || !reflect.DeepEqual(body, otherBody) {
			return false
		}
	}

	// reverse
	for key := range other.UnresolvedBodies {
		if _, ok := p.UnresolvedBodies[key]; !ok {
			return false
		}
	}

	if len(p.ThrowConfig) != len(other.ThrowConfig) {
		return false
	}

	for i, throwConfig := range p.ThrowConfig {
		if !throwConfig.Equals(other.ThrowConfig[i]) {
			return false
		}
	}

	// Compare ForEach (if not nil)
	if (p.ForEach == nil && other.ForEach != nil) || (p.ForEach != nil && other.ForEach == nil) {
		return false
	}
	if p.ForEach != nil && other.ForEach != nil && !hclhelpers.ExpressionsEqual(p.ForEach, other.ForEach) {
		return false
	}

	if (p.LoopConfig == nil && other.LoopConfig != nil) || (p.LoopConfig != nil && other.LoopConfig == nil) {
		return false
	}
	if p.LoopConfig != nil && !p.LoopConfig.Equals(other.LoopConfig) {
		return false
	}

	return p.Name == other.Name &&
		p.Type == other.Type &&
		p.PipelineName == other.PipelineName &&
		utils.PtrEqual(p.MaxConcurrency, other.MaxConcurrency) &&
		reflect.DeepEqual(p.Timeout, other.Timeout) &&
		helpers.StringSliceEqualIgnoreOrder(p.DependsOn, other.DependsOn) &&
		helpers.StringSliceEqualIgnoreOrder(p.CredentialDependsOn, other.CredentialDependsOn) &&
		helpers.StringSliceEqualIgnoreOrder(p.ConnectionDependsOn, other.ConnectionDependsOn) &&
		p.Resolved == other.Resolved
}

func (p *PipelineStepBase) SetPipelineName(pipelineName string) {
	p.PipelineName = pipelineName
}

func (p *PipelineStepBase) GetPipelineName() string {
	return p.PipelineName
}

func (p *PipelineStepBase) GetErrorConfig(evalContext *hcl.EvalContext, ifResolution bool) (*ErrorConfig, hcl.Diagnostics) {

	if p.ErrorConfig == nil {
		return nil, hcl.Diagnostics{}
	}

	if !ifResolution {
		return p.ErrorConfig, hcl.Diagnostics{}
	}

	// do not modify the existing error config, it should always be resolved at runtime
	newErrorConfig := &ErrorConfig{}
	if p.ErrorConfig.UnresolvedAttributes[schema.AttributeTypeIf] != nil {
		ifValue, diags := p.ErrorConfig.UnresolvedAttributes[schema.AttributeTypeIf].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		// If the `if` attribute returns "false" then we return nil for the error config since it doesn't apply
		if !ifValue.True() {
			return nil, hcl.Diagnostics{}
		}

		newErrorConfig.If = utils.ToPointer(ifValue.True())
	}

	if p.ErrorConfig.UnresolvedAttributes[schema.AttributeTypeIgnore] != nil {
		ignoreValue, diags := p.ErrorConfig.UnresolvedAttributes[schema.AttributeTypeIgnore].Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}
		newErrorConfig.Ignore = utils.ToPointer(ignoreValue.True())
	} else if p.ErrorConfig.Ignore != nil {
		if *p.ErrorConfig.Ignore {
			newErrorConfig.Ignore = utils.ToPointer(true)
		}
	}

	return newErrorConfig, hcl.Diagnostics{}
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

func (p *PipelineStepBase) SetPipeline(pipeline *Pipeline) {
	p.Pipeline = pipeline
}

func (p *PipelineStepBase) GetPipeline() *Pipeline {
	return p.Pipeline
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

func (p *PipelineStepBase) GetCredentialDependsOn() []string {
	return p.CredentialDependsOn
}

func (p *PipelineStepBase) GetConnectionDependsOn() []string {
	return p.ConnectionDependsOn
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

func (p *PipelineStepBase) AppendCredentialDependsOn(credentialDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]bool)
	for _, dep := range p.CredentialDependsOn {
		existingDeps[dep] = true
	}

	for _, dep := range credentialDependsOn {
		if !existingDeps[dep] {
			p.CredentialDependsOn = append(p.CredentialDependsOn, dep)
			existingDeps[dep] = true
		}
	}
}

func (p *PipelineStepBase) AppendConnectionDependsOn(connectionDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]struct{}, len(p.ConnectionDependsOn))
	for _, dep := range p.ConnectionDependsOn {
		existingDeps[dep] = struct{}{}
	}

	for _, dep := range connectionDependsOn {
		if _, exists := existingDeps[dep]; !exists {
			p.ConnectionDependsOn = append(p.ConnectionDependsOn, dep)
			existingDeps[dep] = struct{}{}
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

func (p *PipelineStepBase) GetMaxConcurrency(evalContext *hcl.EvalContext) *int {
	if p.MaxConcurrency != nil {
		return p.MaxConcurrency
	} else if p.UnresolvedAttributes[schema.AttributeTypeMaxConcurrency] != nil {
		val, diags := p.UnresolvedAttributes[schema.AttributeTypeMaxConcurrency].Value(evalContext)
		if len(diags) > 0 {
			return nil
		}

		if val == cty.NilVal {
			return nil
		}

		maxConcurrency, err := hclhelpers.CtyToGo(val)
		if err != nil {
			return nil
		}

		maxConcurrencyInt, ok := maxConcurrency.(int)
		if !ok {
			return nil
		}

		return &maxConcurrencyInt
	}
	return p.MaxConcurrency
}

func (p *PipelineStepBase) SetBaseAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
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

	if attr, exists := hclAttributes[schema.AttributeTypeMaxConcurrency]; exists {
		val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
		if stepDiags.HasErrors() {
			diags = append(diags, stepDiags...)
		} else if val != cty.NilVal {
			maxConcurrency, err := hclhelpers.CtyToGo(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse '" + schema.AttributeTypeMaxConcurrency + "' attribute to interface",
					Subject:  &attr.Range,
				})
			} else {
				maxConcurrencyInt, ok := maxConcurrency.(int)
				if !ok {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Value of the attribute '" + schema.AttributeTypeMaxConcurrency + "' must be a whole number: " + p.GetFullyQualifiedName(),
						Subject:  &attr.Range,
					})
				} else {

					p.MaxConcurrency = &maxConcurrencyInt
				}
			}
		}

	}

	if attr, exists := hclAttributes[schema.AttributeTypeTimeout]; exists {
		val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
		if stepDiags.HasErrors() {
			diags = append(diags, stepDiags...)
		} else if val != cty.NilVal {
			duration, err := hclhelpers.CtyToGo(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse '" + schema.AttributeTypeTimeout + "' attribute to interface",
					Subject:  &attr.Range,
				})
			}
			p.Timeout = duration
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

	p.AppendDependsOn(dependsOn...)

	return diags
}

func (p *PipelineStepBase) GetBaseInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	inputs := map[string]interface{}{}

	if p.UnresolvedAttributes[schema.AttributeTypeTimeout] == nil && p.Timeout != nil {
		inputs[schema.AttributeTypeTimeout] = p.Timeout
	} else if p.UnresolvedAttributes[schema.AttributeTypeTimeout] != nil {

		var timeoutDurationCtyValue cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeTimeout], evalContext, &timeoutDurationCtyValue)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}

		goVal, err := hclhelpers.CtyToGo(timeoutDurationCtyValue)
		if err != nil {
			return nil, err
		}
		inputs[schema.AttributeTypeTimeout] = goVal
	}

	return inputs, nil
}

func (p *PipelineStepBase) ValidateBaseAttributes() hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	if p.Timeout != nil {
		switch p.Timeout.(type) {
		case string, int:
			// valid duration
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Value of the attribute '" + schema.AttributeTypeTimeout + "' must be a string or a whole number: " + p.GetFullyQualifiedName(),
				Subject:  p.Range,
			})
		}
	}

	return diags
}

func (p *PipelineStepBase) HandleDecodeBodyDiags(diags hcl.Diagnostics, attributeName string, body hcl.Body) hcl.Diagnostics {
	resolvedDiags := 0

	unresolvedDiags := hcl.Diagnostics{}

	for _, e := range diags {
		if e.Severity == hcl.DiagError {
			if e.Detail == `There is no variable named "step".` || e.Detail == `There is no variable named "credential".` || e.Detail == `There is no variable named "connection".` {
				traversals := e.Expression.Variables()
				dependsOnAdded := false
				for _, traversal := range traversals {
					parts := hclhelpers.TraversalAsStringSlice(traversal)
					if len(parts) > 0 {
						// When the expression/traversal is referencing an index, the index is also included in the parts
						// for example: []string len: 5, cap: 5, ["step","sleep","sleep_1","0","duration"]
						if parts[0] == schema.BlockTypePipelineStep {
							if len(parts) < 3 {
								return diags
							}
							dependsOn := parts[1] + "." + parts[2]
							p.AppendDependsOn(dependsOn)
							dependsOnAdded = true
						} else if parts[0] == schema.BlockTypeCredential {
							if len(parts) < 2 {
								return diags
							}

							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = credential.aws[param.cred].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								p.AppendCredentialDependsOn(dependsOn)
								dependsOnAdded = true
							} else {
								dependsOn := parts[1] + "." + parts[2]
								p.AppendCredentialDependsOn(dependsOn)
								dependsOnAdded = true
							}
						} else if parts[0] == schema.BlockTypeConnection {
							if len(parts) < 2 {
								return diags
							}

							if len(parts) == 2 {
								// dynamic references:
								// step "transform" "aws" {
								// 	value   = connection.aws[param.conn].env
								// }
								dependsOn := parts[1] + ".<dynamic>"
								p.AppendConnectionDependsOn(dependsOn)
								dependsOnAdded = true
							} else {
								dependsOn := parts[1] + "." + parts[2]
								p.AppendConnectionDependsOn(dependsOn)
								dependsOnAdded = true
							}
						}
					}
				}
				if dependsOnAdded {
					resolvedDiags++
				}
			} else if e.Detail == `There is no variable named "result".` && (attributeName == schema.BlockTypeLoop || attributeName == schema.BlockTypeRetry || attributeName == schema.BlockTypeThrow) {
				// result is a reference to the output of the step after it was run, however it should only apply to the loop type block or retry type block
				resolvedDiags++
			} else if e.Detail == `There is no variable named "each".` || e.Detail == `There is no variable named "param".` || e.Detail == "Unsuitable value: value must be known" || e.Detail == `There is no variable named "loop".` || e.Detail == `There is no variable named "retry".` {

				// hcl.decodeBody returns 2 error messages:
				// 1. There's no variable named "param", AND
				// 2. Unsuitable value: value must be known
				resolvedDiags++

			} else {
				unresolvedDiags = append(unresolvedDiags, e)
			}
		}
	}

	// check if all diags have been resolved
	if resolvedDiags == len(diags) {
		if attributeName == schema.BlockTypeThrow {
			return hcl.Diagnostics{}
		} else {
			// * Don't forget to add this, if you change the logic ensure that the code flow still
			// * calls AddUnresolvedBody
			p.AddUnresolvedBody(attributeName, body)
			return hcl.Diagnostics{}
		}
	}

	// There's an error here
	return unresolvedDiags

}

var ValidBaseStepAttributes = []string{
	schema.AttributeTypeTitle,
	schema.AttributeTypeDescription,
	schema.AttributeTypeDependsOn,
	schema.AttributeTypeForEach,
	schema.AttributeTypeIf,
	schema.AttributeTypeTimeout,
	schema.AttributeTypeMaxConcurrency,
}

var ValidDependsOnTypes = []string{
	schema.BlockTypePipelineStep,
}

func (p *PipelineStepBase) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseStepAttributes, name)
}

func interfaceSliceInputFromAttribute(unresolvedAttributes map[string]hcl.Expression, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName string, fieldValue *[]interface{}) (map[string]interface{}, hcl.Diagnostics) {
	var tempValue *[]interface{}

	unresolvedAttrib := unresolvedAttributes[attributeName]

	if unresolvedAttrib == nil {
		tempValue = fieldValue
	} else {
		var args cty.Value

		diags := gohcl.DecodeExpression(unresolvedAttrib, evalContext, &args)
		if diags.HasErrors() {
			return nil, diags
		}

		var err error
		interfaceSlice, err := hclhelpers.CtyToGoInterfaceSlice(args)
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + attributeName + " attribute to interface slice",
					Subject:  unresolvedAttrib.Range().Ptr(),
				},
			}
		}
		if interfaceSlice != nil {
			tempValue = &interfaceSlice
		}
	}

	if tempValue != nil {
		results[attributeName] = *tempValue
	}

	return results, hcl.Diagnostics{}
}

func stringSliceInputFromAttribute(unresolvedAttributes map[string]hcl.Expression, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName string, fieldValue *[]string) (map[string]interface{}, hcl.Diagnostics) {
	var tempValue *[]string

	unresolvedAttrib := unresolvedAttributes[attributeName]

	if unresolvedAttrib == nil {
		tempValue = fieldValue
	} else {
		var args cty.Value

		diags := gohcl.DecodeExpression(unresolvedAttrib, evalContext, &args)
		if diags.HasErrors() {
			return nil, diags
		}

		var err error
		stringSlice, err := hclhelpers.CtyToGoStringSlice(args, args.Type())
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + attributeName + " attribute to string",
					Subject:  unresolvedAttrib.Range().Ptr(),
				},
			}
		}
		if stringSlice != nil {
			tempValue = &stringSlice
		}
	}

	if tempValue != nil {
		results[attributeName] = *tempValue
	}

	return results, hcl.Diagnostics{}
}

func simpleOutputFromAttribute[T any](unresolvedAttributes map[string]hcl.Expression, evalContext *hcl.EvalContext, attributeName string, fieldValue T) (T, hcl.Diagnostics) {
	var tempValue T

	if !helpers.IsNil(fieldValue) {
		if utils.IsPointer(fieldValue) {
			return tempValue, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Value of the attribute '" + attributeName + "' must not be a pointer for simpleOutputFromAttribute to work",
				},
			}
		}
	}

	if unresolvedAttributes[attributeName] == nil {
		if !helpers.IsNil(fieldValue) {
			tempValue = fieldValue
		}
	} else {
		diags := gohcl.DecodeExpression(unresolvedAttributes[attributeName], evalContext, &tempValue)
		if diags.HasErrors() {
			return tempValue, diags
		}
	}

	return tempValue, hcl.Diagnostics{}
}

func simpleTypeInputFromAttribute[T any](unresolvedAttributes map[string]hcl.Expression, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName string, fieldValue T) (map[string]interface{}, hcl.Diagnostics) {
	var tempValue T

	if unresolvedAttributes[attributeName] == nil {
		if !helpers.IsNil(fieldValue) {
			tempValue = fieldValue
		}
	} else {
		diags := gohcl.DecodeExpression(unresolvedAttributes[attributeName], evalContext, &tempValue)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	if !helpers.IsNil(tempValue) {
		if utils.IsPointer(tempValue) {
			// Reflect on tempValue to get its underlying value if it's a pointer
			valueOfTempValue := reflect.ValueOf(tempValue)
			if valueOfTempValue.Kind() == reflect.Ptr && !valueOfTempValue.IsNil() {
				// Dereference the pointer and set the result in the map
				results[attributeName] = valueOfTempValue.Elem().Interface()
			} else {
				results[attributeName] = tempValue
			}
		} else {
			results[attributeName] = tempValue
		}
	}

	return results, hcl.Diagnostics{}
}

func decodeStepAttribute[T any](unresolvedAttributes map[string]hcl.Expression, evalContext *hcl.EvalContext, stepName, attributeName string, fieldValue T) (any, []ConnectionDependency, hcl.Diagnostics) {
	var res any
	var connectionDependencies []ConnectionDependency

	var decodedCtyValue cty.Value

	if unresolvedAttributes[attributeName] == nil {
		if !helpers.IsNil(fieldValue) {
			if utils.IsPointer(fieldValue) {
				valueOfTempValue := reflect.ValueOf(fieldValue)
				if valueOfTempValue.Kind() == reflect.Ptr && !valueOfTempValue.IsNil() {
					// Dereference the pointer and set the result in the map
					res = valueOfTempValue.Elem().Interface()
				} else {
					res = fieldValue
				}
			} else {
				res = fieldValue
			}
		}

	} else {
		diags := gohcl.DecodeExpression(unresolvedAttributes[attributeName], evalContext, &decodedCtyValue)
		if diags.HasErrors() {
			// Handle connection errors
			if IsConnectionError(diags) {
				conns := FindConnectionFromDiags(diags)
				slog.Debug(fmt.Sprintf("Missing connections for step.%s.%s", attributeName, stepName), "connections", conns)
				connectionDependencies = append(connectionDependencies, conns...)
				return res, connectionDependencies, nil
			}
			// Return any other errors
			return res, nil, diags
		}
	}

	if !decodedCtyValue.IsNull() {

		fieldType := reflect.TypeOf(fieldValue)
		if fieldType != nil && fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.String {
			var err error
			res, err = hclhelpers.CtyToGoStringSlice(decodedCtyValue, decodedCtyValue.Type())
			if err != nil {
				return res, nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + attributeName + " attribute to string slice",
						Subject:  unresolvedAttributes[attributeName].Range().Ptr(),
					},
				}
			}
		} else if fieldType != nil && fieldType.Kind() == reflect.Map && fieldType.Key().Kind() == reflect.String && fieldType.Elem().Kind() == reflect.String {
			var err error
			res, err = hclhelpers.CtyToGoMapString(decodedCtyValue)
			if err != nil {
				return res, nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + attributeName + " attribute to map[string]string",
						Subject:  unresolvedAttributes[attributeName].Range().Ptr(),
					},
				}
			}
		} else {
			goVal, err := hclhelpers.CtyToGo(decodedCtyValue)
			if err != nil {
				return res, nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + attributeName + " attribute to interface",
						Subject:  unresolvedAttributes[attributeName].Range().Ptr(),
					},
				}
			}

			res = goVal
		}
	}

	return res, connectionDependencies, hcl.Diagnostics{}
}

func stringMapInputFromAttribute(unresolvedAttributes map[string]hcl.Expression, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName string, fieldValue *map[string]string) (map[string]interface{}, hcl.Diagnostics) {
	if fieldValue != nil {
		results[attributeName] = *fieldValue
	} else if unresolvedAttributes[attributeName] != nil {
		attr := unresolvedAttributes[attributeName]
		val, diags := attr.Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		if val != cty.NilVal {
			mapValues, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				return nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + attributeName + " attribute to map",
						Subject:  attr.Range().Ptr(),
					},
				}
			}

			results[attributeName] = mapValues
		}
	}

	return results, hcl.Diagnostics{}
}

func mapInterfaceInputFromAttribute(unresolvedAttributes map[string]hcl.Expression, results map[string]interface{}, evalContext *hcl.EvalContext, attributeName string, fieldValue *map[string]interface{}) (map[string]interface{}, hcl.Diagnostics) {
	if fieldValue != nil {
		results[attributeName] = *fieldValue
	} else if unresolvedAttributes[attributeName] != nil {
		attr := unresolvedAttributes[attributeName]
		val, diags := attr.Value(evalContext)
		if len(diags) > 0 {
			return nil, diags
		}

		if val != cty.NilVal {
			mapValues, err := hclhelpers.CtyToGoMapInterface(val)
			if err != nil {
				return nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + attributeName + " attribute to map",
						Subject:  attr.Range().Ptr(),
					},
				}
			}

			results[attributeName] = mapValues
		}
	}

	return results, hcl.Diagnostics{}
}

// setField sets the field of a struct pointed to by v to the given value.
// v must be a pointer to a struct, fieldName must be the name of a field in the struct,
// and value must be assignable to the field.
func setField(v interface{}, fieldName string, value interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return perr.BadRequestWithMessage("v must be a pointer to a struct")
	}

	rv = rv.Elem() // Dereference the pointer to get the struct

	field := rv.FieldByName(fieldName)
	if !field.IsValid() {
		return perr.BadRequestWithMessage(fmt.Sprintf("no such field: %s in obj", fieldName))
	}

	if !field.CanSet() {
		return perr.BadRequestWithMessage(fmt.Sprintf("cannot set field %s", fieldName))
	}

	fieldValue := reflect.ValueOf(value)
	if field.Type() != fieldValue.Type() {
		return perr.BadRequestWithMessage("provided value type does not match field type")
	}

	field.Set(fieldValue)
	return nil
}

func setInterfaceSliceAttributeWithResultReference(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool, resultsReference bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, p, resultsReference)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	t, err := hclhelpers.CtyToGoInterfaceSlice(val)
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to interface slice",
				Subject:  &attr.Range,
			},
		}
	}

	if isPtr {
		err = setField(p, fieldName, &t)
	} else {
		err = setField(p, fieldName, t)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
}

func setStringSliceAttributeWithResultReference(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool, resultsReference bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, p, resultsReference)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	t, err := hclhelpers.CtyToGoStringSlice(val, val.Type())
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to string",
				Subject:  &attr.Range,
			},
		}
	}

	if isPtr {
		err = setField(p, fieldName, &t)
	} else {
		err = setField(p, fieldName, t)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
}

func setStringSliceAttribute(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool) hcl.Diagnostics {
	return setStringSliceAttributeWithResultReference(attr, evalContext, p, fieldName, isPtr, false)
}

func setStringAttributeWithResultReference(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool, resultsReference bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, p, resultsReference)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	t, err := hclhelpers.CtyToString(val)
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to string",
				Subject:  &attr.Range,
			},
		}
	}

	if isPtr {
		err = setField(p, fieldName, &t)
	} else {
		err = setField(p, fieldName, t)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
}

func setStringAttribute(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool) hcl.Diagnostics {
	return setStringAttributeWithResultReference(attr, evalContext, p, fieldName, isPtr, false)
}

func setBoolAttributeWithResultReference(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool, resultReference bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, p, resultReference)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	if val.Type() != cty.Bool {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to bool",
				Subject:  &attr.Range,
			},
		}
	}

	t := val.True()

	var err error

	if isPtr {
		err = setField(p, fieldName, &t)
	} else {
		err = setField(p, fieldName, t)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
}
func setBoolAttribute(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool) hcl.Diagnostics {
	return setBoolAttributeWithResultReference(attr, evalContext, p, fieldName, isPtr, false)
}

func setInt64AttributeWithResultReference(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool, resultReference bool) hcl.Diagnostics {
	val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, p, resultReference)
	if stepDiags.HasErrors() {
		return stepDiags
	}

	if val == cty.NilVal {
		return hcl.Diagnostics{}
	}

	if val.Type() != cty.Number {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + " attribute to number",
				Subject:  &attr.Range,
			},
		}
	}

	tPtr, moreDiags := hclhelpers.CtyToInt64(val)
	if moreDiags != nil && moreDiags.HasErrors() {
		return moreDiags
	}

	var err error

	if isPtr {
		err = setField(p, fieldName, tPtr)
	} else {
		err = setField(p, fieldName, *tPtr)
	}

	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to set " + attr.Name + " attribute to struct",
				Subject:  &attr.Range,
			},
		}
	}

	return hcl.Diagnostics{}
}

// func setInt64Attribute(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, fieldName string, isPtr bool) hcl.Diagnostics {
// 	return setInt64AttributeWithResultReference(attr, evalContext, p, fieldName, isPtr, false)
// }

func dependsOnFromExpressions(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface) (cty.Value, hcl.Diagnostics) {
	return dependsOnFromExpressionsWithResultControl(attr, evalContext, p, false)
}

func findTryFunction(expr hcl.Expression) bool {

	binaryOpExpr, ok := expr.(*hclsyntax.BinaryOpExpr)
	if ok {
		return findTryFunction(binaryOpExpr.LHS) || findTryFunction(binaryOpExpr.RHS)
	}

	funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok {
		return false
	}

	if funcExpr.Name == "try" {
		return true
	}

	for _, a1 := range funcExpr.Args {
		if findTryFunction(a1) {
			return true
		}
	}
	return false
}

func allDependsOnFromVariables(traversals []hcl.Traversal) ([]string, []string, []string) {
	var allDependsOn []string
	var allCredentialDependsOn []string
	var allConnectionDependsOn []string

	for _, traversal := range traversals {
		parts := hclhelpers.TraversalAsStringSlice(traversal)
		if len(parts) > 0 {
			// When the expression/traversal is referencing an index, the index is also included in the parts
			// for example: []string len: 5, cap: 5, ["step","sleep","sleep_1","0","duration"]
			if parts[0] == schema.BlockTypePipelineStep {
				if len(parts) < 3 {
					continue
				}
				dependsOn := parts[1] + "." + parts[2]
				allDependsOn = append(allDependsOn, dependsOn)

			} else if parts[0] == schema.BlockTypeCredential {
				if len(parts) < 2 {
					continue
				}

				if len(parts) == 2 {
					// dynamic references:
					// step "transform" "aws" {
					// 	value   = credential.aws[param.cred].env
					// }
					dependsOn := parts[1] + ".<dynamic>"
					allCredentialDependsOn = append(allCredentialDependsOn, dependsOn)

				} else {
					dependsOn := parts[1] + "." + parts[2]
					allCredentialDependsOn = append(allCredentialDependsOn, dependsOn)
				}
			} else if parts[0] == schema.BlockTypeConnection {
				if len(parts) < 2 {
					continue
				}

				if len(parts) == 2 {
					// dynamic references:
					// step "transform" "aws" {
					// 	value   = connection.aws[param.conn].env
					// }
					dependsOn := parts[1] + ".<dynamic>"
					allConnectionDependsOn = append(allConnectionDependsOn, dependsOn)

				} else {
					dependsOn := parts[1] + "." + parts[2]
					allConnectionDependsOn = append(allConnectionDependsOn, dependsOn)
				}
			}
		}
	}

	return allDependsOn, allCredentialDependsOn, allConnectionDependsOn
}

func dependsOnFromExpressionsWithResultControl(attr *hcl.Attribute, evalContext *hcl.EvalContext, p PipelineStepBaseInterface, resultsReference bool) (cty.Value, hcl.Diagnostics) {
	expr := attr.Expr

	// If there is a param in the expression, then we must assume that we can't resolve it at this stage.
	// If the param has a default, it will be fully resolved and when we change the param, Flowpipe doesn't know that the
	// attribute needs to be recalculated
	for _, traversal := range expr.Variables() {
		if traversal.RootName() == "param" {
			p.AddUnresolvedAttribute(attr.Name, expr)
			// Don't return here because there may be other dependencies to be created below

			// special handling if the attribute name is "pipeline"
			//
			// this is to handle the pipeline step:
			/**

			step "pipeline" "run_pipeline {
				pipeline = pipeline[param.name]
			}

			we short circuit it straight away and return. It will be resolved at runtime. We can't do that for other attributes because we
			may do something like:

			value = "${param.foo} and ${step.transform.name.value}"

			so the above has dependency on param.foo AND the step.transform.name. We *need* to add `step.transform.name` to the depends_on list
			so it can't return here

			pipeline attribute is special that it can only reference another pipeline
			*/

			if attr.Name == "pipeline" {
				return cty.NilVal, hcl.Diagnostics{}
			}
		}
		if traversal.RootName() == "var" {
			// if the variable is a late binding variable, then we need to add the connection names
			// to the connection depends on
			connectionNames := modconfig.ResourceNamesFromLateBindingVarTraversal(traversal, evalContext)
			if len(connectionNames) > 0 {
				p.AppendConnectionDependsOn(connectionNames...)
				p.AddUnresolvedAttribute(attr.Name, expr)
			}
		}
	}

	isTryFunction := findTryFunction(expr)
	if isTryFunction {
		p.AddUnresolvedAttribute(attr.Name, expr)

		dependsOn, credsDependsOn, connDependsOn := allDependsOnFromVariables(expr.Variables())
		p.AppendDependsOn(dependsOn...)
		p.AppendCredentialDependsOn(credsDependsOn...)
		p.AppendConnectionDependsOn(connDependsOn...)

		// don't bother continuing, we should not resolve it if there's an "if" function
		return cty.NilVal, hcl.Diagnostics{}
	}

	leftOverDiags := hcl.Diagnostics{}
	// resolve it first if we can
	val, stepDiags := expr.Value(evalContext)
	if stepDiags != nil && stepDiags.HasErrors() {
		resolvedDiags := 0
		for _, e := range stepDiags {
			if e.Severity == hcl.DiagError {
				// is the error caused by referencing a resource whose value will be resolved at runtime
				// (i.e. connection, credential, step output)?,
				if lateBindingValueError(e) {
					// identify connections, credentials and other resources which the step depends on
					dependsOnAdded := handleMissingDependencyError(expr, p)
					if dependsOnAdded {
						resolvedDiags++
					}
					// is the error an expected dependency error (in which case we ignore)
					// (i.e.  try function, each, param, loop, retry)
				} else if expectedMissingDependencyError(e) ||
					tryFunctionError(e) ||
					stepResultError(e, resultsReference) {
					resolvedDiags++
					// is the error caused by referencing a variable whose value will be resolved at runtime
				} else if resourceNames := modconfig.ResourceNamesFromLateBindingVarValueError(e, evalContext); len(resourceNames) > 0 {
					p.AppendConnectionDependsOn(resourceNames...)
					resolvedDiags++
				} else {
					leftOverDiags = append(leftOverDiags, e)
				}
			}
		}

		if len(leftOverDiags) > 0 {
			return cty.NilVal, leftOverDiags
		}

		// check if all diags have been resolved
		if resolvedDiags == len(stepDiags) {
			// mop up any other depends on that may be missed with the above logic
			handleMissingDependencyError(expr, p)

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

func handleMissingDependencyError(expr hcl.Expression, p PipelineStepBaseInterface) bool {
	dependsOn, credsDependsOn, connDependsOn := allDependsOnFromVariables(expr.Variables())
	p.AppendDependsOn(dependsOn...)
	p.AppendCredentialDependsOn(credsDependsOn...)
	p.AppendConnectionDependsOn(connDependsOn...)

	dependsOnAdded := len(dependsOn) > 0 || len(credsDependsOn) > 0 || len(connDependsOn) > 0
	return dependsOnAdded
}

func lateBindingValueError(e *hcl.Diagnostic) bool {
	return e.Detail == `There is no variable named "step".` || e.Detail == `There is no variable named "credential".` || e.Detail == `There is no variable named "connection".`
}

func stepResultError(e *hcl.Diagnostic, resultsReference bool) bool {
	// result is a reference to the output of the step after it was run, however it should only apply to the loop type block or retry type block
	return e.Detail == `There is no variable named "result".` && resultsReference
}

func tryFunctionError(e *hcl.Diagnostic) bool {
	// try function generate this different error message:
	// "Call to function \"try\" failed: no expression succeeded:\n- Unknown variable (at /Users/victorhadianto/z-development/turbot/pipe-fittings/tests/flowpipe_mod_tests/mod_try_function/mod.fp:25,21-25)\n  There is no variable named \"each\".\n- Unknown variable (at /Users/victorhadianto/z-development/turbot/pipe-fittings/tests/flowpipe_mod_tests/mod_try_function/mod.fp:25,85-89)\n  There is no variable named \"each\".\n\nAt least one expression must produce a successful result."
	return strings.Contains(e.Detail, `Call to function "try" failed`) && strings.Contains(e.Detail, `There is no variable named "each".`)
}

func expectedMissingDependencyError(e *hcl.Diagnostic) bool {
	return e.Detail == `There is no variable named "each".` || e.Detail == `There is no variable named "param".` || e.Detail == `There is no variable named "loop".`
}

func IsConnectionScopeTraversal(scopeTraversals *hclsyntax.ScopeTraversalExpr) bool {
	if len(scopeTraversals.Traversal) > 0 {
		if traverserRoot, ok := scopeTraversals.Traversal[0].(hcl.TraverseRoot); ok {
			if traverserRoot.Name == "connection" || traverserRoot.Name == "credential" {
				return true
			}
		}
	}
	return false
}

func IsConnectionIndexTraversal(indexExpr *hclsyntax.IndexExpr) bool {
	if key, ok := indexExpr.Key.(*hclsyntax.ScopeTraversalExpr); ok {
		if IsConnectionScopeTraversal(key) {
			return true
		}
	}

	if collection, ok := indexExpr.Collection.(*hclsyntax.ScopeTraversalExpr); ok {
		if IsConnectionScopeTraversal(collection) {
			return true
		}
	}

	return false
}

func IsConnectionExpr(expr hcl.Expression) bool {
	if scopeTraversals, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		return IsConnectionScopeTraversal(scopeTraversals)
	} else if indexExpr, ok := expr.(*hclsyntax.IndexExpr); ok {
		return IsConnectionIndexTraversal(indexExpr)
	} else if objType, ok := expr.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range objType.Items {
			if IsConnectionExpr(item.ValueExpr) {
				return true
			}
		}
	} else if templateExpr, ok := expr.(*hclsyntax.TemplateExpr); ok {
		for _, part := range templateExpr.Parts {
			if IsConnectionExpr(part) {
				return true
			}
		}
	}

	return false
}

func IsConnectionError(diags hcl.Diagnostics) bool {
	if !diags.HasErrors() {
		return false
	}

	for _, diag := range diags {
		if diag.Expression != nil {
			if IsConnectionExpr(diag.Expression) {
				return true
			}
		}
	}

	return false
}

func GuessRequiredConnectionScopeTraversal(scopeTraversals *hclsyntax.ScopeTraversalExpr) []ConnectionDependency {
	var res []ConnectionDependency

	connectionDependency := IsConnectionScopeTraversal(scopeTraversals)
	if !connectionDependency {
		return res
	}

	dottedString := hclhelpers.TraversalAsString(scopeTraversals.Traversal)
	parts := strings.Split(dottedString, ".")
	if len(parts) < 2 {
		return res
	}

	connectionType := parts[1]
	connectionSource := ""
	if len(parts) > 2 {
		connectionSource = parts[2]
	}

	return []ConnectionDependency{
		{
			Source: connectionSource,
			Type:   connectionType,
		},
	}
}

func GuessRequiredConnectionIndex(indexExpr *hclsyntax.IndexExpr) []ConnectionDependency {
	// value = connection.aws[step.transform.source.value]

	var res []ConnectionDependency
	// check the key if it's a connection dependency
	connectionDependency := false
	if scopeTraverals, ok := indexExpr.Collection.(*hclsyntax.ScopeTraversalExpr); ok {
		connectionDependency = IsConnectionScopeTraversal(scopeTraverals)
	}

	if !connectionDependency {
		return res
	}

	connType := ""
	connSource := ""

	if key, ok := indexExpr.Collection.(*hclsyntax.ScopeTraversalExpr); ok {
		dottedString := hclhelpers.TraversalAsString(key.Traversal)
		parts := strings.Split(dottedString, ".")
		if len(parts) < 2 {
			return res
		}

		connType = parts[1]
	}

	// flatten the keys
	if key, ok := indexExpr.Key.(*hclsyntax.ScopeTraversalExpr); ok {
		connSource = hclhelpers.TraversalAsString(key.Traversal)

	}

	return []ConnectionDependency{
		{
			Source: connSource,
			Type:   connType,
		},
	}
}

func GuessRequiredConnectionRelativeTraversal(relativeTraversal *hclsyntax.RelativeTraversalExpr) []ConnectionDependency {
	var res []ConnectionDependency

	if scopeTraversals, ok := relativeTraversal.Source.(*hclsyntax.ScopeTraversalExpr); ok {
		newRes := GuessRequiredConnectionScopeTraversal(scopeTraversals)
		res = append(res, newRes...)
	} else if indexExpr, ok := relativeTraversal.Source.(*hclsyntax.IndexExpr); ok {
		newRes := GuessRequiredConnectionIndex(indexExpr)
		res = append(res, newRes...)
	}

	return res
}

func GuessRequiredConnectionTemplate(templateExpr *hclsyntax.TemplateExpr) []ConnectionDependency {
	var res []ConnectionDependency

	for _, part := range templateExpr.Parts {
		if scopeTraverals, ok := part.(*hclsyntax.ScopeTraversalExpr); ok {
			newRes := GuessRequiredConnectionScopeTraversal(scopeTraverals)
			res = append(res, newRes...)
		} else if relativeTraversal, ok := part.(*hclsyntax.RelativeTraversalExpr); ok {
			newRes := GuessRequiredConnectionRelativeTraversal(relativeTraversal)
			res = append(res, newRes...)
		} else if indexExpr, ok := part.(*hclsyntax.IndexExpr); ok {
			newRes := GuessRequiredConnectionIndex(indexExpr)
			res = append(res, newRes...)
		}
	}

	return res
}

func FindRequiredConnection(expr hcl.Expression) string {
	connectionType := ""
	connectionSource := ""
	if indexExpr, ok := expr.(*hclsyntax.IndexExpr); ok {
		if collectionExpr, ok := indexExpr.Collection.(*hclsyntax.ScopeTraversalExpr); ok {
			collectionString := hclhelpers.TraversalAsString(collectionExpr.Traversal)
			collectionStringParts := strings.Split(collectionString, ".")
			if len(collectionStringParts) == 2 {
				connectionType = collectionStringParts[1]

			}
		}

		if keyExpr, ok := indexExpr.Key.(*hclsyntax.ScopeTraversalExpr); ok {
			connectionSource = hclhelpers.TraversalAsString(keyExpr.Traversal)
		}
	}

	if connectionType != "" && connectionSource != "" {
		return connectionType + "." + connectionSource
	}

	return ""
}

func GuessRequiredConnection(expr hcl.Expression) []ConnectionDependency {
	var res []ConnectionDependency

	if scopeTraverals, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		newRes := GuessRequiredConnectionScopeTraversal(scopeTraverals)
		res = append(res, newRes...)
	} else if objType, ok := expr.(*hclsyntax.ObjectConsExpr); ok {
		/**
				    step "transform" "repeat" {
		        for_each = step.transform.source.value
		        value = {
		            "value" = "bar"
		            "param_foo" = param.foo
		            "akey" = connection.aws[each.value]
		        }
		    }

			**/
		for _, item := range objType.Items {
			if scopeTraverals, ok := item.ValueExpr.(*hclsyntax.ScopeTraversalExpr); ok {
				newRes := GuessRequiredConnectionScopeTraversal(scopeTraverals)
				res = append(res, newRes...)
			} else if indexExpr, ok := item.ValueExpr.(*hclsyntax.IndexExpr); ok {
				newRes := GuessRequiredConnectionIndex(indexExpr)
				res = append(res, newRes...)
			} else if templateExpr, ok := item.ValueExpr.(*hclsyntax.TemplateExpr); ok {
				newRes := GuessRequiredConnectionTemplate(templateExpr)
				res = append(res, newRes...)
			}

		}
	} else if indexExpr, ok := expr.(*hclsyntax.IndexExpr); ok {
		newRes := GuessRequiredConnectionIndex(indexExpr)
		res = append(res, newRes...)
	}

	return res
}

func FindConnectionFromDiags(diags hcl.Diagnostics) []ConnectionDependency {
	var res []ConnectionDependency

	for _, diag := range diags {
		if diag.Expression != nil {
			newRes := GuessRequiredConnection(diag.Expression)
			res = append(res, newRes...)
		}
	}

	return res
}
