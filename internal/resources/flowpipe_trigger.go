package resources

import (
	"github.com/turbot/pipe-fittings/modconfig"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"

	"github.com/robfig/cron/v3"
)

// The definition of a single Flowpipe Trigger
type Trigger struct {
	modconfig.HclResourceImpl
	modconfig.ResourceWithMetadataImpl

	mod *modconfig.Mod

	FileName        string          `json:"file_name"`
	StartLineNumber int             `json:"start_line_number"`
	EndLineNumber   int             `json:"end_line_number"`
	Params          []PipelineParam `json:"params,omitempty"`

	// 27/09/23 - Args is a combination of both parse time and runtime arguments. "var" should be resolved
	// at parse time, the vars all should be supplied when we start the system.
	//
	// However, args can also contain runtime variable, i.e. self.request_body, self.rows
	//
	// Args are not currently validated at parse time. To validate Args at parse time we need to:
	// - identity which args are parse time (var and param)
	// - validate those parse time args
	//
	ArgsRaw hcl.Expression `json:"-"`

	// TODO: 2024/01/09 - change of direction with Trigger schema, pipeline no longer "common" because Query Trigger no longer has a single pipeline for
	// all its events, similarly for HTTP trigger the pipeline is being moved down to the "method" block.
	Pipeline cty.Value     `json:"-"`
	RawBody hcl.Body      `json:"-" hcl:",remain"`
	Config  TriggerConfig `json:"-"`
	Enabled *bool         `json:"-"`
}

// Implements the ModTreeItem interface
func (t *Trigger) GetMod() *modconfig.Mod {
	return t.mod
}

func (t *Trigger) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	t.FileName = fileName
	t.StartLineNumber = startLineNumber
	t.EndLineNumber = endLineNumber
}

func (t *Trigger) GetParam(paramName string) *PipelineParam {
	for _, param := range t.Params {
		if param.Name == paramName {
			return &param
		}
	}
	return nil
}

func (t *Trigger) GetParams() []PipelineParam {
	return t.Params
}

func (t *Trigger) Equals(other *Trigger) bool {
	if t == nil && other == nil {
		return true
	}

	if t == nil && other != nil || t != nil && other == nil {
		return false
	}

	baseEqual := t.HclResourceImpl.Equals(&t.HclResourceImpl)
	if !baseEqual {
		return false
	}

	// Order of params does not matter, but the value does
	if len(t.Params) != len(other.Params) {
		return false
	}

	// Compare param values
	for _, v := range t.Params {
		otherParam := other.GetParam(v.Name)
		if otherParam == nil {
			return false
		}

		if !v.Equals(otherParam) {
			return false
		}
	}

	// catch name change of the other param
	for _, v := range other.Params {
		pParam := t.GetParam(v.Name)
		if pParam == nil {
			return false
		}
	}

	if !utils.BoolPtrEqual(t.Enabled, other.Enabled) {
		return false
	}

	if t.Pipeline.Equals(other.Pipeline).False() {
		return false
	}

	if !reflect.DeepEqual(t.ArgsRaw, other.ArgsRaw) {
		return false
	}

	if t.Config == nil && !helpers.IsNil(other.Config) || t.Config != nil && helpers.IsNil(other.Config) {
		return false
	}

	if !t.Config.Equals(other.Config) {
		return false
	}

	return t.FullName == other.FullName &&
		t.GetMetadata().ModFullName == other.GetMetadata().ModFullName
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
	schema.AttributeTypeTitle,
	schema.AttributeTypeDocumentation,
	schema.AttributeTypeTags,
	schema.AttributeTypeEnabled,
}

func (t *Trigger) IsBaseAttribute(name string) bool {
	return slices.Contains[[]string, string](ValidBaseTriggerAttributes, name)
}

func (t *Trigger) SetBaseAttributes(mod *modconfig.Mod, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	var diags hcl.Diagnostics

	if attr, exists := hclAttributes[schema.AttributeTypeDescription]; exists {
		desc, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Description = desc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTitle]; exists {
		title, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Title = title
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeDocumentation]; exists {
		doc, moreDiags := hclhelpers.AttributeToString(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Documentation = doc
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeTags]; exists {
		tags, moreDiags := hclhelpers.AttributeToMap(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			resultMap := make(map[string]string)
			for key, value := range tags {
				resultMap[key] = value.(string)
			}
			t.Tags = resultMap
		}
	}

	// TODO: this is now only relevant for Schedule Trigger, move it to the Schedule Trigger
	attr := hclAttributes[schema.AttributeTypePipeline]
	if attr != nil {
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
	}

	if attr, exists := hclAttributes[schema.AttributeTypeEnabled]; exists {
		triggerEnabled, moreDiags := hclhelpers.AttributeToBool(attr, evalContext, true)
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			t.Enabled = triggerEnabled
		}
	}

	if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
		if attr.Expr != nil {
			t.ArgsRaw = attr.Expr
		}
	}

	return diags
}

type TriggerConfig interface {
	SetAttributes(*modconfig.Mod, *Trigger, hcl.Attributes, *hcl.EvalContext) hcl.Diagnostics
	GetUnresolvedAttributes() map[string]hcl.Expression
	SetBlocks(*modconfig.Mod, *Trigger, hcl.Blocks, *hcl.EvalContext) hcl.Diagnostics
	Equals(other TriggerConfig) bool
	GetType() string
	GetConfig(*hcl.EvalContext, *modconfig.Mod) (TriggerConfig, error)
	GetConnectionDependsOn() []string
}

type TriggerSchedule struct {
	Schedule             string                    `json:"schedule"`
	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	ConnectionDependsOn  []string                  `json:"connection_depends_on,omitempty"`
}

func (t *TriggerSchedule) GetConfig(evalContext *hcl.EvalContext, mod *modconfig.Mod) (TriggerConfig, error) {
	return t, nil
}

func (t *TriggerSchedule) AppendsDependsOn(dependsOn ...string) {
}

func (t *TriggerSchedule) AppendCredentialDependsOn(...string) {
}

func (t *TriggerSchedule) AppendConnectionDependsOn(connectionDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]struct{}, len(t.ConnectionDependsOn))
	for _, dep := range t.ConnectionDependsOn {
		existingDeps[dep] = struct{}{}
	}

	for _, dep := range connectionDependsOn {
		if _, exists := existingDeps[dep]; !exists {
			t.ConnectionDependsOn = append(t.ConnectionDependsOn, dep)
			existingDeps[dep] = struct{}{}
		}
	}
}

func (t *TriggerSchedule) GetConnectionDependsOn() []string {
	return t.ConnectionDependsOn
}

func (t *TriggerSchedule) AddUnresolvedAttribute(key string, value hcl.Expression) {
	t.UnresolvedAttributes[key] = value
}

func (t *TriggerSchedule) GetPipeline() *Pipeline {
	return nil
}

func (t *TriggerSchedule) GetUnresolvedAttributes() map[string]hcl.Expression {
	return t.UnresolvedAttributes
}

func (t *TriggerSchedule) GetType() string {
	return schema.TriggerTypeSchedule
}

func (t *TriggerSchedule) Equals(other TriggerConfig) bool {
	otherTrigger, ok := other.(*TriggerSchedule)
	if !ok {
		return false
	}

	if t == nil && !helpers.IsNil(otherTrigger) || t != nil && helpers.IsNil(otherTrigger) {
		return false
	}

	if t == nil && helpers.IsNil(otherTrigger) {
		return true
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(t.UnresolvedAttributes) != len(other.GetUnresolvedAttributes()) {
		return false
	}

	for key, expr := range t.UnresolvedAttributes {
		otherExpr, ok := other.GetUnresolvedAttributes()[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	return t.Schedule == otherTrigger.Schedule
}

func (t *TriggerSchedule) SetAttributes(mod *modconfig.Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			// schedule should never be an unresolved variable, it needs to be fully resolved
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			if val.Type() != cty.String {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The given schedule is not a string",
					Detail:   "The given schedule is not a string",
					Subject:  &attr.Range,
				})
				continue
			}

			t.Schedule = val.AsString()

			if slices.Contains(validIntervals, t.Schedule) {
				continue
			}

			// if it's not an interval, assume it's a cron and attempt to validate the cron expression
			_, err := cron.ParseStandard(t.Schedule)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid cron expression: " + t.Schedule + ". Specify valid intervals hourly, daily, weekly, monthly or valid cron expression",
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

func (t *TriggerSchedule) SetBlocks(mod *modconfig.Mod, trigger *Trigger, hclBlocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}
	return diags
}

var validIntervals = []string{"hourly", "daily", "weekly", "5m", "10m", "15m", "30m", "60m", "1h", "2h", "4h", "6h", "12h", "24h"}

type TriggerQuery struct {
	Sql        string                          `json:"sql"`
	Schedule   string                          `json:"schedule"`
	Database   string                          `json:"database"`
	PrimaryKey string                          `json:"primary_key"`
	Captures   map[string]*TriggerQueryCapture `json:"captures"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	ConnectionDependsOn  []string                  `json:"connection_depends_on,omitempty"`
}

func (t *TriggerQuery) AppendDependsOn(...string) {
}

func (t *TriggerQuery) AppendCredentialDependsOn(...string) {
}

func (t *TriggerQuery) AppendConnectionDependsOn(connectionDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]struct{}, len(t.ConnectionDependsOn))
	for _, dep := range t.ConnectionDependsOn {
		existingDeps[dep] = struct{}{}
	}

	for _, dep := range connectionDependsOn {
		if _, exists := existingDeps[dep]; !exists {
			t.ConnectionDependsOn = append(t.ConnectionDependsOn, dep)
			existingDeps[dep] = struct{}{}
		}
	}
}

func (t *TriggerQuery) GetConnectionDependsOn() []string {
	return t.ConnectionDependsOn
}

func (t *TriggerQuery) AddUnresolvedAttribute(key string, value hcl.Expression) {
	t.UnresolvedAttributes[key] = value
}

func (t *TriggerQuery) GetPipeline() *Pipeline {
	return nil
}

func (t *TriggerQuery) GetUnresolvedAttributes() map[string]hcl.Expression {
	return t.UnresolvedAttributes
}

func (t *TriggerQuery) GetType() string {
	return schema.TriggerTypeQuery
}

func (t *TriggerQuery) GetConfig(evalContext *hcl.EvalContext, mod *modconfig.Mod) (TriggerConfig, error) {

	var database string

	if databaseExpression, ok := t.UnresolvedAttributes[schema.AttributeTypeDatabase]; ok {
		// attribute needs resolving, this case may happen if we specify the entire option as an attribute
		var dbValue cty.Value
		diags := gohcl.DecodeExpression(databaseExpression, evalContext, &dbValue)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError("query_trigger", diags)
		}
		// check if this is a connection string or a connection
		if dbValue.Type() == cty.String {
			database = dbValue.AsString()
		} else {
			c, err := app_specific_connection.CtyValueToConnection(dbValue)
			if err != nil {
				return nil, perr.BadRequestWithMessage("unable to resolve connection attribute: " + err.Error())
			}
			if conn, ok := c.(connection.ConnectionStringProvider); ok {
				database = conn.GetConnectionString()
			} else {
				slog.Warn("connection does not support connection string", "db", c)
				return nil, perr.BadRequestWithMessage("invalid connection reference - only connections which implement GetConnectionString() are supported")
			}
		}
	} else {
		var diags hcl.Diagnostics
		database, diags = simpleOutputFromAttribute(t.GetUnresolvedAttributes(), evalContext, schema.AttributeTypeDatabase, t.Database)
		if diags.HasErrors() {
			return nil, error_helpers.BetterHclDiagsToError("query trigger", diags)
		}
	}

	// if no database is set, get the default database from the mod
	if database == "" {
		var err error
		database, err = mod.GetDefaultConnectionString(evalContext)
		if err != nil {
			return nil, err
		}
	}

	sql, diags := simpleOutputFromAttribute(t.GetUnresolvedAttributes(), evalContext, schema.AttributeTypeSql, t.Sql)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError("query trigger", diags)
	}

	schedule, diags := simpleOutputFromAttribute(t.GetUnresolvedAttributes(), evalContext, schema.AttributeTypeSchedule, t.Schedule)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError("query trigger", diags)
	}

	primaryKey, diags := simpleOutputFromAttribute(t.GetUnresolvedAttributes(), evalContext, schema.AttributeTypePrimaryKey, t.PrimaryKey)
	if diags.HasErrors() {
		return nil, error_helpers.BetterHclDiagsToError("query trigger", diags)
	}

	newT := &TriggerQuery{
		Sql:        sql,
		Schedule:   schedule,
		Database:   database,
		PrimaryKey: primaryKey,
		Captures:   make(map[string]*TriggerQueryCapture),
	}

	for key, value := range t.Captures {
		newT.Captures[key] = &TriggerQueryCapture{
			Type:     value.Type,
			Pipeline: value.Pipeline,
			ArgsRaw:  value.ArgsRaw,
		}
	}

	return newT, nil
}

func (t *TriggerQuery) Equals(other TriggerConfig) bool {
	otherTrigger, ok := other.(*TriggerQuery)
	if !ok {
		return false
	}

	if t == nil && !helpers.IsNil(otherTrigger) || t != nil && helpers.IsNil(otherTrigger) {
		return false
	}

	if t == nil && helpers.IsNil(otherTrigger) {
		return true
	}

	if t.Sql != otherTrigger.Sql {
		return false
	}

	if t.Schedule != otherTrigger.Schedule {
		return false
	}

	if t.Database != otherTrigger.Database {
		return false
	}

	if t.PrimaryKey != otherTrigger.PrimaryKey {
		return false
	}

	if len(t.Captures) != len(otherTrigger.Captures) {
		return false
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(t.UnresolvedAttributes) != len(other.GetUnresolvedAttributes()) {
		return false
	}

	for key, expr := range t.UnresolvedAttributes {
		otherExpr, ok := other.GetUnresolvedAttributes()[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	for key, value := range t.Captures {
		otherValue, exists := otherTrigger.Captures[key]
		if !exists {
			return false
		}

		if !value.Equals(otherValue) {
			return false
		}
	}

	return true
}

type TriggerQueryCapture struct {
	Type     string
	Pipeline cty.Value
	ArgsRaw  hcl.Expression
}

func (c *TriggerQueryCapture) Equals(other *TriggerQueryCapture) bool {
	if c == nil && other == nil {
		return true
	}

	if c == nil && other != nil || c != nil && other == nil {
		return false
	}

	if c.Type != other.Type {
		return false
	}

	if c.Pipeline.Equals(other.Pipeline).False() {
		return false
	}

	if !reflect.DeepEqual(c.ArgsRaw, other.ArgsRaw) {
		return false
	}

	return true
}

func (t *TriggerQuery) SetAttributes(mod *modconfig.Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	// TriggerQuery is the only trigger that supports params for: database, sql and primary key. The other triggers don't have
	// a use case for param support.

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeSchedule:
			// schedule should never be an unresolved variable, it needs to be fully resolved
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			if val.Type() != cty.String {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The given schedule is not a string",
					Detail:   "The given schedule is not a string",
					Subject:  &attr.Range,
				})
				continue
			}

			t.Schedule = val.AsString()

			if slices.Contains(validIntervals, t.Schedule) {
				continue
			}

			// if it's not an interval, assume it's a cron and attempt to validate the cron expression
			_, err := cron.ParseStandard(t.Schedule)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid cron expression: " + t.Schedule + ". Specify valid intervals hourly, daily, weekly, monthly or valid cron expression",
					Detail:   err.Error(),
					Subject:  &attr.Range,
				})
			}

		case schema.AttributeTypeSql:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, t)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				sql, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSchedule + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				t.Sql = sql
			}

		case schema.AttributeTypeDatabase:
			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, t, structFieldName, false)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypePrimaryKey:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, t)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				primaryKey, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeSchedule + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				t.PrimaryKey = primaryKey
			}

		default:
			if !trigger.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Trigger Query: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

var validCaptureBlockTypes = []string{"insert", "update", "delete"}

func (t *TriggerQuery) SetBlocks(mod *modconfig.Mod, trigger *Trigger, hclBlocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	t.Captures = make(map[string]*TriggerQueryCapture)

	for _, captureBlock := range hclBlocks {

		if captureBlock.Type != schema.BlockTypeCapture {
			continue
		}

		if len(captureBlock.Labels) != 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid capture block",
				Detail:   "Capture block must have a single label",
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		captureBlockType := captureBlock.Labels[0]
		if !slices.Contains(validCaptureBlockTypes, captureBlockType) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid capture block type",
				Detail:   "Capture block type must be one of: " + strings.Join(validCaptureBlockTypes, ","),
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		hclAttributes, moreDiags := captureBlock.Body.JustAttributes()
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		attr := hclAttributes[schema.AttributeTypePipeline]
		if attr == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Pipeline attribute is required for capture block",
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		if t.Captures[captureBlockType] != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate capture block",
				Detail:   "Duplicate capture block for type: " + captureBlockType,
				Subject:  &captureBlock.DefRange,
			})
			continue
		}

		triggerCapture := &TriggerQueryCapture{
			Type: captureBlockType,
		}

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
			triggerCapture.Pipeline = val
		}

		if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
			if attr.Expr != nil {
				triggerCapture.ArgsRaw = attr.Expr
			}
		}

		t.Captures[captureBlockType] = triggerCapture
	}

	return diags
}

func (c *TriggerQueryCapture) GetArgs(evalContext *hcl.EvalContext) (Input, hcl.Diagnostics) {

	if c.ArgsRaw == nil {
		return Input{}, hcl.Diagnostics{}
	}

	value, diags := c.ArgsRaw.Value(evalContext)

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

type TriggerHttp struct {
	Url           string                        `json:"url"`
	ExecutionMode string                        `json:"execution_mode"`
	Methods       map[string]*TriggerHTTPMethod `json:"methods"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
	ConnectionDependsOn  []string                  `json:"connection_depends_on,omitempty"`
}

func (t *TriggerHttp) AppendsDependsOn(dependsOn ...string) {
}

func (t *TriggerHttp) AppendCredentialDependsOn(...string) {
}

func (t *TriggerHttp) AppendConnectionDependsOn(connectionDependsOn ...string) {
	// Use map to track existing DependsOn, this will make the lookup below much faster
	// rather than using nested loops
	existingDeps := make(map[string]struct{}, len(t.ConnectionDependsOn))
	for _, dep := range t.ConnectionDependsOn {
		existingDeps[dep] = struct{}{}
	}

	for _, dep := range connectionDependsOn {
		if _, exists := existingDeps[dep]; !exists {
			t.ConnectionDependsOn = append(t.ConnectionDependsOn, dep)
			existingDeps[dep] = struct{}{}
		}
	}
}

func (t *TriggerHttp) GetConnectionDependsOn() []string {
	return t.ConnectionDependsOn
}

func (t *TriggerHttp) AddUnresolvedAttribute(key string, value hcl.Expression) {
	t.UnresolvedAttributes[key] = value
}

func (t *TriggerHttp) GetPipeline() *Pipeline {
	return nil
}

func (t *TriggerHttp) GetUnresolvedAttributes() map[string]hcl.Expression {
	return t.UnresolvedAttributes
}

func (t *TriggerHttp) GetType() string {
	return schema.TriggerTypeHttp
}

func (t *TriggerHttp) GetConfig(*hcl.EvalContext, *modconfig.Mod) (TriggerConfig, error) {
	return t, nil
}

func (t *TriggerHttp) Equals(other TriggerConfig) bool {
	otherTrigger, ok := other.(*TriggerHttp)
	if !ok {
		return false
	}

	if t == nil && !helpers.IsNil(otherTrigger) || t != nil && helpers.IsNil(otherTrigger) {
		return false
	}

	if t == nil && helpers.IsNil(otherTrigger) {
		return true
	}

	if t.Url != otherTrigger.Url {
		return false
	}

	if t.ExecutionMode != otherTrigger.ExecutionMode {
		return false
	}

	if len(t.Methods) != len(otherTrigger.Methods) {
		return false
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(t.UnresolvedAttributes) != len(other.GetUnresolvedAttributes()) {
		return false
	}

	for key, expr := range t.UnresolvedAttributes {
		otherExpr, ok := other.GetUnresolvedAttributes()[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	for key, value := range t.Methods {
		otherValue, exists := otherTrigger.Methods[key]
		if !exists {
			return false
		}

		if !value.Equals(otherValue) {
			return false
		}
	}

	return true
}

type TriggerHTTPMethod struct {
	Type          string
	ExecutionMode string
	Pipeline      cty.Value
	ArgsRaw       hcl.Expression
}

func (c *TriggerHTTPMethod) Equals(other *TriggerHTTPMethod) bool {
	if c == nil && other == nil {
		return true
	}

	if c == nil && other != nil || c != nil && other == nil {
		return false
	}

	if c.Type != other.Type || c.ExecutionMode != other.ExecutionMode {
		return false
	}

	if c.Pipeline.Equals(other.Pipeline).False() {
		return false
	}

	if !reflect.DeepEqual(c.ArgsRaw, other.ArgsRaw) {
		return false
	}

	return true
}

var validExecutionMode = []string{"synchronous", "asynchronous"}
var validMethodBlockTypes = []string{"post", "get"}

func (t *TriggerHttp) SetAttributes(mod *modconfig.Mod, trigger *Trigger, hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	// None of the Trigger Http attributes should be unresolved at parse time. It doesn't make sense to have params for the URL for example

	diags := trigger.SetBaseAttributes(mod, hclAttributes, evalContext)
	if diags.HasErrors() {
		return diags
	}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeExecutionMode:
			val, moreDiags := attr.Expr.Value(evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			if val.Type() != cty.String {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The given execution mode is not a string",
					Detail:   "The given execution mode is not a string",
					Subject:  &attr.Range,
				})
				continue
			}

			t.ExecutionMode = val.AsString()

			if !slices.Contains(validExecutionMode, t.ExecutionMode) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid execution mode",
					Detail:   "The execution mode must be one of: " + strings.Join(validExecutionMode, ","),
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

func (t *TriggerHttp) SetBlocks(mod *modconfig.Mod, trigger *Trigger, hclBlocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	t.Methods = make(map[string]*TriggerHTTPMethod)

	// If no method blocks appear, only 'post' is supported, and the top-level `pipeline`, `args` and `execution_mode` will be applied
	if len(hclBlocks) == 0 {
		triggerMethod := &TriggerHTTPMethod{
			Type: HttpMethodPost,
		}

		// Get the top-level pipeline
		pipeline := trigger.GetPipeline()
		if pipeline == cty.NilVal {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Bad Request",
				Detail:   "Missing required attribute 'pipeline'",
			})
			return diags
		}
		triggerMethod.Pipeline = pipeline

		// Get the top-level args
		pipelineArgs := trigger.ArgsRaw
		if pipelineArgs != nil {
			triggerMethod.ArgsRaw = pipelineArgs
		}

		// Get the top-level execution_mode
		if t.ExecutionMode != "" {
			triggerMethod.ExecutionMode = t.ExecutionMode
		}

		t.Methods[HttpMethodPost] = triggerMethod

		return diags
	}

	// If the method blocks provided, we will consider the configuration provided in the method block
	for _, methodBlock := range hclBlocks {

		if len(methodBlock.Labels) != 1 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid method block",
				Detail:   "Method block must have a single label",
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		methodBlockType := methodBlock.Labels[0]
		if !slices.Contains(validMethodBlockTypes, methodBlockType) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid method block type",
				Detail:   "Method block type must be one of: " + strings.Join(validMethodBlockTypes, ","),
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		hclAttributes, moreDiags := methodBlock.Body.JustAttributes()
		if moreDiags != nil && moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}

		attr := hclAttributes[schema.AttributeTypePipeline]
		if attr == nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Pipeline attribute is required for method block",
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		if t.Methods[methodBlockType] != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate method block",
				Detail:   "Duplicate method block for type: " + methodBlockType,
				Subject:  &methodBlock.DefRange,
			})
			continue
		}

		triggerMethod := &TriggerHTTPMethod{
			Type: methodBlockType,
		}

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
			triggerMethod.Pipeline = val
		}

		if attr, exists := hclAttributes[schema.AttributeTypeArgs]; exists {
			if attr.Expr != nil {
				triggerMethod.ArgsRaw = attr.Expr
			}
		}

		if attr, exists := hclAttributes[schema.AttributeTypeExecutionMode]; exists {
			if attr.Expr != nil {
				val, err := attr.Expr.Value(evalContext)
				if err != nil {
					diags = append(diags, err...)
				}

				executionMode, ctyErr := hclhelpers.CtyToString(val)
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeExecutionMode + " attribute to string",
						Subject:  &attr.Range,
					})
				}

				if !slices.Contains(validExecutionMode, executionMode) {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid execution mode",
						Detail:   "The execution mode must be one of: " + strings.Join(validExecutionMode, ","),
						Subject:  &attr.Range,
					})
				}

				triggerMethod.ExecutionMode = executionMode
			}
		}

		t.Methods[methodBlockType] = triggerMethod
	}

	return diags
}

func (c *TriggerHTTPMethod) GetArgs(evalContext *hcl.EvalContext) (Input, hcl.Diagnostics) {

	if c.ArgsRaw == nil {
		return Input{}, hcl.Diagnostics{}
	}

	value, diags := c.ArgsRaw.Value(evalContext)

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

func NewTrigger(block *hcl.Block, mod *modconfig.Mod, triggerType, triggerName string) *Trigger {

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
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName:        triggerFullName,
			UnqualifiedName: "trigger." + triggerName,
			DeclRange:       block.DefRange,
			BlockType:       block.Type,
		},
		mod: mod,
	}

	switch triggerType {
	case schema.TriggerTypeSchedule:
		trigger.Config = &TriggerSchedule{
			UnresolvedAttributes: make(map[string]hcl.Expression),
		}
	case schema.TriggerTypeQuery:
		trigger.Config = &TriggerQuery{
			UnresolvedAttributes: make(map[string]hcl.Expression),
		}
	case schema.TriggerTypeHttp:
		trigger.Config = &TriggerHttp{
			UnresolvedAttributes: make(map[string]hcl.Expression),
		}
	default:
		return nil
	}

	return trigger
}

// GetTriggerTypeFromTriggerConfig returns the type of the trigger from the trigger config
func GetTriggerTypeFromTriggerConfig(config TriggerConfig) string {
	switch config.(type) {
	case *TriggerSchedule:
		return schema.TriggerTypeSchedule
	case *TriggerQuery:
		return schema.TriggerTypeQuery
	case *TriggerHttp:
		return schema.TriggerTypeHttp
	}

	return ""
}
