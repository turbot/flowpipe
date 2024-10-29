package pipeline_test

import (
	"context"
	parse2 "github.com/turbot/flowpipe/internal/parse"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingParamValidation(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse2.LoadPipelines(context.TODO(), "./pipelines/missing_param_validation.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.missing_param_validation_test"]
	if validateMyParam == nil {
		assert.Fail("missing_param_validation_test pipeline not found")
		return
	}

	stringValid := map[string]interface{}{
		"address_line_2": "Westminster",
	}

	assert.Equal(0, len(parse2.ValidateParams(validateMyParam, stringValid, nil)))

	stringInvalid := map[string]interface{}{
		"address_line_2": 123,
	}

	errs := parse2.ValidateParams(validateMyParam, stringInvalid, nil)
	assert.Equal(2, len(errs))
	assert.Equal("Bad Request: invalid data type for parameter 'address_line_2' wanted string but received int", errs[0].Error())
	assert.Equal("Bad Request: missing parameter: address_line_2", errs[1].Error())

	invalidParam := map[string]interface{}{
		"invalid": "foo",
	}
	errs = parse2.ValidateParams(validateMyParam, invalidParam, nil)
	assert.Equal(2, len(errs))
	assert.Equal("Bad Request: unknown parameter specified 'invalid'", errs[0].Error())
	assert.Equal("Bad Request: missing parameter: address_line_2", errs[1].Error())

	noParam := map[string]interface{}{}
	errs = parse2.ValidateParams(validateMyParam, noParam, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: missing parameter: address_line_2", errs[0].Error())
}
