package modconfig

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"

	"github.com/robfig/cron/v3"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	ctx context.Context

	// 27/09/23 - Args is introduces combination of both parse time and runtime arguments. "var" should be resolved
	// at parse time, the vars all should be supplied when we start the system. However, args can also contain
	// runtime variable, i.e. self.request_body, self.rows
	//
	ArgsRaw hcl.Expression `json:"-"`

	Pipeline cty.Value `json:"-"`
	RawBody  hcl.Body  `json:"-" hcl:",remain"`

	Config ITriggerConfig `json:"-"`
}

func (p *Trigger) Equals(other *Trigger) bool {

	pipelineName := p.Pipeline.AsValueMap()[schema.LabelName].AsString()
	otherPipelineName := other.Pipeline.AsValueMap()[schema.LabelName].AsString()

	scheduleString := ""
	otherScheduleString := ""

	switch config := p.Config.(type) {
	case *TriggerSchedule:
		scheduleString = config.Schedule
	case *TriggerInterval:
		scheduleString = config.Schedule

	}

	switch config := other.Config.(type) {
	case *TriggerSchedule:
		otherScheduleString = config.Schedule
	case *TriggerInterval:
		otherScheduleString = config.Schedule
	}

	return p.FullName == other.FullName &&
		pipelineName == otherPipelineName &&
		scheduleString == otherScheduleString &&
		p.GetMetadata().ModFullName == other.GetMetadata().ModFullName
}

func (t *Trigger) GetPipeline() cty.Value {
	return t.Pipeline
}

func (t *Trigger) GetArgs(evalContext *hcl.EvalContext) (Input, hcl.Diagnostics) {

	if t.ArgsRaw == nil {
		return Input{}, hcl.Diagnostics{}
	}

	value, diags := t.ArgsRaw.Value(evalContext)

	if diags.HasErrors() {
		return Input{}, diags
	}

	retVal, err := hclhelpers.CtyToGoMapInterface(value)
	if err != nil {
		return Input{}, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute to Go values",
		}}
	}
	return retVal, diags
}

var ValidBaseTriggerAttributes = []string{
	schema.AttributeTypeDescription,
	schema.AttributeTypePipeline,
	schema.AttributeTypeArgs,
}

func (t *Trigger) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseTriggerAttributes, name)
}

func (t *Trigger) SetBaseAttributes(mod *Mod, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, moreDiags := hclhelpers.AttributeToString(attr, nil, false)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Description = desc
		}
	}

	// Pipeline is a required attribute, we don't need to validate it here because
	// it should be defined in the Trigger Schema
	attr := hclAttributes[schema.AttributeTypePipeline]

	expr := attr.Expr

	val, err := expr.Value(evalContext)
	if err != nil {
		// For Trigger's Pipeline reference, all it needs is the pipeline. It can't possibly use the output of a pipeline
		// so if the Pipeline is not parsed (yet) then the error message is:
		// Summary: "Unknown variable"
		// Detail: "There is no variable named \"pipeline\"."
		//
		// Do not unpack the error and create a new "Diagnostic", leave the original error message in
		// and let the "Mod processing" determine if there's an unresolved block

		// Don't error out, it's fine to unable to find the reference, we will try again later
		diags = append(diags, err...)
	} else {
		t.Pipeline = val
	}

	if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
		if attr.Expr != nil {
			t.ArgsRaw = attr.Expr
			// expr := attr.Expr
			// vals, moreDiags := expr.Value(evalContext)
			// if moreDiags != nil {
			// 	diags = append(diags, moreDiags...)
			// } else {
			// 	goVals, err := hclhelpers.CtyToGoMapInterface(vals)
			// 	if err != nil {
			// 		diags = append(diags, &hcl.Diagnostic{
			// 			Severity: hcl.DiagError,
			// 			Summary:  "Unable to parse " + schema.AttributeTypeArgs + " Trigger attribute to Go values",
			// 			Subject:  &attr.Range,
			// 		})
			// 	}
			// 	t.Args = goVals
			// }
		}
	}

	return diags
}

type ITriggerConfig interface {
	SetAttributes(*Mod, *Trigger, hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
}

type TriggerSchedule struct {
	Schedule string `json:"schedule"`
}

func (t *TriggerSchedule) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
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
			if !trigger.IsBaseAttribute(name) {
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
	Schedule string `json:"schedule"`
}

var validIntervals = []string{"hourly", "daily", "weekly", "monthly"}

func (t *TriggerInterval) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
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
			if !trigger.IsBaseAttribute(name) {
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
	Sql              string   `json:"sql"`
	Schedule         string   `json:"schedule"`
	ConnectionString string   `json:"connection_string"`
	PrimaryKey       string   `json:"primary_key"`
	Events           []string `json:"events"`
}

func (t *TriggerQuery) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
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
			if !trigger.IsBaseAttribute(name) {
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
	Url string `json:"url"`
}

func (t *TriggerHttp) SetAttributes(mod *Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		default:
			if !trigger.IsBaseAttribute(name) {
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

func NewTrigger(ctx context.Context, block *hcl.Block, mod *Mod, triggerType, triggerName string) *Trigger {

	triggerFullName := triggerType + "." + triggerName

	if mod != nil {
		modName := mod.Name()
		if strings.HasPrefix(modName, "mod") {
			modName = strings.TrimPrefix(modName, "mod.")
		}
		triggerFullName = modName + ".trigger." + triggerFullName
	} else {
		triggerFullName = "local.trigger." + triggerFullName
	}

	trigger := &Trigger{
		HclResourceImpl: HclResourceImpl{
			FullName:        triggerFullName,
			UnqualifiedName: "trigger." + triggerName,
			DeclRange:       block.DefRange,
			blockType:       block.Type,
		},
		ctx: ctx,
	}

	switch triggerType {
	case schema.TriggerTypeSchedule:
		trigger.Config = &TriggerSchedule{}
	case schema.TriggerTypeInterval:
		trigger.Config = &TriggerInterval{}
	case schema.TriggerTypeQuery:
		trigger.Config = &TriggerQuery{}
	case schema.TriggerTypeHttp:
		trigger.Config = &TriggerHttp{}
	default:
		return nil
	}

	return trigger
}

// GetTriggerTypeFromTriggerConfig returns the type of the trigger from the trigger config
func GetTriggerTypeFromTriggerConfig(config ITriggerConfig) string {
	switch config.(type) {
	case *TriggerSchedule:
		return schema.TriggerTypeSchedule
	case *TriggerInterval:
		return schema.TriggerTypeInterval
	case *TriggerQuery:
		return schema.TriggerTypeQuery
	case *TriggerHttp:
		return schema.TriggerTypeHttp
	}

	return ""
}
