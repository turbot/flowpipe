package pipeline

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/tests/test_init"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/zclconf/go-cty/cty"
)

type coerceValueTest struct {
	title         string
	resource      resources.ResourceWithParam
	input         map[string]string
	expected      map[string]interface{}
	errorExpected bool
}

type resourceWithParams struct {
	Params []resources.PipelineParam
}

func (r *resourceWithParams) GetParam(name string) *resources.PipelineParam {
	for _, p := range r.Params {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func (r *resourceWithParams) GetParams() []resources.PipelineParam {
	return r.Params
}

var coerceValueTests = []coerceValueTest{
	{
		title: "Coerce string value",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_string",
					Type: cty.String,
				},
			},
		},
		input: map[string]string{
			"param_string": "val_one",
		},
		expected: map[string]interface{}{
			"param_string": "val_one",
		},
	},
	{
		title: "Coerce int value",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_int",
					Type: cty.Number,
				},
			},
		},
		input: map[string]string{
			"param_int": "123",
		},
		expected: map[string]interface{}{
			"param_int": 123,
		},
	},
	{
		title: "Coerce bool value",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_bool",
					Type: cty.Bool,
				},
			},
		},
		input: map[string]string{
			"param_bool": "true",
		},
		expected: map[string]interface{}{
			"param_bool": true,
		},
	},
	{
		title: "Coerce bool value - invalid",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_bool",
					Type: cty.Bool,
				},
			},
		},
		input: map[string]string{
			"param_bool": "hello",
		},
		errorExpected: true,
	},
	{
		title: "Coerce connection",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{})),
				},
			},
		},
		input: map[string]string{
			"param_connection": "connection.aws.default",
		},
		expected: map[string]interface{}{
			"param_connection": map[string]interface{}{
				"short_name":    "default",
				"type":          "aws",
				"resource_type": "connection",
				"temporary":     true,
			},
		},
	},
	{
		title: "Coerce notifier",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_notifier",
					Type: cty.Capsule("notifier", reflect.TypeOf(&resources.NotifierImpl{})),
				},
			},
		},
		input: map[string]string{
			"param_notifier": "notifier.slack",
		},
		expected: map[string]interface{}{
			"param_notifier": map[string]interface{}{
				"name":          "slack",
				"resource_type": "notifier",
			},
		},
	},
	{
		title: "list of connections",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.List(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "[connection.aws.default,connection.aws.example]",
		},
		expected: map[string]interface{}{
			"param_connection": []interface{}{
				map[string]interface{}{
					"short_name":    "default",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
				map[string]interface{}{
					"short_name":    "example",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
			},
		},
	},
	{
		title: "list of connections - invalid name",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.List(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "[connection.aws.default,connection.aws.does_not_exist]",
		},
		errorExpected: true,
	},
	{
		title: "list of connections - invalid type",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.List(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "[connection.aws.default,connection.foo.default]",
		},
		errorExpected: true,
	},
	{
		title: "list of connection but just one",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.List(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "[connection.aws.default]",
		},
		expected: map[string]interface{}{
			"param_connection": []interface{}{
				map[string]interface{}{
					"short_name":    "default",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
			},
		},
	},
	{
		title: "map of connections",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.Map(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "{default=connection.aws.default,example=connection.aws.example}",
		},
		expected: map[string]interface{}{
			"param_connection": map[string]interface{}{
				"default": map[string]interface{}{
					"short_name":    "default",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
				"example": map[string]interface{}{
					"short_name":    "example",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
			},
		},
	},
	{
		title: "map of connections - invalid name",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.Map(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "{default=connection.aws.default,example=connection.aws.does_not_exist}",
		},
		errorExpected: true,
	},
	{
		title: "map of connections - invalid type",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.Map(cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{}))),
				},
			},
		},
		input: map[string]string{
			"param_connection": "{default=connection.aws.default,example=connection.foo.default}",
		},
		errorExpected: true,
	},
	{
		title: "object of different types",
		resource: &resourceWithParams{
			Params: []resources.PipelineParam{
				{
					Name: "param_connection",
					Type: cty.Object(map[string]cty.Type{
						"aws": cty.Capsule("aws", reflect.TypeOf(connection.AwsConnection{})),
						"foo": cty.String,
						"bar": cty.Bool,
					}),
				},
			},
		},
		input: map[string]string{
			"param_connection": `{aws=connection.aws.default,foo="hello",bar=true}`,
		},
		expected: map[string]interface{}{
			"param_connection": map[string]interface{}{
				"aws": map[string]interface{}{
					"short_name":    "default",
					"type":          "aws",
					"resource_type": "connection",
					"temporary":     true,
				},
				"foo": "hello",
				"bar": true,
			},
		},
	},
}

func TestCoerceCustomValue(tm *testing.T) {
	test_init.SetAppSpecificConstants()
	variables := map[string]cty.Value{
		"connection": cty.ObjectVal(map[string]cty.Value{
			"aws": cty.ObjectVal(map[string]cty.Value{
				"default": cty.ObjectVal(map[string]cty.Value{
					"short_name":    cty.StringVal("default"),
					"type":          cty.StringVal("aws"),
					"temporary":     cty.BoolVal(true),
					"resource_type": cty.StringVal("connection"),
				}),
				"example": cty.ObjectVal(map[string]cty.Value{
					"short_name":    cty.StringVal("example"),
					"type":          cty.StringVal("aws"),
					"temporary":     cty.BoolVal(true),
					"resource_type": cty.StringVal("connection"),
				}),
			}),
		}),
		"notifier": cty.ObjectVal(map[string]cty.Value{
			"slack": cty.ObjectVal(map[string]cty.Value{
				"name":          cty.StringVal("slack"),
				"resource_type": cty.StringVal("notifier"),
			}),
			"default": cty.ObjectVal(map[string]cty.Value{
				"name":          cty.StringVal("default"),
				"resource_type": cty.StringVal("notifier"),
			}),
		}),
	}
	evalCtx := &hcl.EvalContext{
		Variables: variables,
	}

	for _, tc := range coerceValueTests {
		tm.Run(tc.title, func(t *testing.T) {
			assert := assert.New(t)

			// pass nil evalCtx so it will not validate the actual connection/notifier against the
			// config
			result, err := parse.CoerceParams(tc.resource, tc.input, evalCtx)

			if tc.errorExpected && len(err) == 0 {
				assert.Fail("Expected error but got none")
				return
			}

			if tc.errorExpected && len(err) > 0 {
				return
			}

			if len(err) > 0 {
				assert.Fail("Error while coercing value", "error", err)
				return
			}

			assert.True(reflect.DeepEqual(tc.expected, result))
		})
	}
}
