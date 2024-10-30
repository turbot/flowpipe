package pipeline_test

import (
	"context"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestContainerStep(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := parse.LoadPipelines(context.TODO(), "./pipelines/container.fp")
	assert.Nil(err, "error found")
	assert.Equal(4, len(pipelines), "wrong number of pipelines")

	if pipelines["local.pipeline.pipeline_step_container"] == nil {
		assert.Fail("pipeline_step_container pipeline not found")
		return
	}

	step := pipelines["local.pipeline.pipeline_step_container"].GetStep("container.container_test1")
	if step == nil {
		assert.Fail("container step not found")
		return
	}

	inputs, err := step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("container_test1", inputs[schema.AttributeTypeName])
	assert.Equal("test/image", inputs[schema.AttributeTypeImage])

	assert.Equal(60000, inputs[schema.AttributeTypeTimeout])
	assert.Equal(int64(128), inputs[schema.AttributeTypeMemory])
	assert.Equal(int64(64), inputs[schema.AttributeTypeMemoryReservation])
	assert.Equal(int64(-1), inputs[schema.AttributeTypeMemorySwap])
	assert.Equal(int64(60), inputs[schema.AttributeTypeMemorySwappiness])
	assert.Equal(int64(512), inputs[schema.AttributeTypeCpuShares])

	assert.Equal(false, inputs[schema.AttributeTypeReadOnly])
	assert.Equal("flowpipe", inputs[schema.AttributeTypeUser])
	assert.Equal(".", inputs[schema.AttributeTypeWorkdir])

	if _, ok := inputs[schema.AttributeTypeCmd].([]string); !ok {
		assert.Fail("attribute cmd should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeCmd].([]string)))

	if _, ok := inputs[schema.AttributeTypeEntrypoint].([]string); !ok {
		assert.Fail("attribute entry_point should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeEntrypoint].([]string)))

	if _, ok := inputs[schema.AttributeTypeEnv].(map[string]string); !ok {
		assert.Fail("env block is not defined correctly")
	}
	env := inputs[schema.AttributeTypeEnv].(map[string]string)
	assert.Equal("hello world", env["ENV_TEST"])

	// Pipeline 2

	if pipelines["local.pipeline.pipeline_step_with_param"] == nil {
		assert.Fail("pipeline_step_with_param pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_step_with_param"].GetStep("container.container_test1")
	if step == nil {
		assert.Fail("container step not found")
		return
	}

	paramVal := cty.ObjectVal(map[string]cty.Value{
		"region":     cty.StringVal("ap-south-1"),
		"image":      cty.StringVal("test/image"),
		"timeout":    cty.NumberIntVal(120000),
		"cpu_shares": cty.NumberIntVal(512),
		"cmd": cty.ListVal([]cty.Value{
			cty.StringVal("foo"),
			cty.StringVal("bar"),
		}),
		"entry_point": cty.ListVal([]cty.Value{
			cty.StringVal("foo"),
			cty.StringVal("bar"),
			cty.StringVal("baz"),
		}),
		"memory":             cty.NumberIntVal(128),
		"memory_reservation": cty.NumberIntVal(64),
		"memory_swap":        cty.NumberIntVal(-1),
		"memory_swappiness":  cty.NumberIntVal(60),
		"read_only":          cty.BoolVal(true),
		"container_user":     cty.StringVal("flowpipe"),
		"work_dir":           cty.StringVal("."),
	})

	evalContext := &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("container_test1", inputs[schema.AttributeTypeName])
	assert.Equal("test/image", inputs[schema.AttributeTypeImage])
	assert.Equal(120000, inputs[schema.AttributeTypeTimeout])
	assert.Equal(int64(512), inputs[schema.AttributeTypeCpuShares])
	assert.Equal(int64(128), inputs[schema.AttributeTypeMemory])
	assert.Equal(int64(64), inputs[schema.AttributeTypeMemoryReservation])
	assert.Equal(int64(-1), inputs[schema.AttributeTypeMemorySwap])
	assert.Equal(int64(60), inputs[schema.AttributeTypeMemorySwappiness])

	assert.Equal(true, inputs[schema.AttributeTypeReadOnly])
	assert.Equal("flowpipe", inputs[schema.AttributeTypeUser])
	assert.Equal(".", inputs[schema.AttributeTypeWorkdir])

	if _, ok := inputs[schema.AttributeTypeCmd].([]string); !ok {
		assert.Fail("attribute cmd should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeCmd].([]string)))

	if _, ok := inputs[schema.AttributeTypeEntrypoint].([]string); !ok {
		assert.Fail("attribute entrypoint should be a list of strings")
	}
	assert.Equal(3, len(inputs[schema.AttributeTypeEntrypoint].([]string)))

	if _, ok := inputs[schema.AttributeTypeEnv].(map[string]string); !ok {
		assert.Fail("env block is not defined correctly")
	}
	env = inputs[schema.AttributeTypeEnv].(map[string]string)
	assert.Equal("ap-south-1", env["REGION"])

	// Pipeline 3

	if pipelines["local.pipeline.pipeline_step_container_timeout_string"] == nil {
		assert.Fail("pipeline_step_container_timeout_string pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_step_container_timeout_string"].GetStep("container.container_test1")
	if step == nil {
		assert.Fail("container step not found")
		return
	}

	inputs, err = step.GetInputs(nil)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("container_test1", inputs[schema.AttributeTypeName])
	assert.Equal("test/image", inputs[schema.AttributeTypeImage])
	assert.Equal("60s", inputs[schema.AttributeTypeTimeout])
	assert.Equal(int64(128), inputs[schema.AttributeTypeMemory])
	assert.Equal(int64(64), inputs[schema.AttributeTypeMemoryReservation])
	assert.Equal(int64(-1), inputs[schema.AttributeTypeMemorySwap])
	assert.Equal(int64(60), inputs[schema.AttributeTypeMemorySwappiness])
	assert.Equal(int64(512), inputs[schema.AttributeTypeCpuShares])

	assert.Equal(false, inputs[schema.AttributeTypeReadOnly])
	assert.Equal("flowpipe", inputs[schema.AttributeTypeUser])
	assert.Equal(".", inputs[schema.AttributeTypeWorkdir])

	if _, ok := inputs[schema.AttributeTypeCmd].([]string); !ok {
		assert.Fail("attribute cmd should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeCmd].([]string)))

	if _, ok := inputs[schema.AttributeTypeEntrypoint].([]string); !ok {
		assert.Fail("attribute entry_point should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeEntrypoint].([]string)))

	if _, ok := inputs[schema.AttributeTypeEnv].(map[string]string); !ok {
		assert.Fail("env block is not defined correctly")
	}
	env = inputs[schema.AttributeTypeEnv].(map[string]string)
	assert.Equal("hello world", env["ENV_TEST"])

	// Pipeline 4

	if pipelines["local.pipeline.pipeline_step_container_with_param_timeout_string"] == nil {
		assert.Fail("pipeline_step_container_with_param_timeout_string pipeline not found")
		return
	}

	step = pipelines["local.pipeline.pipeline_step_container_with_param_timeout_string"].GetStep("container.container_test1")
	if step == nil {
		assert.Fail("container step not found")
		return
	}

	paramVal = cty.ObjectVal(map[string]cty.Value{
		"region":     cty.StringVal("ap-south-1"),
		"image":      cty.StringVal("test/image"),
		"timeout":    cty.StringVal("120s"),
		"cpu_shares": cty.NumberIntVal(512),
		"cmd": cty.ListVal([]cty.Value{
			cty.StringVal("foo"),
			cty.StringVal("bar"),
		}),
		"entry_point": cty.ListVal([]cty.Value{
			cty.StringVal("foo"),
			cty.StringVal("bar"),
			cty.StringVal("baz"),
		}),
		"memory":             cty.NumberIntVal(128),
		"memory_reservation": cty.NumberIntVal(64),
		"memory_swap":        cty.NumberIntVal(-1),
		"memory_swappiness":  cty.NumberIntVal(60),
		"read_only":          cty.BoolVal(true),
		"container_user":     cty.StringVal("flowpipe"),
		"work_dir":           cty.StringVal("."),
	})

	evalContext = &hcl.EvalContext{}
	evalContext.Variables = map[string]cty.Value{}
	evalContext.Variables["param"] = paramVal

	inputs, err = step.GetInputs(evalContext)
	if err != nil {
		assert.Fail("error getting inputs")
		return
	}
	assert.Equal("container_test1", inputs[schema.AttributeTypeName])
	assert.Equal("test/image", inputs[schema.AttributeTypeImage])
	assert.Equal("120s", inputs[schema.AttributeTypeTimeout])
	assert.Equal(int64(512), inputs[schema.AttributeTypeCpuShares])

	assert.Equal(int64(128), inputs[schema.AttributeTypeMemory])
	assert.Equal(int64(64), inputs[schema.AttributeTypeMemoryReservation])
	assert.Equal(int64(-1), inputs[schema.AttributeTypeMemorySwap])
	assert.Equal(int64(60), inputs[schema.AttributeTypeMemorySwappiness])

	assert.Equal(true, inputs[schema.AttributeTypeReadOnly])
	assert.Equal("flowpipe", inputs[schema.AttributeTypeUser])
	assert.Equal(".", inputs[schema.AttributeTypeWorkdir])

	if _, ok := inputs[schema.AttributeTypeCmd].([]string); !ok {
		assert.Fail("attribute cmd should be a list of strings")
	}
	assert.Equal(2, len(inputs[schema.AttributeTypeCmd].([]string)))

	if _, ok := inputs[schema.AttributeTypeEntrypoint].([]string); !ok {
		assert.Fail("attribute entrypoint should be a list of strings")
	}
	assert.Equal(3, len(inputs[schema.AttributeTypeEntrypoint].([]string)))

	if _, ok := inputs[schema.AttributeTypeEnv].(map[string]string); !ok {
		assert.Fail("env block is not defined correctly")
	}
	env = inputs[schema.AttributeTypeEnv].(map[string]string)
	assert.Equal("ap-south-1", env["REGION"])
}
