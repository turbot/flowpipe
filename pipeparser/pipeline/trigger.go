package pipeline

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"

	"github.com/robfig/cron/v3"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	ctx         context.Context
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Args        Input   `json:"args"`

	Pipeline cty.Value `json:"-"`
	RawBody  hcl.Body  `json:"-" hcl:",remain"`
}

func (t *Trigger) GetName() string {
	return t.Name
}

func (t *Trigger) GetDescription() *string {
	return t.Description
}

func (t *Trigger) GetPipeline() cty.Value {
	return t.Pipeline
}

func (t *Trigger) SetName(name string) {
	t.Name = name
}

func (t *Trigger) SetContext(ctx context.Context) {
	t.ctx = ctx
}

func (t *Trigger) GetArgs() Input {
	return t.Args
}

var ValidBaseTriggerAttributes = []string{
	schema.AttributeTypeDescription,
	schema.AttributeTypePipeline,
	schema.AttributeTypeArgs,
}

func (t *Trigger) IsBaseAttribute(name string) bool {
	return helpers.StringSliceContains(ValidBaseTriggerAttributes, name)
}

func (t *Trigger) SetBaseAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, diag := hclhelpers.AttributeToString(attr, nil, false)
		if diag != nil && diag.Severity == hcl.DiagError {
			diags = append(diags, diag)
		} else {
			t.Description = desc
		}
	}

	// Pipeline is a required attribute, we don't need to validate it here because
	// it should be defined in the Trigger Schema
	attr := hclAttributes[schema.AttributeTypePipeline]

	expr := attr.Expr
	// Try to validate the pipeline reference. It's OK to do this here because by the time
	// we parse the triggers, we should have loaded all the pipelines in the Parser Context.
	//
	// Can't do it for the step references because the pipeline that a step refer to may not be parsed
	// yet.
	val, err := expr.Value(parseContext.EvalCtx)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + schema.AttributeTypePipeline + " attribute: " + err.Error(),
			Subject:  &attr.Range,
		})
	} else {
		t.Pipeline = val
	}

	if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
		if attr.Expr != nil {
			expr := attr.Expr
			vals, err := expr.Value(nil)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute",
					Subject:  &attr.Range,
				})

			} else {
				goVals, err := hclhelpers.CtyToGoMapInterface(vals)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute to Go values",
						Subject:  &attr.Range,
					})
				}
				t.Args = goVals
			}
		}
	}

	return diags
}

type ITrigger interface {
	SetContext(context.Context)
	SetName(string)
	GetName() string
	GetDescription() *string
	GetPipeline() cty.Value
	GetArgs() Input
	SetAttributes(hcl.Attributes, *pipeparser.ParseContext) hcl.Diagnostics
}

type TriggerSchedule struct {
	Trigger
	Schedule string `json:"schedule"`
}

func (t *TriggerSchedule) SetAttributes(hclAttributes hcl.Attributes, ctx *pipeparser.ParseContext) hcl.Diagnostics {
	diags := t.SetBaseAttributes(hclAttributes, ctx)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			val, _ := attr.Expr.Value(nil)
			t.Schedule = val.AsString()

			// validate cron format
			_, err := cron.ParseStandard(t.Schedule)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid cron expression: " + t.Schedule,
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}
		default:
			if !t.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Schedule: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

type TriggerInterval struct {
	Trigger
	Schedule string `json:"schedule"`
}

var validIntervals = []string{"hourly", "daily", "weekly", "monthly"}

func (t *TriggerInterval) SetAttributes(hclAttributes hcl.Attributes, ctx *pipeparser.ParseContext) hcl.Diagnostics {
	diags := t.SetBaseAttributes(hclAttributes, ctx)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			val, _ := attr.Expr.Value(nil)
			t.Schedule = val.AsString()

			if !helpers.StringSliceContains(validIntervals, strings.ToLower(t.Schedule)) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid interval",
					Detail:   "The interval must be one of: " + strings.Join(validIntervals, ","),
					Subject:  &attr.Range,
				})
			}

		default:
			if !t.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Interval: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

type TriggerQuery struct {
	Trigger
	Sql              string   `json:"sql"`
	Schedule         string   `json:"schedule"`
	ConnectionString string   `json:"connection_string"`
	PrimaryKey       string   `json:"primary_key"`
	Events           []string `json:"events"`
}

func (t *TriggerQuery) SetAttributes(hclAttributes hcl.Attributes, ctx *pipeparser.ParseContext) hcl.Diagnostics {
	diags := t.SetBaseAttributes(hclAttributes, ctx)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			val, _ := attr.Expr.Value(nil)
			t.Schedule = val.AsString()

			// validate cron format
			_, err := cron.ParseStandard(t.Schedule)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid cron expression: " + t.Schedule,
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}
		case schema.AttributeTypeSql:
			val, _ := attr.Expr.Value(nil)
			t.Sql = val.AsString()
		case schema.AttributeTypeConnectionString:
			val, _ := attr.Expr.Value(nil)
			t.ConnectionString = val.AsString()
		case schema.AttributeTypePrimaryKey:
			val, _ := attr.Expr.Value(nil)
			t.PrimaryKey = val.AsString()
		case schema.AttributeTypeEvents:
			val, _ := attr.Expr.Value(nil)
			var err error
			t.Events, err = hclhelpers.CtyTupleToArrayOfStrings(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + schema.AttributeTypeEvents + " Trigger attribute to Go values",
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}
		default:
			if !t.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Interval: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}
	return diags
}

type TriggerHttp struct {
	Trigger
}

func (t *TriggerHttp) SetAttributes(hclAttributes hcl.Attributes, ctx *pipeparser.ParseContext) hcl.Diagnostics {
	diags := t.SetBaseAttributes(hclAttributes, ctx)
	if diags.HasErrors() {
		return diags
	}

	return diags
}

func NewTrigger(ctx context.Context, triggerType, triggerName string) ITrigger {
	var trigger ITrigger

	switch triggerType {
	case schema.TriggerTypeSchedule:
		trigger = &TriggerSchedule{}
	case schema.TriggerTypeInterval:
		trigger = &TriggerInterval{}
	case schema.TriggerTypeQuery:
		trigger = &TriggerQuery{}
	case schema.TriggerTypeHttp:
		trigger = &TriggerHttp{}
	default:
		return nil
	}

	trigger.SetName(triggerName)
	trigger.SetContext(ctx)
	return trigger
}
