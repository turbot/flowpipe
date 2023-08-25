package hclhelpers

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

func AttributeToString(attr *hcl.Attribute, evalContext *hcl.EvalContext, allowExpression bool) (*string, hcl.Diagnostics) {
	if attr.Expr == nil {
		return nil, nil
	}

	expr := attr.Expr
	if len(expr.Variables()) > 0 && !allowExpression {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expression not allowed in" + attr.Name,
			Subject:  &attr.Range,
		}}
	}

	val, err := attr.Expr.Value(evalContext)

	if err != nil {
		return nil, err
	}

	if val.Type() != cty.String {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + attr.Name + " attribute as string",
			Subject:  &attr.Range,
		}}
	}

	stringValue := val.AsString()
	return &stringValue, hcl.Diagnostics{}
}

func AttributeToInt(attr *hcl.Attribute, evalContext *hcl.EvalContext, allowExpression bool) (*int64, hcl.Diagnostics) {
	if attr.Expr == nil {
		return nil, nil
	}

	expr := attr.Expr
	if len(expr.Variables()) > 0 && !allowExpression {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expression not allowed in" + attr.Name,
			Subject:  &attr.Range,
		}}
	}

	val, err := attr.Expr.Value(evalContext)

	if err != nil {
		return nil, err
	}

	if val.Type() != cty.Number {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + attr.Name + " attribute as number",
			Subject:  &attr.Range,
		}}
	}

	bigFloatValue := val.AsBigFloat()

	if !bigFloatValue.IsInt() {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + attr.Name + " attribute as int",
			Subject:  &attr.Range,
		}}
	}

	int64Value, _ := bigFloatValue.Int64()
	return &int64Value, hcl.Diagnostics{}
}

func AttributeToBool(attr *hcl.Attribute, evalContext *hcl.EvalContext, allowExpression bool) (*bool, hcl.Diagnostics) {
	if attr.Expr == nil {
		return nil, nil
	}

	expr := attr.Expr
	if len(expr.Variables()) > 0 && !allowExpression {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expression not allowed in" + attr.Name,
			Subject:  &attr.Range,
		}}
	}

	val, err := attr.Expr.Value(evalContext)

	if err != nil {
		return nil, err
	}

	if val.Type() != cty.Bool {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + attr.Name + " attribute as boolean",
			Subject:  &attr.Range,
		}}
	}

	if val.True() {
		res := true
		return &res, hcl.Diagnostics{}
	}

	res := false
	return &res, hcl.Diagnostics{}
}

func AttributeToMap(attr *hcl.Attribute, evalContext *hcl.EvalContext, allowExpression bool) (map[string]interface{}, hcl.Diagnostics) {
	if attr.Expr == nil {
		return nil, nil
	}

	diags := hcl.Diagnostics{}

	expr := attr.Expr
	if len(expr.Variables()) > 0 && !allowExpression {

		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Expression not allowed in" + attr.Name,
			Detail:   "Expression not allowed in" + attr.Name + " and we have expression",
			Subject:  &attr.Range,
		}}
	}

	val, err := attr.Expr.Value(evalContext)

	if err != nil {
		diags = append(diags, err...)
	}

	if !val.Type().IsObjectType() {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + attr.Name + " attribute as map",
			Detail:   "Type is: " + val.Type().FriendlyName() + " which is not an Object type",
			Subject:  &attr.Range,
		}}
	}

	valMap := val.AsValueMap()

	if valMap == nil {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse " + attr.Name + " attribute as map",
			Detail:   "Unable to parse " + attr.Name + " attribute as map",
			Subject:  &attr.Range,
		}}
	}

	res := make(map[string]interface{})
	for k, v := range valMap {
		var err error
		res[k], err = CtyToGo(v)

		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse " + attr.Name + "[" + k + "] attribute as map",
				Detail:   err.Error(),
				Subject:  &attr.Range,
			})
		}

	}

	return res, diags
}
