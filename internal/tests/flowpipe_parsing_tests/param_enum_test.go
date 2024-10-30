package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	fparse "github.com/turbot/flowpipe/internal/parse"
)

func TestParamEnum(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := fparse.LoadPipelines(context.TODO(), "./pipelines/param_enum_test.fp")
	assert.Nil(err, "error found")

	validateMyParam := pipelines["local.pipeline.param_enum_test"]
	if validateMyParam == nil {
		assert.Fail("missing_param_validation_test pipeline not found")
		return
	}

	stringValid := map[string]interface{}{
		"city": "New York",
	}

	assert.Equal(0, len(fparse.ValidateParams(validateMyParam, stringValid, nil)))

	stringInvalid := map[string]interface{}{
		"city": "Sydney",
	}

	errs := fparse.ValidateParams(validateMyParam, stringInvalid, nil)
	assert.Equal(1, len(errs))
	assert.Equal("Bad Request: invalid value for param city", errs[0].Error())

	numValid := map[string]string{
		"number": "345",
	}

	res, errs := fparse.CoerceParams(validateMyParam, numValid, nil)
	if len(errs) > 0 {
		assert.Fail("Error found", errs)
		return
	}
	assert.Equal(345, res["number"])
}
