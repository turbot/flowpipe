package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestStepOutput(t *testing.T) {
	assert := assert.New(t)

	pipelines, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/test.fp")
	assert.Nil(err, "error found")

	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

	if pipelines["local.pipeline.step_output"] == nil {
		assert.Fail("step_output pipeline not found")
		return
	}

	assert.Equal(2, len(pipelines["local.pipeline.step_output"].Steps), "wrong number of steps")

	startStep := pipelines["local.pipeline.step_output"].GetStep("echo.start_step")
	assert.NotNil(startStep.GetOutputConfig()["start_output"])

	// objectVal := cty.ObjectVal(map[string]cty.Value{
	// 	"echo": cty.ObjectVal(map[string]cty.Value{
	// 		"start_step": cty.ObjectVal(map[string]cty.Value{
	// 			"text": cty.StringVal("foo"),
	// 			"output": cty.ObjectVal(map[string]cty.Value{
	// 				"start_output": cty.ObjectVal(map[string]cty.Value{
	// 					"value": cty.StringVal("bar"),
	// 				}),
	// 			}),
	// 		}),
	// 		"end_step": cty.ObjectVal(map[string]cty.Value{
	// 			"text": cty.StringVal("bar"),
	// 		}),
	// 	}),
	// })

	// step := pipelines["local.pipeline.step_output"].GetStep("echo.end_step")

	// // panic(fmt.Sprintf("%+v", step.GetOutputConfig()["start_output"]))

	// evalContext := &hcl.EvalContext{}
	// evalContext.Variables = map[string]cty.Value{}
	// evalContext.Variables["step"] = objectVal

	// inputs, err := step.GetInputs(evalContext)
	// if err != nil {
	// 	assert.Fail("error getting inputs")
	// 	return
	// }

	// panic(fmt.Sprintf("%+v", inputs))

	// // assert.Nil(pipelines["step_output"].Steps[0].GetOutputConfig()["start_output"])
}
