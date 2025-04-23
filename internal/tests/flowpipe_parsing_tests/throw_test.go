package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/pipe-fittings/schema"
)

func TestThrow(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/throw.fp")
	assert.Nil(err, "error found")

	pipeline := pipelines["local.pipeline.throw_simple_no_unresolved"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	throwConfigs := pipeline.Steps[0].GetThrowConfig()

	assert.Equal(1, len(throwConfigs))
	assert.True(len(throwConfigs[0].UnresolvedAttributes) > 0)
	assert.NotNil(throwConfigs[0].UnresolvedAttributes[schema.AttributeTypeIf])
	assert.Equal("foo", *throwConfigs[0].Message)

	// update 2023-03-08 - to make things easy we always set If as unresolved attribute, it will make
	// the code path easier to manage
	assert.Nil(throwConfigs[0].If)

	pipeline = pipelines["local.pipeline.throw_simple_unresolved"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	throwConfigs = pipeline.Steps[0].GetThrowConfig()

	assert.Equal(1, len(throwConfigs))
	assert.True(len(throwConfigs[0].UnresolvedAttributes) > 0)
	assert.NotNil(throwConfigs[0].UnresolvedAttributes[schema.AttributeTypeIf])
	assert.Equal("foo", *throwConfigs[0].Message)

	pipeline = pipelines["local.pipeline.throw_multiple"]
	if pipeline == nil {
		assert.Fail("pipeline not found")
		return
	}

	// step 0 -> transform.base
	// step 1 -> transform.base_2
	// step 2 -> transform.throw
	assert.Equal(2, len(pipeline.Steps[2].GetDependsOn()))
	assert.Equal("transform.base", pipeline.Steps[2].GetDependsOn()[0])
	assert.Equal("transform.base_2", pipeline.Steps[2].GetDependsOn()[1])

	throwConfigs = pipeline.Steps[2].GetThrowConfig()

	assert.Equal(4, len(throwConfigs))

	assert.True(len(throwConfigs[0].UnresolvedAttributes) == 2)
	assert.True(len(throwConfigs[1].UnresolvedAttributes) == 2)

	assert.True(len(throwConfigs[2].UnresolvedAttributes) == 1)
	assert.True(len(throwConfigs[3].UnresolvedAttributes) == 1)
}
