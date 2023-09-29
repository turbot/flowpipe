package hclhelpers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/go-kit/helpers"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
)

func isNumeric(i interface{}) bool {
	switch i.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128:
		return true
	default:
		return false
	}
}

func isSliceOfNumeric(slice interface{}) bool {
	value := reflect.ValueOf(slice)

	if value.Kind() != reflect.Slice {
		return false
	}

	if value.Len() == 0 {
		// An empty slice is not considered a slice of numeric values.
		return false
	}

	for i := 0; i < value.Len(); i++ {
		element := value.Index(i).Interface()
		if !isNumeric(element) {
			return false
		}
	}

	return true
}

func isSliceOfStrings(i interface{}) bool {
	// Check if it's a slice.
	if slice, ok := i.([]interface{}); ok {
		// Iterate over the elements of the slice.
		for _, v := range slice {
			// Check if the element is a string or an interface{} that contains a string.
			if _, isString := v.(string); !isString {
				// If any element is not a string, return false.
				return false
			}
		}
		// All elements are either strings or interface{} containing strings, return true.
		return true
	} else if _, ok := i.([]string); ok {
		// It's a []string, so return true.
		return true
	}
	// It's not a slice of strings or interfaces, return false.
	return false
}

func isStringMap(i interface{}) bool {
	// Check if the input is an interface
	if m, ok := i.(map[string]interface{}); ok {
		// Iterate over the map and check if all values are strings
		for _, v := range m {
			if _, isString := v.(string); !isString {
				return false
			}
		}
		return true
	} else if _, ok := i.(map[string]string); ok {
		// It's a []string, so return true.
		return true
	}
	return false
}

func isNumericMap(i interface{}) bool {
	// Check if the input is actually a map
	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Map {
		return false
	}

	// Iterate over the map and check the type of each value
	for _, key := range val.MapKeys() {
		value := val.MapIndex(key)
		if value.Kind() != reflect.Int && value.Kind() != reflect.Int8 && value.Kind() != reflect.Int16 && value.Kind() != reflect.Int32 && value.Kind() != reflect.Int64 &&
			value.Kind() != reflect.Uint && value.Kind() != reflect.Uint8 && value.Kind() != reflect.Uint16 && value.Kind() != reflect.Uint32 && value.Kind() != reflect.Uint64 &&
			value.Kind() != reflect.Float32 && value.Kind() != reflect.Float64 &&
			value.Kind() != reflect.Complex64 && value.Kind() != reflect.Complex128 {
			return false
		}
	}

	return true
}

func GoTypeMatchesCtyType(val interface{}, ctyType cty.Type) bool {
	if helpers.IsNil(val) {
		return false
	}

	if ctyType == cty.String {
		return reflect.TypeOf(val).Kind() == reflect.String
	}

	if ctyType == cty.Number {
		return isNumeric(val)
	}

	if ctyType == cty.Bool {
		return reflect.TypeOf(val).Kind() == reflect.Bool
	}

	if ctyType == cty.List(cty.String) {
		return isSliceOfStrings(val)
	}

	if ctyType == cty.List(cty.Number) {
		return isSliceOfNumeric(val)
	}

	if ctyType.IsListType() || ctyType.IsTupleType() {
		_, ok := val.([]interface{})
		return ok
	}

	if ctyType == cty.Map(cty.String) {
		return isStringMap(val)
	}

	if ctyType == cty.Map(cty.Number) {
		return isNumericMap(val)
	}

	if ctyType.IsMapType() || ctyType.IsObjectType() {
		return reflect.ValueOf(val).Kind() == reflect.Map
	}

	return false
}

func CoerceStringToGoBasedOnCtyType(input string, typ cty.Type) (interface{}, error) {
	// Check if the provided type is one of the supported types
	if typ == cty.String {
		return input, nil
	}

	if typ == cty.Number {
		// Attempt to convert the input string to an integer
		intValue, err := strconv.Atoi(input)
		if err != nil {
			floatValue, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return nil, perr.BadRequestWithMessage(fmt.Sprintf("unable to convert '%s' to a number", input))
			}
			return floatValue, nil
		}
		return intValue, err
	}

	if typ == cty.Bool {
		// Attempt to convert the input string to a boolean
		boolValue, err := strconv.ParseBool(input)
		if err != nil {
			return nil, perr.BadRequestWithMessage(fmt.Sprintf("unable to convert '%s' to a boolean", input))
		}
		return boolValue, nil
	}

	fakeFilename := fmt.Sprintf("<value for var.%s>", input)
	expr, diags := hclsyntax.ParseExpression([]byte(input), fakeFilename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("flowpipe", diags)
	}

	val, valDiags := expr.Value(nil)
	diags = append(diags, valDiags...)

	if typ.IsListType() || typ.IsTupleType() {

		if typ == cty.List(cty.String) {
			res, err := CtyToGoStringSlice(val)
			if err != nil {
				return nil, err
			}
			return res, error_helpers.HclDiagsToError("flowpipe", diags)
		}

		if typ == cty.List(cty.Number) {
			res, err := CtyToGoNumericSlice(val)
			if err != nil {
				return nil, err
			}
			return res, error_helpers.HclDiagsToError("flowpipe", diags)
		}

		if typ == cty.List(cty.Bool) {
			res, err := CtyToGoBoolSlice(val)
			if err != nil {
				return nil, err
			}
			return res, error_helpers.HclDiagsToError("flowpipe", diags)
		}

		res, err := CtyToGoInterfaceSlice(val)
		if err != nil {
			return nil, err
		}
		return res, error_helpers.HclDiagsToError("flowpipe", diags)
	}

	if typ.IsMapType() || typ.IsObjectType() {
		if typ == cty.Map(cty.String) {
			res, err := CtyToGoMapString(val)
			if err != nil {
				return nil, err
			}
			return res, error_helpers.HclDiagsToError("flowpipe", diags)
		}

		if typ == cty.Map(cty.Number) {
			res, err := CtyToGoMapNumeric(val)
			if err != nil {
				return nil, err
			}
			return res, error_helpers.HclDiagsToError("flowpipe", diags)
		}

		if typ == cty.Map(cty.Bool) {
			res, err := CtyToGoMapBool(val)
			if err != nil {
				return nil, err
			}
			return res, error_helpers.HclDiagsToError("flowpipe", diags)
		}

		res, err := CtyToGoMapInterface(val)
		if err != nil {
			return nil, err
		}
		return res, error_helpers.HclDiagsToError("flowpipe", diags)
	}

	return nil, perr.BadRequestWithMessage(fmt.Sprintf("unsupported type %s", typ.FriendlyName()))
}

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
					res = append(res, targetf)
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
			return nil, perr.BadRequestWithMessage("expected string type, but got " + v.Type().FriendlyName())
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

func CtyToGoNumericSlice(v cty.Value) (val []float64, err error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsListType() && !ty.IsTupleType() {
		return nil, perr.BadRequestWithMessage("expected list type")
	}

	var res []float64
	it := v.ElementIterator()
	for it.Next() {
		_, v := it.Element()

		// Return error if any of the value in the slice is not a number
		if v.Type() != cty.Number {
			return nil, perr.BadRequestWithMessage("expected number type, but got " + v.Type().FriendlyName())
		}

		var target float64
		err = gocty.FromCtyValue(v, &target)
		if err != nil {
			return nil, err
		}
		res = append(res, target)
	}

	return res, nil

}

func CtyToGoBoolSlice(v cty.Value) (val []bool, err error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsListType() && !ty.IsTupleType() {
		return nil, perr.BadRequestWithMessage("expected list type")
	}

	var res []bool
	it := v.ElementIterator()
	for it.Next() {
		_, v := it.Element()

		// Return error if any of the value in the slice is not a number
		if v.Type() != cty.Bool {
			return nil, perr.BadRequestWithMessage("expected number type, but got " + v.Type().FriendlyName())
		}

		res = append(res, v.True())
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

func CtyToGoMapString(v cty.Value) (map[string]string, error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsMapType() && !ty.IsObjectType() {
		return nil, perr.BadRequestWithMessage("expected map type")
	}

	res := map[string]string{}

	valueMap := v.AsValueMap()

	for k, v := range valueMap {
		if v.Type() != cty.String {
			return nil, perr.BadRequestWithMessage("expected string type, but got " + v.Type().FriendlyName())
		}

		target, err := CtyToString(v)
		if err != nil {
			return nil, err
		}
		res[k] = target
	}

	return res, nil
}

func CtyToGoMapNumeric(v cty.Value) (map[string]float64, error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsMapType() && !ty.IsObjectType() {
		return nil, perr.BadRequestWithMessage("expected map type")
	}

	res := map[string]float64{}

	valueMap := v.AsValueMap()

	for k, v := range valueMap {
		if v.Type() != cty.Number {
			return nil, perr.BadRequestWithMessage("expected number type, but got " + v.Type().FriendlyName())
		}
		var target float64
		err := gocty.FromCtyValue(v, &target)
		if err != nil {
			return nil, err
		}
		res[k] = target
	}

	return res, nil
}

func CtyToGoMapBool(v cty.Value) (map[string]bool, error) {
	if v.IsNull() || !v.IsWhollyKnown() {
		return nil, nil
	}
	ty := v.Type()
	if !ty.IsMapType() && !ty.IsObjectType() {
		return nil, fmt.Errorf("expected list type")
	}

	res := map[string]bool{}

	valueMap := v.AsValueMap()

	for k, v := range valueMap {
		res[k] = v.True()
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
