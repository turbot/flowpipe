package types

import (
	"testing"

	"github.com/turbot/pipe-fittings/perr"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestStepAsHclVariables(t *testing.T) {
	assert := assert.New(t)

	stepOutput := modconfig.Output{
		Data: map[string]interface{}{},
	}

	stepOutput.Data["string"] = "one"
	stepOutput.Data["int"] = 25
	stepOutput.Data["bool"] = true

	stepOutput.Errors = []modconfig.StepError{
		{
			Error: perr.ErrorModel{Detail: "one"},
		},
		{
			Error:               perr.ErrorModel{Detail: "two"},
			PipelineExecutionID: "1234",
		},
	}

	hclVariables, err := stepOutput.AsCtyMap()
	if err != nil {
		assert.Fail("Error converting step output to HCL variables", err)
		return
	}

	assert.Equal("one", hclVariables["string"].AsString())
	assert.Equal(true, hclVariables["int"].AsBigFloat().IsInt())

	var intVal int
	err = gocty.FromCtyValue(hclVariables["int"], &intVal)
	if err != nil {
		assert.Fail("Unable to convert cty value to int")
		return
	}

	assert.Equal(25, intVal)
	assert.Equal(cty.True, hclVariables["bool"])

	errors := hclVariables["errors"]
	errorSlice := errors.AsValueSlice()
	assert.Equal(2, len(errorSlice), "there should be 2 errors")
	assert.Equal("one", errorSlice[0].AsValueMap()["error"].AsValueMap()["detail"].AsString())
	assert.Equal("two", errorSlice[1].AsValueMap()["error"].AsValueMap()["detail"].AsString())
	assert.Equal("1234", errorSlice[1].AsValueMap()["pipeline_execution_id"].AsString())
}
