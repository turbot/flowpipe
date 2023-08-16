package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestStepAsHclVariables(t *testing.T) {
	assert := assert.New(t)

	stepOutput := pipeline.Output{
		Data: map[string]interface{}{},
	}

	stepOutput.Data["string"] = "one"
	stepOutput.Data["int"] = 25
	stepOutput.Data["bool"] = true

	stepOutput.Errors = []pipeline.StepError{
		{
			Message: "one",
		},
		{
			Message:             "two",
			PipelineExecutionID: "1234",
		},
	}

	hclVariables, err := stepOutput.AsCtyValue()
	if err != nil {
		assert.Fail("Error converting step output to HCL variables", err)
		return
	}

	hclVariablesMap := hclVariables.AsValueMap()

	assert.Equal("one", hclVariablesMap["string"].AsString())
	assert.Equal(true, hclVariablesMap["int"].AsBigFloat().IsInt())

	var intVal int
	err = gocty.FromCtyValue(hclVariablesMap["int"], &intVal)
	if err != nil {
		assert.Fail("Unable to convert cty value to int")
		return
	}

	assert.Equal(25, intVal)
	assert.Equal(cty.True, hclVariablesMap["bool"])

	errors := hclVariablesMap["errors"]
	errorSlice := errors.AsValueSlice()
	assert.Equal(2, len(errorSlice), "there should be 2 errors")
	assert.Equal("one", errorSlice[0].AsValueMap()["message"].AsString())
	assert.Equal("two", errorSlice[1].AsValueMap()["message"].AsString())
	assert.Equal("1234", errorSlice[1].AsValueMap()["pipeline_execution_id"].AsString())
}
