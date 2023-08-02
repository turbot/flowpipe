package funcs

import (
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// error_message:  Given a reference to a step, error_message will return a string containing the first error message, if any. If there were no errors,
// it will return an empty string. This is useful for simple step primitives.
var ErrorMessageFunc = function.New(&function.Spec{
	Description: ` Given a reference to a step, error_message will return a string containing the first error message, if any. If there were no errors, it will return an empty string. This is useful for simple step primitives.`,
	Params: []function.Parameter{
		{
			Name:             "step",
			Type:             cty.DynamicPseudoType,
			AllowUnknown:     true,
			AllowDynamicType: true,
			AllowNull:        true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {

		if len(args) == 0 {
			return cty.StringVal(""), nil
		}

		val := args[0]

		valueMap := val.AsValueMap()
		if valueMap == nil {
			return cty.StringVal(""), nil
		}

		if valueMap["errors"].IsNull() {
			return cty.StringVal(""), nil
		}

		errors := valueMap["errors"].AsValueSlice()
		if len(errors) == 0 {
			return cty.StringVal(""), nil
		}

		firstError := errors[0]
		firstErrorMap := firstError.AsValueMap()

		return firstErrorMap[schema.AttributeTypeMessage], nil
	},
})
