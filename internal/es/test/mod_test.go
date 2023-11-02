package es_test

// Basic imports
import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
)

type ModTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *ModTestSuite) SetupSuite() {

	err := os.Setenv("RUN_MODE", "TEST_ES")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("P_VAR_var_from_env", "from env")
	if err != nil {
		panic(err)
	}

	filepaths.PipesComponentWorkspaceDataDir = ".flowpipe"
	filepaths.PipesComponentModsFileName = "mod.hcl"
	filepaths.PipesComponentDefaultVarsFileName = "flowpipe.pvars"
	filepaths.PipesComponentDefaultInstallDir = "~/.flowpipe"

	constants.PipesComponentModDataExtension = ".hcl"
	constants.PipesComponentVariablesExtension = ".pvars"
	constants.PipesComponentAutoVariablesExtension = ".auto.pvars"
	constants.PipesComponentEnvInputVarPrefix = "P_VAR_"
	constants.PipesComponentAppName = "flowpipe"

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// clear the output dir before each test
	outputPath := path.Join(cwd, "output")

	// Check if the directory exists
	_, err = os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
		if err != nil {
			panic(err)
		}

	}

	pipelineDirPath := path.Join(cwd, "test_suite_mod")

	viper.GetViper().Set(constants.ArgModLocation, pipelineDirPath)
	viper.GetViper().Set(constants.ArgOutputDir, outputPath)
	viper.GetViper().Set(constants.ArgLogDir, outputPath)

	// Create a single, global context for the application
	ctx := context.Background()

	ctx = fplog.ContextWithLogger(ctx)
	ctx, err = config.ContextWithConfig(ctx)
	if err != nil {
		panic(err)
	}

	suite.ctx = ctx

	// We use the cache to store the pipelines
	cache.InMemoryInitialize(nil)

	m, err := manager.NewManager(ctx)
	suite.manager = m

	if err != nil {
		panic(err)
	}

	err = m.Initialize()
	if err != nil {
		panic(err)
	}

	// We don't do manager.Start() here because we don't want to start the API and Scheduler service

	esService, err := es.NewESService(ctx)
	if err != nil {
		panic(err)
	}
	err = esService.Start()
	if err != nil {
		panic(err)
	}
	esService.Status = "running"
	esService.StartedAt = utils.TimeNow()

	suite.esService = esService

	// Give some time for Watermill to fully start
	time.Sleep(2 * time.Second)

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *ModTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}
	suite.TearDownSuiteRunCount++
}

func (suite *ModTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *ModTestSuite) AfterTest(suiteName, testName string) {
	time.Sleep(2 * time.Second)
}

func (suite *ModTestSuite) TestSimplestPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal("Hello World", pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestSimpleForEachWithSleep() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.for_each_with_sleep", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal("ends", pex.PipelineOutput["val"].(map[string]interface{})["text"])
	assert.Equal("1s", pex.PipelineOutput["val_sleep"].(map[string]interface{})["0"].(map[string]interface{})["duration"])
	assert.Equal("2s", pex.PipelineOutput["val_sleep"].(map[string]interface{})["1"].(map[string]interface{})["duration"])
	assert.Equal("3s", pex.PipelineOutput["val_sleep"].(map[string]interface{})["2"].(map[string]interface{})["duration"])
}

func (suite *ModTestSuite) TestSimpleTwoStepsPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_two_steps", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal("Hello World", pex.PipelineOutput["val"])
	assert.Equal("Hello World: Hello World", pex.PipelineOutput["val_two"])

}

func (suite *ModTestSuite) XSkipTestSimpleLoop() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_loop", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal(3, len(pex.StepStatus["echo.repeat"]["0"].StepExecutions), "there should be 3 step executions executed by the loop")
	assert.Equal("Hello World", pex.PipelineOutput["val"])
	assert.Equal("Hello World: Hello World", pex.PipelineOutput["val_two"])

}

func (suite *ModTestSuite) TestCallingPipelineInDependentMod() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.echo_one", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("Hello World from Depend A", pex.PipelineOutput["echo_one_output"])

	// value should be: ${step.echo.var_one.text} + ${var.var_depend_a_one}
	assert.Equal("Hello World from Depend A: this is the value of var_one + this is the value of var_one", pex.PipelineOutput["echo_one_output_val_var_one"])
}

func (suite *ModTestSuite) TestModVars() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.echo_with_variable", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("Hello World: this is the value of var_one", pex.PipelineOutput["echo_one_output"])
	assert.Equal("Hello World Two: I come from flowpipe.vars file", pex.PipelineOutput["echo_two_output"])
	assert.Equal("Hello World Two: I come from flowpipe.vars file and Hello World Two: I come from flowpipe.vars file", pex.PipelineOutput["echo_three_output"])
	assert.Equal("value of locals_one", pex.PipelineOutput["echo_four_output"])
	assert.Equal("10 AND Hello World Two: I come from flowpipe.vars file AND value of locals_one", pex.PipelineOutput["echo_five_output"])
}

func (suite *ModTestSuite) TestSimpleNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.nested_simple_top", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("hello from the middle world", pex.PipelineOutput["val"])
	assert.Equal("two: hello from the middle world", pex.PipelineOutput["val_two"])
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithOutputClash() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.nested_simple_with_clash_merged_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "failed" {
		assert.Fail("Pipeline execution should fail")
		return
	}
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Detail, "output block 'val' already exists in step 'middle'")
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithMergedOutput() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.nested_simple_top_with_merged_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("hello from the middle world", pex.PipelineOutput["val"])
	assert.Equal("two: hello from the middle world", pex.PipelineOutput["val_two"])
	assert.Equal("step output", pex.PipelineOutput["val_step_output"])
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithForEach() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.nested_simple_top_with_for_each", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("hot mulligan", pex.PipelineOutput["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val_param"])
	assert.Equal("sugarcult", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val_param"])
	assert.Equal("the wonder years", pex.PipelineOutput["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val_param"])
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithForEachAndMergedOutput() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.nested_simple_top_with_for_each_with_merged_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("band: hot mulligan", pex.PipelineOutput["step_output_1"])
	assert.Equal("band: sugarcult", pex.PipelineOutput["step_output_2"])
	assert.Equal("band: the wonder years", pex.PipelineOutput["step_output_3"])

	assert.Equal("hot mulligan", pex.PipelineOutput["val_param_1"])
	assert.Equal("sugarcult", pex.PipelineOutput["val_param_2"])
	assert.Equal("the wonder years", pex.PipelineOutput["val_param_3"])

}

func (suite *ModTestSuite) TestPipelineWithStepOutput() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.with_step_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	echoNameStepOutputs := pex.AllNativeStepOutputs["echo"]["name"].(map[string]*modconfig.Output)
	assert.Equal(3, len(echoNameStepOutputs))
	assert.Equal("artist name: Real Friends", echoNameStepOutputs["0"].Data["text"])
	assert.Equal("artist name: A Day To Remember", echoNameStepOutputs["1"].Data["text"])
	assert.Equal("artist name: The Story So Far", echoNameStepOutputs["2"].Data["text"])

	secondStepStepOutputs := pex.AllNativeStepOutputs["echo"]["second_step"].(map[string]*modconfig.Output)
	assert.Equal(3, len(secondStepStepOutputs))
	assert.Equal("second_step: album name: Maybe This Place Is The Same And We're Just Changing", secondStepStepOutputs["0"].Data["text"])
	assert.Equal("second_step: album name: Common Courtesy", secondStepStepOutputs["1"].Data["text"])
	assert.Equal("second_step: album name: What You Don't See", secondStepStepOutputs["2"].Data["text"])

}

func (suite *ModTestSuite) TestPipelineWithForEach() {
	assert := assert.New(suite.T())
	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.run_me_controller", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("Hello: spock", pex.PipelineOutput["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("Hello: kirk", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("Hello: sulu", pex.PipelineOutput["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"])
}

func (suite *ModTestSuite) TestPipelineForEachTrippleNested() {
	assert := assert.New(suite.T())
	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.run_me_top", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	/* Expected output:


		"val": {
	        "0": {
	            "output": {},
	            "val": {
	                "0": {
	                    "output": {},
	                    "val": "bottom: aaa - spock"
	                },
	                "1": {
	                    "output": {},
	                    "val": "bottom: bbb - spock"
	                },
	                "2": {
	                    "output": {},
	                    "val": "bottom: ccc - spock"
	                }
	            }
	        },
	        "1": {
	            "output": {},
	            "val": {
	                "0": {
	                    "output": {},
	                    "val": "bottom: aaa - kirk"
	                },
	                "1": {
	                    "output": {},
	                    "val": "bottom: bbb - kirk"
	                },
	                "2": {
	                    "output": {},
	                    "val": "bottom: ccc - kirk"
	                }
	            }
	        },
	        "2": {
	            "output": {},
	            "val": {
	                "0": {
	                    "output": {},
	                    "val": "bottom: aaa - sulu"
	                },
	                "1": {
	                    "output": {},
	                    "val": "bottom: bbb - sulu"
	                },
	                "2": {
	                    "output": {},
	                    "val": "bottom: ccc - sulu"
	                }
	            }
	        }
	    }
		**/

	assert.Equal("bottom: aaa - spock", pex.PipelineOutput["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: bbb - spock", pex.PipelineOutput["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: ccc - spock", pex.PipelineOutput["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: aaa - kirk", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: bbb - kirk", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: ccc - kirk", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: aaa - sulu", pex.PipelineOutput["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: bbb - sulu", pex.PipelineOutput["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("bottom: ccc - sulu", pex.PipelineOutput["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"])
}

func (suite *ModTestSuite) TestPipelineWithArgs() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.calling_pipeline_with_params", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// We pass "bar" as name here
	assert.Equal("echo bar", pex.PipelineOutput["val"])
	assert.Equal("echo baz foo bar", pex.PipelineOutput["val_expr"])
	assert.Equal("echo this is the value of var_one", pex.PipelineOutput["val_from_val"])
}

func (suite *ModTestSuite) TestJsonArray() {
	// test for a bug where Flowpipe was assuming that JSON must be of map[string]interface{}
	assert := assert.New(suite.T())

	arrayInput := []string{"a", "b", "c"}

	pipelineInput := &modconfig.Input{
		"request_body": arrayInput,
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.json_array", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("foo", pex.PipelineOutput["val"].([]interface{})[0])
	assert.Equal("bar", pex.PipelineOutput["val"].([]interface{})[1])
	assert.Equal("baz", pex.PipelineOutput["val"].([]interface{})[2])

	// The output is re-formatted this way by jsonplaceholder.typicode.com, the array is turned into a map with the index as the key (as a string)
	assert.Equal("a", pex.PipelineOutput["val_two"].(map[string]interface{})["0"])
	assert.Equal("b", pex.PipelineOutput["val_two"].(map[string]interface{})["1"])
	assert.Equal("c", pex.PipelineOutput["val_two"].(map[string]interface{})["2"])
}

func (suite *ModTestSuite) TestPipelineWithForLoop() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.for_map", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	expectedStrings := []string{"janis joplin was 27", "jerry garcia was 53", "jimi hendrix was 27"}
	foundStrings := []string{
		pex.PipelineOutput["text_1"].(string),
		pex.PipelineOutput["text_2"].(string),
		pex.PipelineOutput["text_3"].(string),
	}
	less := func(a, b string) bool { return a < b }
	equalIgnoreOrder := cmp.Diff(expectedStrings, foundStrings, cmpopts.SortSlices(less)) == ""
	if !equalIgnoreOrder {
		assert.Fail("test_suite_mod.pipeline.for_map output not correct")
		return
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.set_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}
	assert.Equal("[0] guitar", pex.PipelineOutput["val_1"])
	assert.Equal("[1] bass", pex.PipelineOutput["val_2"])
	assert.Equal("[2] drums", pex.PipelineOutput["val_3"])

	assert.Equal(3, len(pex.PipelineOutput["val"].(map[string]interface{})))
	assert.Equal("[1] bass", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["text"])
}

func (suite *ModTestSuite) SkipTestDoUntil() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.do_until", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

}

func (suite *ModTestSuite) TestJsonAsOutput() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.json_output", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("uk", pex.PipelineOutput["country"])

	pex = nil
	pipelineCmd = nil

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.parent_json_output", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("London", pex.PipelineOutput["city"])
}

func (suite *ModTestSuite) TestMapReduce() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.reduce_map", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal(2, len(pex.PipelineOutput["val"].(map[string]interface{})))
	assert.Equal("green_day: Green Day", pex.PipelineOutput["val"].(map[string]interface{})["green_day"].(map[string]interface{})["text"])
}

func (suite *ModTestSuite) TestNested() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.top", 500*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Contains(pex.PipelineOutput["val_two"], "createIssue(input: {repositoryId: \\\"hendrix\\\", title: \\\"hello world\\\"}")

}

func (suite *ModTestSuite) TestForEach() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.for_each_empty_test", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Nil(pex.PipelineOutput["val"])

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.for_each_non_collection", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")

	if pex.Status != "failed" {
		assert.Fail("Pipeline should have failed")
		return
	}
}

func (suite *ModTestSuite) TestIntegrations() {
	assert := assert.New(suite.T())

	rootMod := suite.manager.RootMod
	assert.NotNil(rootMod)

	integrations := rootMod.ResourceMaps.Integrations["test_suite_mod.integration.slack.slack_app_from_var"]
	if integrations == nil {
		assert.Fail("test_suite_mod.integration.slack.slack_app_from_var not found")
		return
	}
}

func (suite *ModTestSuite) XXTestHttpPipelines() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.http_post_url_encoded", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}
}

// TODO : Add back the test to validatet he input step

// func (suite *ModTestSuite) TestPipelineInputStep() {
// 	assert := assert.New(suite.T())

// 	// Slack notify
// 	pipelineInput := &modconfig.Input{
// 		"channel": "#random",
// 	}

// 	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_slack_notify", 100*time.Millisecond, pipelineInput)
// 	if err != nil {
// 		assert.Fail("Error creating execution", err)
// 		return
// 	}

// 	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "started")
// 	if err != nil {
// 		assert.Fail("Error getting pipeline execution", err.Error())
// 		return
// 	}
// 	assert.Equal(1, len(pex.StepExecutionOrder["input.input"]))

// 	stepExecutionID := pex.StepExecutionOrder["input.input"][0]
// 	stepExecution := pex.StepExecutions[stepExecutionID]

// 	assert.NotNil(stepExecution.Input)

// 	stepInput := stepExecution.Input

// 	if _, ok := stepInput[schema.AttributeTypeNotifies].([]interface{}); !ok {
// 		assert.Fail("Input should have notifies")
// 		return
// 	}
// 	notifies := stepInput[schema.AttributeTypeNotifies].([]interface{})
// 	assert.Equal(1, len(notifies))

// 	if _, ok := notifies[0].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert notify to map[string]interface{}")
// 		return
// 	}
// 	notifyMap := notifies[0].(map[string]interface{})
// 	assert.Equal("#random", notifyMap[schema.AttributeTypeChannel].(string))

// 	if _, ok := notifyMap[schema.AttributeTypeIntegration].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert integration to map[string]interface{}")
// 		return
// 	}
// 	integrationMap := notifyMap[schema.AttributeTypeIntegration].(map[string]interface{})
// 	assert.Equal("slack", integrationMap[schema.AttributeTypeType].(string))
// 	assert.Equal("abcde", integrationMap[schema.AttributeTypeToken].(string))

// 	// Email notify
// 	pipelineInput = &modconfig.Input{}

// 	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_email_notify", 100*time.Millisecond, pipelineInput)
// 	if err != nil {
// 		assert.Fail("Error creating execution", err)
// 		return
// 	}

// 	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "started")
// 	if err != nil {
// 		assert.Fail("Error getting pipeline execution", err.Error())
// 		return
// 	}
// 	assert.Equal(1, len(pex.StepExecutionOrder["input.input"]))

// 	stepExecutionID = pex.StepExecutionOrder["input.input"][0]
// 	stepExecution = pex.StepExecutions[stepExecutionID]

// 	assert.NotNil(stepExecution.Input)

// 	stepInput = stepExecution.Input

// 	if _, ok := stepInput[schema.AttributeTypeNotifies].([]interface{}); !ok {
// 		assert.Fail("Input should have notifies")
// 		return
// 	}
// 	notifies = stepInput[schema.AttributeTypeNotifies].([]interface{})
// 	assert.Equal(1, len(notifies))

// 	if _, ok := notifies[0].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert notify to map[string]interface{}")
// 		return
// 	}
// 	notifyMap = notifies[0].(map[string]interface{})
// 	assert.Equal("awesomebob@blahblah.com", notifyMap[schema.AttributeTypeTo].(string))

// 	if _, ok := notifyMap[schema.AttributeTypeIntegration].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert integration to map[string]interface{}")
// 		return
// 	}
// 	integrationMap = notifyMap[schema.AttributeTypeIntegration].(map[string]interface{})
// 	assert.Equal("email", integrationMap[schema.AttributeTypeType].(string))
// 	assert.Equal("foo bar baz", integrationMap[schema.AttributeTypeSmtpHost].(string))
// 	assert.Equal("bar foo baz", integrationMap[schema.AttributeTypeDefaultSubject].(string))
// 	assert.Equal("baz bar foo", integrationMap[schema.AttributeTypeSmtpUsername].(string))

// 	// Notifies test
// 	pipelineInput = &modconfig.Input{
// 		"channel": "#random",
// 	}

// 	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_notifies", 100*time.Millisecond, pipelineInput)
// 	if err != nil {
// 		assert.Fail("Error creating execution", err)
// 		return
// 	}

// 	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "started")
// 	if err != nil {
// 		assert.Fail("Error getting pipeline execution", err.Error())
// 		return
// 	}
// 	assert.Equal(1, len(pex.StepExecutionOrder["input.input"]))

// 	stepExecutionID = pex.StepExecutionOrder["input.input"][0]
// 	stepExecution = pex.StepExecutions[stepExecutionID]

// 	assert.NotNil(stepExecution.Input)

// 	stepInput = stepExecution.Input

// 	if _, ok := stepInput[schema.AttributeTypeNotifies].([]interface{}); !ok {
// 		assert.Fail("Input should have notifies")
// 		return
// 	}
// 	notifies = stepInput[schema.AttributeTypeNotifies].([]interface{})
// 	assert.Equal(2, len(notifies))

// 	if _, ok := notifies[0].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert notify to map[string]interface{}")
// 		return
// 	}
// 	notifyMap = notifies[0].(map[string]interface{})
// 	assert.Equal("#random", notifyMap[schema.AttributeTypeChannel].(string))

// 	if _, ok := notifyMap[schema.AttributeTypeIntegration].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert integration to map[string]interface{}")
// 		return
// 	}
// 	integrationMap = notifyMap[schema.AttributeTypeIntegration].(map[string]interface{})
// 	assert.Equal("slack", integrationMap[schema.AttributeTypeType].(string))
// 	assert.Equal("abcde", integrationMap[schema.AttributeTypeToken].(string))

// 	if _, ok := notifies[1].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert notify to map[string]interface{}")
// 		return
// 	}
// 	notifyMap = notifies[1].(map[string]interface{})
// 	assert.Equal("awesomebob@blahblah.com", notifyMap[schema.AttributeTypeTo].(string))

// 	if _, ok := notifyMap[schema.AttributeTypeIntegration].(map[string]interface{}); !ok {
// 		assert.Fail("Unable to convert integration to map[string]interface{}")
// 		return
// 	}
// 	integrationMap = notifyMap[schema.AttributeTypeIntegration].(map[string]interface{})
// 	assert.Equal("foo bar baz", integrationMap[schema.AttributeTypeSmtpHost].(string))
// 	assert.Equal("bar foo baz", integrationMap[schema.AttributeTypeDefaultSubject].(string))
// 	assert.Equal("baz bar foo", integrationMap[schema.AttributeTypeSmtpUsername].(string))
// 	assert.Equal("email", integrationMap[schema.AttributeTypeType].(string))
// }

func TestModTestingSuite(t *testing.T) {
	suite.Run(t, &ModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
