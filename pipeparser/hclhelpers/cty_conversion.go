package hclhelpers

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

// CtyToJSON converts a cty value to it;s JSON representation
func CtyToJSON(val cty.Value) (string, error) {

	if !val.IsWhollyKnown() {
		return "", fmt.Errorf("cannot serialize unknown values")
	}

	if val.IsNull() {
		return "{}", nil
	}

	buf, err := json.Marshal(val, val.Type())
	if err != nil {
		return "", err
	}

	return string(buf), nil

}

// CtyToString convert a cty value into a string representation of the value
func CtyToString(v cty.Value) (valStr string, err error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return "", nil
	}
	ty := v.Type()
	switch {
	case ty.IsTupleType(), ty.IsListType():
		{
			var array []string
			if array, err = ctyTupleToArrayOfPgStrings(v); err == nil {
				valStr = fmt.Sprintf("[%s]", strings.Join(array, ","))
			}
			return
		}
	}

	switch ty {
	case cty.Bool:
		var target bool
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%v", target)
		}
	case cty.Number:
		var target int
		if err = gocty.FromCtyValue(v, &target); err == nil {
			valStr = fmt.Sprintf("%d", target)
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				valStr = fmt.Sprintf("%d", target)
			}
		}
	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			valStr = target
		}
	default:
		var json string
		// wrap as postgres string
		if json, err = CtyToJSON(v); err == nil {
			valStr = json
		}

	}

	return valStr, err
}

func CtyToInt64(val cty.Value) (*int64, hcl.Diagnostics) {
	if val.Type() != cty.Number {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse value as number",
		}}
	}

	bigFloatValue := val.AsBigFloat()

	if !bigFloatValue.IsInt() {
		return nil, hcl.Diagnostics{&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unable to parse value as int",
		}}
	}

	int64Value, _ := bigFloatValue.Int64()
	return &int64Value, hcl.Diagnostics{}
}

func CtyToGoInterfaceSlice(v cty.Value) (val []interface{}, err error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsListType() && !ty.IsTupleType() {
		return nil, fmt.Errorf("expected list type")
	}

	var res []interface{}
	it := v.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		switch v.Type() {
		case cty.Bool:
			var target bool
			err = gocty.FromCtyValue(v, &target)
			if err != nil {
				return nil, err
			}
			res = append(res, target)
		case cty.String:
			var target string
			err = gocty.FromCtyValue(v, &target)
			if err != nil {
				return nil, err
			}
			res = append(res, target)
		case cty.Number:
			var target int
			if err = gocty.FromCtyValue(v, &target); err == nil {
				res = append(res, target)
			} else {
				var targetf float64
				if err = gocty.FromCtyValue(v, &targetf); err == nil {
					res = append(res, target)
				} else {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf("unsupported type %s", v.Type().FriendlyName())
		}
	}
	return res, nil
}

func CtyToGoStringSlice(v cty.Value) (val []string, err error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsListType() && !ty.IsTupleType() {
		return nil, perr.BadRequestWithMessage("expected list type")
	}

	var res []string
	it := v.ElementIterator()
	for it.Next() {
		_, v := it.Element()

		// Return error if any of the value in the slice is not a string
		if v.Type() != cty.String {
			return nil, hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unable to parse value as string",
			}}
		}

		var target string
		err = gocty.FromCtyValue(v, &target)
		if err != nil {
			return nil, err
		}
		res = append(res, target)
	}

	return res, nil
}

func CtyToGoMapInterface(v cty.Value) (map[string]interface{}, error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsMapType() && !ty.IsObjectType() {
		return nil, fmt.Errorf("expected list type")
	}

	res := map[string]interface{}{}

	valueMap := v.AsValueMap()

	for k, v := range valueMap {
		target, err := CtyToGo(v)
		if err != nil {
			return nil, err
		}
		res[k] = target
	}

	return res, nil
}

func CtyToGo(v cty.Value) (val interface{}, err error) {
	if v.IsNull() {
		return nil, nil
	}

	ty := v.Type()
	switch {
	case ty.IsTupleType(), ty.IsListType():
		{
			target, err := ctyTupleToSliceOfInterfaces(v)
			if err != nil {
				return nil, err
			}
			val = target
		}
	case ty.IsMapType(), ty.IsObjectType():
		{
			target, err := CtyToGoMapInterface(v)
			if err != nil {
				return nil, err
			}
			val = target
		}
	}

	switch ty {
	case cty.Bool:
		var target bool
		if err = gocty.FromCtyValue(v, &target); err == nil {
			val = target
		}

	case cty.Number:
		var target int
		if err = gocty.FromCtyValue(v, &target); err == nil {
			val = target
		} else {
			var targetf float64
			if err = gocty.FromCtyValue(v, &targetf); err == nil {
				val = targetf
			}
		}
	case cty.String:
		var target string
		if err := gocty.FromCtyValue(v, &target); err == nil {
			val = target
		}

	default:
		var json string
		// wrap as postgres string
		if json, err = CtyToJSON(v); err == nil {
			val = json
		}
	}

	return
}

// CtyTypeToHclType converts a cty type to a hcl type
// accept multiple types and use the first non null and non dynamic one
func CtyTypeToHclType(types ...cty.Type) string {
	// find which if any of the types are non nil and not dynamic
	t := getKnownType(types)
	if t == cty.NilType {
		return ""
	}

	friendlyName := t.FriendlyName()

	// func to convert from ctyt aggregate syntax to hcl
	convertAggregate := func(prefix string) (string, bool) {
		if strings.HasPrefix(friendlyName, prefix) {
			return fmt.Sprintf("%s(%s)", strings.TrimSuffix(prefix, " of "), strings.TrimPrefix(friendlyName, prefix)), true
		}
		return "", false
	}

	if convertedName, isList := convertAggregate("list of "); isList {
		return convertedName
	}
	if convertedName, isMap := convertAggregate("map of "); isMap {
		return convertedName
	}
	if convertedName, isSet := convertAggregate("set of "); isSet {
		return convertedName
	}
	if friendlyName == "tuple" {
		elementTypes := t.TupleElementTypes()
		if len(elementTypes) == 0 {
			// we cannot determine the eleemnt type
			return "list"
		}
		// if there are element types, use the first one (assume homogeneous)
		underlyingType := elementTypes[0]
		return fmt.Sprintf("list(%s)", CtyTypeToHclType(underlyingType))
	}
	if friendlyName == "dynamic" {
		return ""
	}
	return friendlyName
}

// from a list oif cty typoes, return the first which is non nil and not dynamic
func getKnownType(types []cty.Type) cty.Type {
	for _, t := range types {
		if t != cty.NilType && !t.HasDynamicTypes() {
			return t
		}
	}
	return cty.NilType
}

func ctyTupleToArrayOfPgStrings(val cty.Value) ([]string, error) {
	var res []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		// decode the value into a postgres compatible
		valStr, err := CtyToPostgresString(v)
		if err != nil {
			return nil, err
		}

		res = append(res, valStr)
	}
	return res, nil
}

func ctyTupleToSliceOfInterfaces(val cty.Value) ([]interface{}, error) {
	var res []interface{}
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()

		target, err := CtyToGo(v)
		if err != nil {
			return nil, err
		}
		res = append(res, target)
	}
	return res, nil
}

func CtyTupleToArrayOfStrings(val cty.Value) ([]string, error) {
	var res []string
	it := val.ElementIterator()
	for it.Next() {
		_, v := it.Element()

		var valStr string
		if err := gocty.FromCtyValue(v, &valStr); err != nil {
			return nil, err
		}

		res = append(res, valStr)
	}
	return res, nil
}

func ConvertMapOrSliceToCtyValue(data interface{}) (cty.Value, error) {
	// Convert the input data to cty.Value based on its type
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		return convertSliceToCtyValue(data)
	case reflect.Map:
		return ConvertMapToCtyValue(data)
	default:
		// For other types, convert it as a single value using convertInterfaceToCtyValue
		return ConvertInterfaceToCtyValue(data)
	}
}

func ConvertInterfaceToCtyValue(v interface{}) (cty.Value, error) {
	// Use reflection to determine the underlying type and convert it to cty.Value
	//
	// file: go-cty/cty/gocty/time_implied.go/func impliedType
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		ctyType, err := gocty.ImpliedType(v)
		if err != nil {
			return cty.NilVal, err
		}

		val, err := gocty.ToCtyValue(v, ctyType)
		if err != nil {
			return cty.NilVal, err
		}
		return val, nil
	case reflect.Slice:
		return convertSliceToCtyValue(v)
	case reflect.Map:
		return ConvertMapToCtyValue(v)

	// Add more cases here for other types as needed.
	default:
		// If the type is not recognized, return a cty.NilVal as a placeholder
		return cty.NilVal, nil
	}
}

func convertSliceToCtyValue(v interface{}) (cty.Value, error) {
	// Convert the slice to a []interface{} and recursively convert it to cty values
	slice := v.([]interface{})
	ctyValues := make([]cty.Value, len(slice))
	for i, item := range slice {
		var err error
		ctyValues[i], err = ConvertInterfaceToCtyValue(item)
		if err != nil {
			return cty.NilVal, err
		}
	}

	// Create a cty.TupleVal from the cty values
	tupleVal := cty.TupleVal(ctyValues)

	// Return the cty.TupleVal as a cty.Value
	return tupleVal, nil
}

func ConvertMapToCtyValue(v interface{}) (cty.Value, error) {
	// Convert the map to a map[string]interface{} and recursively convert it to cty values
	mapData := v.(map[string]interface{})
	ctyValues := make(map[string]cty.Value, len(mapData))
	for key, value := range mapData {
		var err error
		ctyValues[key], err = ConvertInterfaceToCtyValue(value)
		if err != nil {
			return cty.NilVal, err
		}
	}

	// Create a cty.ObjectVal from the cty values
	objectVal := cty.ObjectVal(ctyValues)

	// Return the cty.ObjectVal as a cty.Value
	return objectVal, nil
}
