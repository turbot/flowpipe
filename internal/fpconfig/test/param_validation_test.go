package pipeline_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestParamValidation(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/pipeline_param_validation.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.validate_my_param"]
	if validateMyParam == nil {
		assert.Fail("validate_my_param pipeline not found")
		return
	}

	stringValid := map[string]interface{}{
		"my_token": "abc",
	}

	assert.Equal(0, len(validateMyParam.ValidatePipelineParam(stringValid)))

	stringInvalid := map[string]interface{}{
		"my_token": 123,
	}

	errors := validateMyParam.ValidatePipelineParam(stringInvalid)
	assert.Equal(1, len(errors))
	assert.Equal("Bad Request: invalid type for parameter 'my_token'", errors[0].Error())

	invalidParam := map[string]interface{}{
		"invalid": "foo",
	}
	errors = validateMyParam.ValidatePipelineParam(invalidParam)
	assert.Equal(1, len(errors))
	assert.Equal("Bad Request: unknown parameter specified 'invalid'", errors[0].Error())

	allValid := map[string]interface{}{
		"my_token":      "123",
		"my_number":     123,
		"my_number_two": 123.45,
		"my_bool":       true,
	}

	errors = validateMyParam.ValidatePipelineParam(allValid)
	assert.Equal(0, len(errors))

	invalidNum := map[string]interface{}{
		"my_token":      "123",
		"my_number":     123,
		"my_number_two": "123.45",
		"my_bool":       true,
	}
	errors = validateMyParam.ValidatePipelineParam(invalidNum)
	assert.Equal(1, len(errors))
	assert.Equal("Bad Request: invalid type for parameter 'my_number_two'", errors[0].Error())

	moreThanOneInvalids := map[string]interface{}{
		"my_token":      "123",
		"my_number":     "a",
		"my_number_two": "123.45",
		"my_bool":       "true",
	}
	errors = validateMyParam.ValidatePipelineParam(moreThanOneInvalids)
	assert.Equal(3, len(errors))

	expectedErrors := []string{
		"Bad Request: invalid type for parameter 'my_number'",
		"Bad Request: invalid type for parameter 'my_bool'",
		"Bad Request: invalid type for parameter 'my_number_two'",
	}

	actualErrors := []string{}
	for _, err := range errors {
		actualErrors = append(actualErrors, err.Error())
	}

	less := func(a, b string) bool { return a < b }
	equalIgnoreOrder := cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	paramList := map[string]interface{}{
		"list_string":       []string{"foo", "bar"},
		"list_number":       []float64{1.23, 4.56},
		"list_number_two":   []float32{1.23, 4.56},
		"list_number_three": []int64{1, 4},
	}

	errors = validateMyParam.ValidatePipelineParam(paramList)
	assert.Equal(0, len(errors))

	paramListMoreNumberType := map[string]interface{}{
		"list_string":       []string{"foo", "bar"},
		"list_number":       []int{1, 4, 5, 6},
		"list_number_two":   []uint{1, 4, 5},
		"list_number_three": []int16{1, 4},
	}

	errors = validateMyParam.ValidatePipelineParam(paramListMoreNumberType)
	assert.Equal(0, len(errors))

	paramListAsInterface := map[string]interface{}{
		"list_string":       []interface{}{"foo", "bar"},
		"list_number":       []interface{}{1, 4, -4, 6},
		"list_number_two":   []interface{}{1, 4, 5.5}, // mixed float and int
		"list_number_three": []interface{}{1, 4},
	}

	errors = validateMyParam.ValidatePipelineParam(paramListAsInterface)
	assert.Equal(0, len(errors))

	paramNotList := map[string]interface{}{
		"list_string":     "foo",
		"list_number":     1,
		"list_number_two": 1.23,
	}

	errors = validateMyParam.ValidatePipelineParam(paramNotList)
	assert.Equal(3, len(errors))

	expectedErrors = []string{
		"Bad Request: invalid type for parameter 'list_string'",
		"Bad Request: invalid type for parameter 'list_number'",
		"Bad Request: invalid type for parameter 'list_number_two'",
	}

	actualErrors = []string{}
	for _, err := range errors {
		actualErrors = append(actualErrors, err.Error())
	}

	equalIgnoreOrder = cmp.Equal(expectedErrors, actualErrors, cmpopts.SortSlices(less))
	assert.True(equalIgnoreOrder, "expected errors do not match")

	listAny := map[string]interface{}{
		"list_any":       []interface{}{"foo", 1, 1.23, true},
		"list_any_two":   []interface{}{"foo", "bar", "baz"},
		"list_any_three": []interface{}{1, 2, 3},
	}

	errors = validateMyParam.ValidatePipelineParam(listAny)
	assert.Equal(0, len(errors))

	stringMap := map[string]interface{}{
		"map_of_string": map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}

	errors = validateMyParam.ValidatePipelineParam(stringMap)
	assert.Equal(0, len(errors))

	stringMapGeneric := map[string]interface{}{
		"map_of_string": map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
		},
	}
	errors = validateMyParam.ValidatePipelineParam(stringMapGeneric)
	assert.Equal(0, len(errors))

	stringMapGenericInvalid := map[string]interface{}{
		"map_of_string": map[string]interface{}{
			"foo": "bar",
			"baz": 123,
		},
	}
	errors = validateMyParam.ValidatePipelineParam(stringMapGenericInvalid)
	assert.Equal(1, len(errors))
	assert.Equal("Bad Request: invalid type for parameter 'map_of_string'", errors[0].Error())

	numberMap := map[string]interface{}{
		"map_of_number": map[string]float64{
			"foo": 1.23,
			"baz": 4.56,
		},
	}
	errors = validateMyParam.ValidatePipelineParam(numberMap)
	assert.Equal(0, len(errors))

	numberMapInvalid := map[string]interface{}{
		"map_of_number": map[string]interface{}{
			"foo": "1.23",
			"baz": "4.56",
		},
	}
	errors = validateMyParam.ValidatePipelineParam(numberMapInvalid)
	assert.Equal(1, len(errors))

	numberMapInvalid = map[string]interface{}{
		"map_of_number": map[string]string{
			"foo": "1.23",
			"baz": "4.56",
		},
		"map_of_number_two": 4,
	}
	errors = validateMyParam.ValidatePipelineParam(numberMapInvalid)
	assert.Equal(2, len(errors))

	numberMap = map[string]interface{}{
		"map_of_number": map[string]float64{
			"foo": 1.23,
			"baz": 4.56,
		},
		"map_of_number_two": map[string]int{
			"foo": 1,
			"baz": 4,
		},
	}
	errors = validateMyParam.ValidatePipelineParam(numberMap)
	assert.Equal(0, len(errors))

	numberMap = map[string]interface{}{
		"map_of_number": map[string]int16{
			"foo": 1,
			"baz": 4,
		},
		"map_of_number_two": map[string]uint32{
			"foo": 1,
			"baz": 4,
		},
	}
	errors = validateMyParam.ValidatePipelineParam(numberMap)
	assert.Equal(0, len(errors))

	anyMap := map[string]interface{}{
		"map_of_string": map[string]interface{}{
			"foo": "bar",
			"baz": "123",
		},
		"map_of_any": map[string]int16{
			"foo": 1,
			"baz": 4,
		},
		"map_of_any_two": map[string]string{
			"foo": "1",
			"baz": "4",
		},
		"map_of_any_three": map[string]interface{}{
			"foo": 1,
			"baz": "4",
		},
	}
	errors = validateMyParam.ValidatePipelineParam(anyMap)
	assert.Equal(0, len(errors))

	anyMapInvalid := map[string]interface{}{
		"map_of_any":       []interface{}{1, 2, 3},
		"map_of_any_two":   []interface{}{"foo", 2, 3},
		"map_of_any_three": 23,
	}
	errors = validateMyParam.ValidatePipelineParam(anyMapInvalid)
	assert.Equal(3, len(errors))

}
