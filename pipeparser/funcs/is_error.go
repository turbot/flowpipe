package funcs

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// is_error: Given a reference to a step, is_error returns a boolean true
// if there are 1 or more errors, or false it there are no errors.
var IsErrorFunc = function.New(&function.Spec{
	Description: `Given a reference to a step, is_error returns a boolean true if there are 1 or more errors, or false it there are no errors.`,
	Params: []function.Parameter{
		{
			Name:             "step",
			Type:             cty.DynamicPseudoType,
			AllowUnknown:     true,
			AllowDynamicType: true,
			AllowNull:        true,
		},
	},
	Type: function.StaticReturnType(cty.Bool),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		val := args[0]

		valueMap := val.AsValueMap()
		if valueMap == nil {
			return cty.True, nil
		}

		if valueMap["errors"].IsNull() {
			return cty.False, nil
		}

		return cty.True, nil
	},
})
