package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestMissingParamValidation(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/missing_param_validation.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.missing_param_validation_test"]
	if validateMyParam == nil {
		assert.Fail("missing_param_validation_test pipeline not found")
		return
	}

	stringValid := map[string]interface{}{
		"address_line_2": "Westminster",
	}

	assert.Equal(0, len(validateMyParam.ValidatePipelineParam(stringValid)))

	stringInvalid := map[string]interface{}{
		"address_line_2": 123,
	}

	errs := validateMyParam.ValidatePipelineParam(stringInvalid)
	assert.Equal(2, len(errs))
	assert.Equal("Bad Request: invalid type for parameter 'address_line_2'", errs[0].Error())
	assert.Equal("Bad Request: missing parameter: address_line_2", errs[1].Error())

	invalidParam := map[string]interface{}{
		"invalid": "foo",
	}
	errs = validateMyParam.ValidatePipelineParam(invalidParam)
	assert.Equal(2, len(errs))
	assert.Equal("Bad Request: unknown parameter specified 'invalid'", errs[0].Error())
	assert.Equal("Bad Request: missing parameter: address_line_2", errs[1].Error())

	noParam := map[string]interface{}{}
	errs = validateMyParam.ValidatePipelineParam(noParam)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: missing parameter: address_line_2", errs[0].Error())
}
