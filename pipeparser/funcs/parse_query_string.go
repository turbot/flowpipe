package funcs

import (
	"net/url"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var ParseQueryString = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "query",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.Map(cty.String)),
	Impl: func(args []cty.Value, retType cty.Type) (ret cty.Value, err error) {
		query := args[0].AsString()

		parsedQuery, err := url.ParseQuery(query)
		if err != nil {
			return cty.UnknownVal(cty.Map(cty.String)), err
		}

		result := make(map[string]cty.Value)
		for key, values := range parsedQuery {
			// Taking the last value if multiple values are provided for the same key
			result[key] = cty.StringVal(values[len(values)-1])
		}

		return cty.MapVal(result), nil
	},
})
