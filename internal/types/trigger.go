package types

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl/v2"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	ctx         context.Context
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty" hcl:"description,optional" cty:"description"`

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

func (p *Trigger) SetBaseAttributes(hclAttributes hcl.Attributes, parseContext *pipeparser.ParseContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, diag := hclhelpers.AttributeToString(attr, nil, false)
		if diag != nil && diag.Severity == hcl.DiagError {
			diags = append(diags, diag)
		} else {
			p.Description = desc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypePipeline]; exists {
		if attr.Expr != nil {
			expr := attr.Expr
			val, err := expr.Value(parseContext.EvalCtx)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unable to parse " + schema.AttributeTypePipeline + " attribute: " + err.Error(),
					Subject:  &attr.Range,
				})
			} else {
				p.Pipeline = val
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
		}
	}
	return nil
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
		}
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
	default:
		return nil
	}

	trigger.SetName(triggerName)
	trigger.SetContext(ctx)
	return trigger
}

type PrintableTrigger struct {
	Items interface{}
}

func (PrintableTrigger) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {

	return nil, nil
}

func (p PrintableTrigger) GetItems() interface{} {
	return p.Items
}

func (p PrintableTrigger) GetTable() (Table, error) {
	return Table{}, nil
}

func (PrintableTrigger) GetColumns() (columns []TableColumnDefinition) {
	return []TableColumnDefinition{
		{
			Name:        "TYPE",
			Type:        "string",
			Description: "The type of the trigger",
		},
		{
			Name:        "NAME",
			Type:        "string",
			Description: "The name of the trigger",
		},
	}
}

// This type is used by the API to return a list of triggers.
type ListTriggerResponse struct {
	Items     []Trigger `json:"items"`
	NextToken *string   `json:"next_token,omitempty"`
}
