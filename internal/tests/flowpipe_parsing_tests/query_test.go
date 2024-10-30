package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/pipe-fittings/schema"
)

func TestQueryStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/query.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.query"] == nil {
		assert.Fail("query pipeline not found")
		return
	}

	step := pipelines["local.pipeline.query"].GetStep("query.query_1")
	if step == nil {
		assert.Fail("query step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("select * from foo", inputs[schema.AttributeTypeSql])
	assert.Equal("this is a connection string", inputs[schema.AttributeTypeDatabase])
	assert.Equal(60000, inputs[schema.AttributeTypeTimeout])
}

func TestQueryStepWithArgs(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/query.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.query_with_args"] == nil {
		assert.Fail("query pipeline not found")
		return
	}

	step := pipelines["local.pipeline.query_with_args"].GetStep("query.query_1")
	if step == nil {
		assert.Fail("query step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("select * from foo where bar = $1 and baz = $2", inputs[schema.AttributeTypeSql])
	assert.Equal(60000, inputs[schema.AttributeTypeTimeout])

	assert.Equal("this is a connection string", inputs[schema.AttributeTypeDatabase])

	args, ok := inputs[schema.AttributeTypeArgs].([]interface{})
	if !ok {
		assert.Fail("args not found")
		return
	}
	assert.Equal(2, len(args))
	assert.Equal("two", args[0])
	assert.Equal(10, args[1])
}
