package resources

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/cty_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func CustomValueValidation(name string, setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	// this time we check if the given setting, i.e.
	// name = "example
	// type = "aws"

	// for connection actually exists in the eval context

	if hclhelpers.IsListLike(setting.Type()) {
		return pipelineParamCustomValueListValidation(name, setting, evalCtx)
	}

	if !hclhelpers.IsMapLike(setting.Type()) {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "The value for param must be an object: " + name,
		}
		return hcl.Diagnostics{diag}
	}

	settingValueMap := setting.AsValueMap()

	// get resource type (if present)
	valueResourceType, _ := cty_helpers.StringValueFromCtyMap(settingValueMap, "resource_type")
	valueType, _ := cty_helpers.StringValueFromCtyMap(settingValueMap, "type")

	if connection.ConnectionTypeMeetsRequiredType(schema.BlockTypeConnection, valueResourceType, valueType) {
		if settingValueMap["type"].IsNull() {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "The value for param must have a 'type' key: " + name,
			}
			return hcl.Diagnostics{diag}
		}

		// check if the connection actually exists in the eval context
		allConnections := evalCtx.Variables[schema.BlockTypeConnection]
		if allConnections == cty.NilVal {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "No connection found: " + name,
			}
			return hcl.Diagnostics{diag}
		}
		connectionType, ok := cty_helpers.StringValueFromCtyMap(settingValueMap, "type")
		if !ok {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "The value for param must have a 'type' key: " + name,
			}}
		}
		connectionName, ok := cty_helpers.StringValueFromCtyMap(settingValueMap, "short_name")
		if !ok {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "The value for param must have a 'short_name' key: " + name,
			}}
		}

		if allConnections.Type().IsMapType() || allConnections.Type().IsObjectType() {
			allConnectionsMap := allConnections.AsValueMap()
			if allConnectionsMap[connectionType].IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No connection found for the given connection type: " + connectionType,
				}
				return hcl.Diagnostics{diag}
			}

			connectionTypeMap := allConnectionsMap[connectionType].AsValueMap()
			if connectionTypeMap[connectionName].IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No connection found for the given connection name: " + connectionName,
				}
				return hcl.Diagnostics{diag}
			} else {
				// TRUE
				return hcl.Diagnostics{}
			}
		}
	} else if valueResourceType == schema.BlockTypeNotifier {
		// check if the connection actually exists in the eval context
		allNotifiers := evalCtx.Variables[schema.BlockTypeNotifier]
		if allNotifiers == cty.NilVal {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "No notifier found: " + name,
			}
			return hcl.Diagnostics{diag}
		}

		notifierName, ok := cty_helpers.StringValueFromCtyMap(settingValueMap, "name")
		if !ok {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "The value for param must have a 'name' key: " + name,
			}}
		}

		if allNotifiers.Type().IsMapType() || allNotifiers.Type().IsObjectType() {
			allNotifiersMap := allNotifiers.AsValueMap()

			if allNotifiersMap[notifierName].IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "No notifier found for the given notifier name: " + notifierName,
				}
				return hcl.Diagnostics{diag}
			} else {
				// TRUE
				return hcl.Diagnostics{}
			}
		}
	} else if len(settingValueMap) > 0 {
		diags := hcl.Diagnostics{}
		for _, v := range settingValueMap {
			if v.IsNull() {
				diag := &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "The value for param must not have a null value: " + name,
				}
				return hcl.Diagnostics{diag}
			}

			if !hclhelpers.IsComplexType(v.Type()) {
				// this test is meant for custom value validation, there's no need to test if it's not these type, i.e. connection or notifier
				continue
			}

			// this test is meant for custom value validation, there's no need to test if it's not these type, i.e. connection or notifier
			nestedDiags := CustomValueValidation(name, v, evalCtx)
			diags = append(diags, nestedDiags...)
		}

		return diags
	}

	diag := &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid value for param " + name,
		Detail:   "Invalid value for param " + name,
	}
	return hcl.Diagnostics{diag}
}

func pipelineParamCustomValueListValidation(name string, setting cty.Value, evalCtx *hcl.EvalContext) hcl.Diagnostics {

	if !hclhelpers.IsListLike(setting.Type()) {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid value for param " + name,
			Detail:   "The value for param must be a list",
		}
		return hcl.Diagnostics{diag}
	}

	var diags hcl.Diagnostics
	for it := setting.ElementIterator(); it.Next(); {
		_, element := it.Element()
		diags = append(diags, CustomValueValidation(name, element, evalCtx)...)
	}

	return diags
}
