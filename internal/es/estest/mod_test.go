package estest

// Basic imports
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/turbot/pipe-fittings/sanitize"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/cache"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	fpconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/container"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

type ModTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	server                *http.Server
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
	err = os.Setenv("FP_VAR_var_from_env", "from env")
	if err != nil {
		panic(err)
	}

	suite.server = StartServer()

	// sets app specific constants defined in pipe-fittings
	viper.SetDefault("main.version", "0.0.0-test.0")
	viper.SetDefault(constants.ArgProcessRetention, 604800) // 7 days
	localcmdconfig.SetAppSpecificConstants()

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pipelineDirPath := path.Join(cwd, "test_suite_mod")

	viper.GetViper().Set(constants.ArgModLocation, pipelineDirPath)

	// delete flowpipe.db
	flowpipeDbFilename := filepaths.FlowpipeDBFileName()

	_, err = os.Stat(flowpipeDbFilename)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove(flowpipeDbFilename)
		if err != nil {
			panic(err)
		}
	}

	// Create a single, global context for the application
	ctx := context.Background()
	suite.ctx = ctx

	// We use the cache to store the pipelines
	cache.InMemoryInitialize(nil)

	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx, manager.WithESService()).Start()
	error_helpers.FailOnError(err)
	suite.esService = m.ESService

	suite.manager = m

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

	suite.server.Shutdown(suite.ctx) //nolint:errcheck // just a test case
	suite.TearDownSuiteRunCount++
}

func (suite *ModTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *ModTestSuite) AfterTest(suiteName, testName string) {
}

func (suite *ModTestSuite) TestSimplestPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestCallingPipelineInDependentMod() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// Run two pipeline at the same time
	_, pipelineCmd, _ := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.echo_one", 100*time.Millisecond, pipelineInput)
	_, pipelineCmd2, _ := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.echo_one", 100*time.Millisecond, pipelineInput)

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 20, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("Hello World from Depend A", pex.PipelineOutput["echo_one_output"])

	// value should be: ${step.echo.var_one.text} + ${var.var_depend_a_one}
	assert.Equal("Hello World from Depend A: this is the value of var_one + this is the value of var_one", pex.PipelineOutput["echo_one_output_val_var_one"])

	_, pex2, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd2.Event, pipelineCmd2.PipelineExecutionID, 100*time.Millisecond, 20, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("Hello World from Depend A", pex2.PipelineOutput["echo_one_output"])

	// value should be: ${step.echo.var_one.text} + ${var.var_depend_a_one}
	assert.Equal("Hello World from Depend A: this is the value of var_one + this is the value of var_one", pex2.PipelineOutput["echo_one_output_val_var_one"])
}

func (suite *ModTestSuite) TestSimpleForEachWithSleep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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
	assert.Equal("ends", pex.PipelineOutput["val"].(map[string]interface{})["value"])
	assert.Equal("1s", pex.PipelineOutput["val_sleep"].(map[string]interface{})["0"].(map[string]interface{})["duration"])
	assert.Equal("2s", pex.PipelineOutput["val_sleep"].(map[string]interface{})["1"].(map[string]interface{})["duration"])
	assert.Equal("3s", pex.PipelineOutput["val_sleep"].(map[string]interface{})["2"].(map[string]interface{})["duration"])
}

func (suite *ModTestSuite) TestSimpleTwoStepsPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestSimpleLoop() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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
	assert.Equal("iteration: 0", pex.PipelineOutput["val_1"])
	assert.Equal("override for 0", pex.PipelineOutput["val_2"])
	assert.Equal("override for 1", pex.PipelineOutput["val_3"])

	// Now check the integrity of the StepStatus

	assert.Equal(1, len(pex.StepStatus["transform.repeat"]), "there should only be 1 element because this isn't a for_each step")
	assert.Equal(3, len(pex.StepStatus["transform.repeat"]["0"].StepExecutions))
}

func (suite *ModTestSuite) TestLoopWithForEachAndNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_with_for_each_and_nested_pipeline", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.NotNil(pex)

	// We should have 3 for_each and each for_each has an exact 3 iterations.
	// check the output, it should not double up on the iteration count

	assert.Equal(3, len(pex.StepStatus["pipeline.repeat"]))
	assert.Equal(3, len(pex.StepStatus["pipeline.repeat"]["0"].StepExecutions))
	assert.Equal(3, len(pex.StepStatus["pipeline.repeat"]["1"].StepExecutions))
	assert.Equal(3, len(pex.StepStatus["pipeline.repeat"]["2"].StepExecutions))

	assert.Equal("0: oasis", pex.StepStatus["pipeline.repeat"]["0"].StepExecutions[0].Output.Data["output"].(map[string]interface{})["val"])
	assert.Equal("1: oasis", pex.StepStatus["pipeline.repeat"]["0"].StepExecutions[1].Output.Data["output"].(map[string]interface{})["val"])
	assert.Equal("2: oasis", pex.StepStatus["pipeline.repeat"]["0"].StepExecutions[2].Output.Data["output"].(map[string]interface{})["val"])

	assert.Equal("0: blur", pex.StepStatus["pipeline.repeat"]["1"].StepExecutions[0].Output.Data["output"].(map[string]interface{})["val"])
	assert.Equal("1: blur", pex.StepStatus["pipeline.repeat"]["1"].StepExecutions[1].Output.Data["output"].(map[string]interface{})["val"])
	assert.Equal("2: blur", pex.StepStatus["pipeline.repeat"]["1"].StepExecutions[2].Output.Data["output"].(map[string]interface{})["val"])

	assert.Equal("0: radiohead", pex.StepStatus["pipeline.repeat"]["2"].StepExecutions[0].Output.Data["output"].(map[string]interface{})["val"])
	assert.Equal("1: radiohead", pex.StepStatus["pipeline.repeat"]["2"].StepExecutions[1].Output.Data["output"].(map[string]interface{})["val"])
	assert.Equal("2: radiohead", pex.StepStatus["pipeline.repeat"]["2"].StepExecutions[2].Output.Data["output"].(map[string]interface{})["val"])
}

func (suite *ModTestSuite) TestSimpleForEach() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_for_each", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.NotNil(pex)

	// We should have 3 for_each and each for_each has exactly 1 execution.
	// check the output, it should not double up on the iteration count
	assert.Equal(3, len(pex.StepStatus["transform.echo"]))
	assert.Equal(1, len(pex.StepStatus["transform.echo"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["transform.echo"]["1"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["transform.echo"]["2"].StepExecutions))

	// Print out the step status if the step executions is not exactly 1
	if len(pex.StepStatus["transform.echo"]["0"].StepExecutions) != 1 ||
		len(pex.StepStatus["transform.echo"]["1"].StepExecutions) != 1 ||
		len(pex.StepStatus["transform.echo"]["2"].StepExecutions) != 1 {
		s, err := prettyjson.Marshal(pex.StepStatus["echo.repeat"])

		if err != nil {
			assert.Fail("Error marshalling pipeline output", err)
			return
		}

		fmt.Println(string(s)) //nolint:forbidigo // test
	}

	assert.Equal("0: foo bar", pex.StepStatus["transform.echo"]["0"].StepExecutions[0].Output.Data["value"])
	assert.Equal("1: foo baz", pex.StepStatus["transform.echo"]["1"].StepExecutions[0].Output.Data["value"])
	assert.Equal("2: foo qux", pex.StepStatus["transform.echo"]["2"].StepExecutions[0].Output.Data["value"])
}

func (suite *ModTestSuite) TestForEachOneAndForEachTwo() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.for_each_one_and_for_each_two", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.NotNil(pex)

	// Count how many step executions in each Step to ensure we don't have duplicate executions
	assert.Equal(1, len(pex.StepStatus["transform.first"]["0"].StepExecutions))
	assert.Equal(12, len(pex.StepStatus["transform.echo"]))
	assert.Equal(1, len(pex.StepStatus["transform.last"]["0"].StepExecutions))
}

func (suite *ModTestSuite) TestLoopWithForEach() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_with_for_each", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.NotNil(pex)

	// We should have 3 for_each and each for_each has exactly 4 iterations.
	// check the output, it should not double up on the iteration count
	assert.Equal(3, len(pex.StepStatus["transform.repeat"]))
	assert.Equal(4, len(pex.StepStatus["transform.repeat"]["0"].StepExecutions))
	assert.Equal(4, len(pex.StepStatus["transform.repeat"]["1"].StepExecutions))
	assert.Equal(4, len(pex.StepStatus["transform.repeat"]["2"].StepExecutions))

	// Print out the step status if the step executions is not exactly 4
	if len(pex.StepStatus["transform.repeat"]["0"].StepExecutions) != 4 ||
		len(pex.StepStatus["transform.repeat"]["1"].StepExecutions) != 4 ||
		len(pex.StepStatus["transform.repeat"]["2"].StepExecutions) != 4 {
		s, err := prettyjson.Marshal(pex.StepStatus["echo.repeat"])

		if err != nil {
			assert.Fail("Error marshalling pipeline output", err)
			return
		}

		fmt.Println(string(s)) //nolint:forbidigo // test
	}

	assert.Equal("iteration: 0 - oasis", pex.StepStatus["transform.repeat"]["0"].StepExecutions[0].Output.Data["value"])
	assert.Equal("iteration: 1 - oasis", pex.StepStatus["transform.repeat"]["0"].StepExecutions[1].Output.Data["value"])
	assert.Equal("iteration: 2 - oasis", pex.StepStatus["transform.repeat"]["0"].StepExecutions[2].Output.Data["value"])
	assert.Equal("iteration: 3 - oasis", pex.StepStatus["transform.repeat"]["0"].StepExecutions[3].Output.Data["value"])

	assert.Equal("iteration: 0 - blur", pex.StepStatus["transform.repeat"]["1"].StepExecutions[0].Output.Data["value"])
	assert.Equal("iteration: 1 - blur", pex.StepStatus["transform.repeat"]["1"].StepExecutions[1].Output.Data["value"])
	assert.Equal("iteration: 2 - blur", pex.StepStatus["transform.repeat"]["1"].StepExecutions[2].Output.Data["value"])
	assert.Equal("iteration: 3 - blur", pex.StepStatus["transform.repeat"]["1"].StepExecutions[3].Output.Data["value"])

	assert.Equal("iteration: 0 - radiohead", pex.StepStatus["transform.repeat"]["2"].StepExecutions[0].Output.Data["value"])
	assert.Equal("iteration: 1 - radiohead", pex.StepStatus["transform.repeat"]["2"].StepExecutions[1].Output.Data["value"])
	assert.Equal("iteration: 2 - radiohead", pex.StepStatus["transform.repeat"]["2"].StepExecutions[2].Output.Data["value"])
	assert.Equal("iteration: 3 - radiohead", pex.StepStatus["transform.repeat"]["2"].StepExecutions[3].Output.Data["value"])
}

func (suite *ModTestSuite) TestSimpleLoopWithIndex() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_loop_index", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("iteration: 0", pex.PipelineOutput["val_1"])
	assert.Equal("iteration: 1", pex.PipelineOutput["val_2"])
	assert.Equal("iteration: 2", pex.PipelineOutput["val_3"])

	// Now check the integrity of the StepStatus

	assert.Equal(1, len(pex.StepStatus["transform.repeat"]), "there should only be 1 element because this isn't a for_each step")
	assert.Equal(3, len(pex.StepStatus["transform.repeat"]["0"].StepExecutions))
	assert.Equal(false, pex.StepStatus["transform.repeat"]["0"].StepExecutions[1].StepLoop.LoopCompleted)
	assert.Equal(true, pex.StepStatus["transform.repeat"]["0"].StepExecutions[2].StepLoop.LoopCompleted)

	assert.Equal(1, pex.StepStatus["transform.repeat"]["0"].StepExecutions[0].StepLoop.Index, "step loop index at the execution is actually to be used for the next loop, it should be offset by one")
	assert.Equal(2, pex.StepStatus["transform.repeat"]["0"].StepExecutions[1].StepLoop.Index)
	assert.Equal(2, pex.StepStatus["transform.repeat"]["0"].StepExecutions[2].StepLoop.Index, "the last index should be the same with the second last becuse loop ends here, so it's not incremented")
}

func (suite *ModTestSuite) TestLoopWithForEachWithSleep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// We have to use the sleep step here to avoid concurrency issue with the planner
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_with_for_each_sleep", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	// Yeah this is a long test, the sleep is 4 seconds x 3
	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 14, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	// Now check the integrity of the StepStatus

	assert.Equal(2, len(pex.StepStatus["sleep.repeat"]), "there should only be 2 elements here because we have a for_each")
	assert.Equal(3, len(pex.StepStatus["sleep.repeat"]["0"].StepExecutions), "we should have 3 executions in each for_each")
	assert.Equal(3, len(pex.StepStatus["sleep.repeat"]["1"].StepExecutions), "we should have 3 executions in each for_each")

	assert.Equal(false, pex.StepStatus["sleep.repeat"]["0"].StepExecutions[1].StepLoop.LoopCompleted)
	assert.Equal(true, pex.StepStatus["sleep.repeat"]["0"].StepExecutions[2].StepLoop.LoopCompleted)

	assert.Equal(false, pex.StepStatus["sleep.repeat"]["1"].StepExecutions[1].StepLoop.LoopCompleted)
	assert.Equal(true, pex.StepStatus["sleep.repeat"]["1"].StepExecutions[2].StepLoop.LoopCompleted)

	assert.Equal(1, pex.StepStatus["sleep.repeat"]["0"].StepExecutions[0].StepLoop.Index, "step loop index at the execution is actually to be used for the next loop, it should be offset by one")
	assert.Equal(2, pex.StepStatus["sleep.repeat"]["0"].StepExecutions[1].StepLoop.Index)
	assert.Equal(2, pex.StepStatus["sleep.repeat"]["0"].StepExecutions[2].StepLoop.Index, "the last index should be the same with the second last becuse loop ends here, so it's not incremented")

	assert.Equal(1, pex.StepStatus["sleep.repeat"]["1"].StepExecutions[0].StepLoop.Index, "step loop index at the execution is actually to be used for the next loop, it should be offset by one")
	assert.Equal(2, pex.StepStatus["sleep.repeat"]["1"].StepExecutions[1].StepLoop.Index)
	assert.Equal(2, pex.StepStatus["sleep.repeat"]["1"].StepExecutions[2].StepLoop.Index, "the last index should be the same with the second last becuse loop ends here, so it's not incremented")
}

func (suite *ModTestSuite) TestSimpleNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestDynamicParamNested() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.top_dynamic", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("A", pex.PipelineOutput["val_a"].(map[string]interface{})["output"].(map[string]interface{})["val"])
	assert.Equal("B", pex.PipelineOutput["val_b"].(map[string]interface{})["output"].(map[string]interface{})["val"])

	// run it again with param this time
	pipelineInput = modconfig.Input{
		"pipe": "middle_dynamic_c",
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.top_dynamic", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("A", pex.PipelineOutput["val_a"].(map[string]interface{})["output"].(map[string]interface{})["val"])

	// Now it should be running middle_dynamic_c pipeline
	assert.Equal("C", pex.PipelineOutput["val_b"].(map[string]interface{})["output"].(map[string]interface{})["val"])
}

func (suite *ModTestSuite) TestDynamicParamNestedStepRef() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.top_dynamic_step_ref", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("C", pex.PipelineOutput["val"].(map[string]interface{})["output"].(map[string]interface{})["val"])
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithOutputClash() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.nested_simple_with_clash_merged_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution should fail")
		return
	}
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Detail, "output block 'val' already exists in step 'middle'")
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithMergedOutput() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

	assert.Equal("hello fr"+
		"om the middle world", pex.PipelineOutput["val"])
	assert.Equal("two: hello from the middle world", pex.PipelineOutput["val_two"])
	assert.Equal("step output", pex.PipelineOutput["val_step_output"])
}

func (suite *ModTestSuite) TestSimpleNestedPipelineWithForEach() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

	pipelineInput := modconfig.Input{}

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

	pipelineInput := modconfig.Input{}

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

	assert.Equal(3, len(pex.StepStatus["transform.name"]))
	assert.Equal("artist name: Real Friends", pex.StepStatus["transform.name"]["0"].StepExecutions[0].Output.Data["value"])
	assert.Equal("artist name: A Day To Remember", pex.StepStatus["transform.name"]["1"].StepExecutions[0].Output.Data["value"])
	assert.Equal("artist name: The Story So Far", pex.StepStatus["transform.name"]["2"].StepExecutions[0].Output.Data["value"])

	assert.Equal(3, len(pex.StepStatus["transform.second_step"]))
	assert.Equal("second_step: album name: Maybe This Place Is The Same And We're Just Changing", pex.StepStatus["transform.second_step"]["0"].StepExecutions[0].Output.Data["value"])
	assert.Equal("second_step: album name: Common Courtesy", pex.StepStatus["transform.second_step"]["1"].StepExecutions[0].Output.Data["value"])
	assert.Equal("second_step: album name: What You Don't See", pex.StepStatus["transform.second_step"]["2"].StepExecutions[0].Output.Data["value"])
}

func (suite *ModTestSuite) TestPipelineWithForEach() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestPipelineWithForEachContainer() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.parent_with_foreach", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 80, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}
	assert.Equal(5, len(pex.PipelineOutput["vals"].(map[string]interface{})))

	assert.Equal("name-0\n", pex.PipelineOutput["vals"].(map[string]interface{})["0"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["stdout"])
	assert.Equal("name-1\n", pex.PipelineOutput["vals"].(map[string]interface{})["1"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["stdout"])
	assert.Equal("name-2\n", pex.PipelineOutput["vals"].(map[string]interface{})["2"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["stdout"])
	assert.Equal("name-3\n", pex.PipelineOutput["vals"].(map[string]interface{})["3"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["stdout"])
	assert.Equal("name-4\n", pex.PipelineOutput["vals"].(map[string]interface{})["4"].(map[string]interface{})["output"].(map[string]interface{})["val"].(map[string]interface{})["stdout"])
}

func (suite *ModTestSuite) TestPipelineForEachTrippleNested() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}

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

	pipelineInput := modconfig.Input{}

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

	pipelineInput := modconfig.Input{
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

	assert.Equal("[\"a\",\"b\",\"c\"]", pex.PipelineOutput["val_request_body"])
}

func (suite *ModTestSuite) TestPipelineWithForLoop() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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
	assert.Equal("[1] bass", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["value"])
}

func (suite *ModTestSuite) TestJsonAsOutput() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.reduce_map", 100*time.Millisecond, pipelineInput)

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

	assert.Equal(3, len(pex.PipelineOutput["val"].(map[string]interface{})))
	assert.Equal("green_day: Green Day", pex.PipelineOutput["val"].(map[string]interface{})["green_day"].(map[string]interface{})["value"])
	assert.Equal("sum_41: Sum 41", pex.PipelineOutput["val"].(map[string]interface{})["sum_41"].(map[string]interface{})["value"])
	assert.Equal(0, len(pex.PipelineOutput["val"].(map[string]interface{})["blink_182"].(map[string]interface{})))
}

func (suite *ModTestSuite) TestListReduce() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.reduce_list", 100*time.Millisecond, pipelineInput)

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

	assert.Equal(6, len(pex.PipelineOutput["val"].(map[string]interface{})))
	assert.Equal(0, len(pex.PipelineOutput["val"].(map[string]interface{})["0"].(map[string]interface{})))
	assert.Equal(0, len(pex.PipelineOutput["val"].(map[string]interface{})["2"].(map[string]interface{})))
	assert.Equal(0, len(pex.PipelineOutput["val"].(map[string]interface{})["4"].(map[string]interface{})))

	assert.Equal("1: 2", pex.PipelineOutput["val"].(map[string]interface{})["1"].(map[string]interface{})["value"])
}

func (suite *ModTestSuite) TestNested() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestForEachEmptyAndNonCollection() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestPipelineTransformStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.pipeline_with_transform_step", 200*time.Millisecond, pipelineInput)
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

	assert.Equal(1, len(pex.StepStatus["transform.basic_transform"]))
	if _, ok := pex.StepStatus["transform.basic_transform"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue].(string); !ok {
		assert.Fail("Unable to convert output to string")
		return
	}
	assert.Equal("This is a simple transform step", pex.StepStatus["transform.basic_transform"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])

	assert.Equal(1, len(pex.StepStatus["transform.basic_transform_refers_param"]))
	if _, ok := pex.StepStatus["transform.basic_transform_refers_param"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue].(float64); !ok {
		assert.Fail("Unable to convert output to float64")
		return
	}
	assert.Equal(float64(10), pex.StepStatus["transform.basic_transform_refers_param"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])

	assert.Equal(1, len(pex.StepStatus["transform.depends_on_transform_step"]))
	assert.Equal(1, len(pex.StepStatus["transform.depends_on_transform_step"]["0"].StepExecutions))
	if _, ok := pex.StepStatus["transform.depends_on_transform_step"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue].(string); !ok {
		assert.Fail("Unable to convert output to string")
		return
	}
	assert.Equal("This is a simple transform step - test123", pex.StepStatus["transform.depends_on_transform_step"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])

	// Pipeline 2

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.pipeline_with_transform_step_string_list", 200*time.Millisecond, pipelineInput)
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

	assert.Equal(4, len(pex.StepStatus["transform.transform_test"]))
	if _, ok := pex.StepStatus["transform.transform_test"]["3"].StepExecutions[0].Output.Data[schema.AttributeTypeValue].(string); !ok {
		assert.Fail("Unable to convert output to string")
		return
	}
	assert.Equal("user is brian", pex.StepStatus["transform.transform_test"]["0"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])
	assert.Equal("user is freddie", pex.StepStatus["transform.transform_test"]["1"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])
	assert.Equal("user is john", pex.StepStatus["transform.transform_test"]["2"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])
	assert.Equal("user is roger", pex.StepStatus["transform.transform_test"]["3"].StepExecutions[0].Output.Data[schema.AttributeTypeValue])

	// Pipeline 3
	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.transform_step_for_map", 200*time.Millisecond, pipelineInput)
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

	assert.Equal(3, len(pex.StepStatus["transform.text_1"]))
	assert.Equal("janis joplin was 27", pex.PipelineOutput["text_1"].(map[string]interface{})["janis"].(map[string]interface{})["value"])
	assert.Equal("jimi hendrix was 27", pex.PipelineOutput["text_1"].(map[string]interface{})["jimi"].(map[string]interface{})["value"])
	assert.Equal("jerry garcia was 53", pex.PipelineOutput["text_1"].(map[string]interface{})["jerry"].(map[string]interface{})["value"])
}

func (suite *ModTestSuite) TestStepHttpTimeout() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.pipeline_with_http_timeout", 200*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("failed", pex.Status)

	// Step #1
	stepOutput := pex.StepStatus["http.http_with_timeout_string"]["0"].StepExecutions[0].Output
	stepError := stepOutput.Errors[0].Error
	assert.Contains(stepError.Detail, "Client.Timeout exceeded")

	// Step #2
	stepOutput = pex.StepStatus["http.http_with_timeout_number"]["0"].StepExecutions[0].Output
	stepError = stepOutput.Errors[0].Error
	assert.Contains(stepError.Detail, "Client.Timeout exceeded")

	// Step #3
	stepOutput = pex.StepStatus["http.http_with_timeout_string_unresolved"]["0"].StepExecutions[0].Output
	stepError = stepOutput.Errors[0].Error
	assert.Contains(stepError.Detail, "Client.Timeout exceeded")

	// Step #4
	stepOutput = pex.StepStatus["http.http_with_timeout_number_unresolved"]["0"].StepExecutions[0].Output
	stepError = stepOutput.Errors[0].Error
	assert.Contains(stepError.Detail, "Client.Timeout exceeded")
}

func (suite *ModTestSuite) TestStepSleep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.pipeline_with_sleep_step_int_duration", 200*time.Millisecond, pipelineInput)
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
	assert.Equal(1, len(pex.StepStatus["sleep.sleep_test"]))

	flowpipeMd := pex.StepStatus["sleep.sleep_test"]["0"].StepExecutions[0].Output.Flowpipe
	startTime := flowpipeMd[schema.AttributeTypeStartedAt].(time.Time)
	finishTime := flowpipeMd[schema.AttributeTypeFinishedAt].(time.Time)
	diff := finishTime.Sub(startTime)
	assert.Equal(float64(0), math.Floor(diff.Seconds()), "output does not match the provided duration")

}

func (suite *ModTestSuite) TestNestedPipelineErrorBubbleUp() {
	assert := assert.New(suite.T())
	_, cmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.validate_error", 500*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 50, "failed")
	if err != nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)
	assert.NotNil(pex.Errors)

	assert.NotNil(pex.StepStatus["pipeline.pipeline_step"]["0"].StepExecutions[0].Output.Errors)

	assert.NotNil(pex.PipelineOutput["errors"])
	assert.Equal(int(404), pex.PipelineOutput["errors"].([]modconfig.StepError)[0].Error.Status)
}

func (suite *ModTestSuite) TestModVars() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

func (suite *ModTestSuite) TestErrorRetry() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 3 times. First attempt + 2 retries
	assert.Equal(3, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorRetryWithIf() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_if", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 3 times. First attempt + 2 retries
	assert.Equal(3, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)

	// But we should only see the "final" error in the pipeline error
	assert.Equal(1, len(pex.Errors))
}

func (suite *ModTestSuite) TestErrorWithIfMultiStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_if_multi_step", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 3 times. First attempt + 2 retries
	assert.Equal(3, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))

	if len(pex.StepStatus["http.bad_http"]["0"].StepExecutions) != 3 {
		// Print some debugging info trying to understand why this pipeline test often fails
		jsonData, err := json.Marshal(pex.StepStatus["http.bad_http"]["0"].StepExecutions)
		if err != nil {
			log.Fatal(err)
		}

		// Convert JSON bytes to string and print
		jsonString := string(jsonData)
		fmt.Println(jsonString) //nolint:forbidigo // test code
		fmt.Println()           //nolint:forbidigo // test code
		fmt.Println()           //nolint:forbidigo // test code

		ex, err := execution.GetExecution(pipelineCmd.Event.ExecutionID)
		if err != nil {
			log.Fatal(err)
		}

		jsonData, err = json.Marshal(ex)
		if err != nil {
			log.Fatal(err)
		}

		jsonString = string(jsonData)
		fmt.Println(jsonString) //nolint:forbidigo // test code

		return
	}

	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)

	assert.Equal(1, len(pex.StepStatus["http.bad_http_2"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http_2"]["0"].StepExecutions[0].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http_2"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)

	// But we should only see the "final" error in the pipeline errors
	//
	// There are 3 steps, 2 errors out 1 successful
	assert.Equal(2, len(pex.Errors))
}

func (suite *ModTestSuite) TestErroWithIfNotMatch() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_if_not_match", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 1 time. There is retry block, but that retry block has an if that does not match.
	assert.Equal(1, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorRetryWithBackoff() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_backoff", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 60, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 3 times. First attempt + 2 retries (max attempts = 3)
	assert.Equal(3, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)

	step1EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].EndTime
	step2StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].StartTime

	duration := step2StartTime.Sub(step1EndTime)
	if duration < 2*time.Second {
		assert.Fail("The gap should be at least 2 seconds but " + duration.String())
	}

	step2EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].EndTime
	step3StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].StartTime

	duration = step3StartTime.Sub(step2EndTime)
	if duration < 2*time.Second {
		assert.Fail("The gap should be at least 2 seconds but " + duration.String())
	}
}

func (suite *ModTestSuite) TestErrorRetryWithLinearBackoff() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_linear_backoff", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 60, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 5 times. First attempt + 4 retries. Max attempts = 5
	assert.Equal(5, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[4].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[4].Output.Errors[0].Error.Status)

	step1EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].EndTime
	step2StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].StartTime

	duration := step2StartTime.Sub(step1EndTime)
	if duration < 100*time.Millisecond {
		assert.Fail("The gap should be at least 100ms but " + duration.String())
	}

	// the second attempt should be 100ms after the first one
	step2EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].EndTime
	step3StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].StartTime

	duration = step3StartTime.Sub(step2EndTime)
	if duration < 100*time.Millisecond {
		assert.Fail("The gap should be at least 100ms but " + duration.String())
	}

	step3EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].EndTime
	step4StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[3].StartTime

	duration = step4StartTime.Sub(step3EndTime)

	// Linear backoff, now it should be 200ms
	if duration < 200*time.Millisecond {
		assert.Fail("The gap should be at least 200ms but " + duration.String())
	}

	step4EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[3].EndTime
	step5StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[4].StartTime

	duration = step5StartTime.Sub(step4EndTime)

	// Linear backoff, now it should be 300ms
	if duration < 300*time.Millisecond {
		assert.Fail("The gap should be at least 300ms but " + duration.String())
	}
}

func (suite *ModTestSuite) TestErrorRetryWithExponentialBackoff() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_exponential_backoff", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 60, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	if pex.Status != "failed" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// The step should be executed 5 times. First attempt + 4 retries. Max attempts = 5
	assert.Equal(5, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[4].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[4].Output.Errors[0].Error.Status)

	step1EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].EndTime
	step2StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].StartTime

	duration := step2StartTime.Sub(step1EndTime)
	if duration < 100*time.Millisecond {
		assert.Fail("The gap should be at least 100ms but " + duration.String())
	}

	step2EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[1].EndTime
	step3StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].StartTime

	duration = step3StartTime.Sub(step2EndTime)
	if duration < 200*time.Millisecond {
		assert.Fail("The gap should be at least 200ms but " + duration.String())
	}

	step3EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[2].EndTime
	step4StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[3].StartTime

	duration = step4StartTime.Sub(step3EndTime)

	if duration < 400*time.Millisecond {
		assert.Fail("The gap should be at least 400ms but " + duration.String())
	}

	step4EndTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[3].EndTime
	step5StartTime := pex.StepStatus["http.bad_http"]["0"].StepExecutions[4].StartTime

	duration = step5StartTime.Sub(step4EndTime)

	if duration < 800*time.Millisecond {
		assert.Fail("The gap should be at least 800ms but " + duration.String())
	}

}

func (suite *ModTestSuite) TestTransformLoop() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.transform_loop", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal(3, len(pex.StepStatus["transform.foo"]["0"].StepExecutions))

	assert.Equal("loop: 0", pex.PipelineOutput["val_1"])
	assert.Equal("loop: 0", pex.PipelineOutput["val_2"])
	assert.Equal("loop: 1", pex.PipelineOutput["val_3"])
}

func (suite *ModTestSuite) TestForEachAndForEach() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.for_each_and_for_each", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// must check length to ensure that we're not duplicating the run (common issues with for_each)
	assert.Equal(3, len(pex.PipelineOutput["first"].(map[string]interface{})))
	assert.Equal(3, len(pex.PipelineOutput["second"].(map[string]interface{})))

	assert.Equal("bach", pex.PipelineOutput["first"].(map[string]interface{})["0"].(map[string]interface{})["value"])
	assert.Equal("mozart", pex.PipelineOutput["first"].(map[string]interface{})["1"].(map[string]interface{})["value"])
	assert.Equal("beethoven", pex.PipelineOutput["first"].(map[string]interface{})["2"].(map[string]interface{})["value"])

	assert.Equal("coltrane", pex.PipelineOutput["second"].(map[string]interface{})["0"].(map[string]interface{})["value"])
	assert.Equal("davis", pex.PipelineOutput["second"].(map[string]interface{})["1"].(map[string]interface{})["value"])
	assert.Equal("monk", pex.PipelineOutput["second"].(map[string]interface{})["2"].(map[string]interface{})["value"])
}

func (suite *ModTestSuite) TestErrorInForEach() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_in_for_each", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// There are 3 instances on the for_each, all of them failed just one time (no retry configured)
	assert.Equal(1, len(pex.StepStatus["http.bad_http"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["http.bad_http"]["1"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["http.bad_http"]["2"].StepExecutions))

	assert.Equal("failed", pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["1"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["http.bad_http"]["2"].StepExecutions[0].Output.Status)

	assert.Equal(404, pex.StepStatus["http.bad_http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["1"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["http.bad_http"]["2"].StepExecutions[0].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorInForEachNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_in_for_each_nested_pipeline", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// There are 3 instances on the for_each, all of them failed just one time (no retry configured)
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["1"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["2"].StepExecutions))

	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["2"].StepExecutions[0].Output.Status)

	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["pipeline.http"]["2"].StepExecutions[0].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorRetryWithNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_with_nested_pipeline", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// 4 executions in total. 1 initial attempt + 3 retries
	assert.Equal(4, len(pex.StepStatus["pipeline.http"]["0"].StepExecutions))

	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[1].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[2].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[3].Output.Status)

	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[1].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[2].Output.Errors[0].Error.Status)
	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[3].Output.Errors[0].Error.Status)

	step1EndTime := pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].EndTime
	step2StartTime := pex.StepStatus["pipeline.http"]["0"].StepExecutions[1].StartTime

	duration := step2StartTime.Sub(step1EndTime)
	if duration < 1*time.Second {
		assert.Fail("The gap should at least 1 second but " + duration.String())
	}

	step2EndTime := pex.StepStatus["pipeline.http"]["0"].StepExecutions[1].EndTime
	step3StartTime := pex.StepStatus["pipeline.http"]["0"].StepExecutions[2].StartTime

	duration = step3StartTime.Sub(step2EndTime)
	if duration < 1*time.Second {
		assert.Fail("The gap should at least 1 second but " + duration.String())
	}
}

func (suite *ModTestSuite) TestErrorInForEachNestedPipelineOneWorks() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_in_for_each_nested_pipeline_one_works", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// There are 3 instances on the for_each, all of them failed just one time (no retry configured)
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["1"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["2"].StepExecutions))

	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("finished", pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["2"].StepExecutions[0].Output.Status)

	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(0, len(pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Errors))
	assert.Equal(404, pex.StepStatus["pipeline.http"]["2"].StepExecutions[0].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorInForEachNestedPipelineOneWorksErrorIgnored() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_in_for_each_nested_pipeline_one_works_error_ignored", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	// There are 3 instances on the for_each, all of them failed just one time (no retry configured)
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["1"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["pipeline.http"]["2"].StepExecutions))

	assert.Equal("failed", pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("finished", pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Status)
	assert.Equal("failed", pex.StepStatus["pipeline.http"]["2"].StepExecutions[0].Output.Status)

	assert.Equal(404, pex.StepStatus["pipeline.http"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
	assert.Equal(0, len(pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Errors))
	assert.Equal(404, pex.StepStatus["pipeline.http"]["2"].StepExecutions[0].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorWithThrowSimple() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_simple", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// retry does not catch throw, so there should only be 1 step execution here
	assert.Equal(1, len(pex.Errors))
	assert.Equal("from throw block", pex.Errors[0].Error.Detail)
}

func (suite *ModTestSuite) TestErrorWithThrowButIgnored() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_but_ignored", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// Step throws an error, but it's ignored s the pipeline should finish rather than fail
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished status is: " + pex.Status)
		return
	}

	assert.Equal(0, len(pex.Errors))
}

func (suite *ModTestSuite) TestErrorWithMultipleThrows() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_multiple_throws", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	// retry does not catch throw, so there should only be 1 step execution here
	assert.Equal(1, len(pex.Errors))
	assert.Equal("from throw block bar", pex.Errors[0].Error.Detail)
}

func (suite *ModTestSuite) TestErrorWithThrowSimpleNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_simple_nested_pipeline", 500*time.Millisecond, pipelineInput)

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
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal(1, len(pex.Errors))
	assert.Equal("from throw block", pex.Errors[0].Error.Detail)
}

func (suite *ModTestSuite) TestPipelineWithTransformStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.pipeline_with_transform_step", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")

	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("This is a simple transform step", pex.PipelineOutput["basic_transform"])
	assert.Equal("This is a simple transform step - test123", pex.PipelineOutput["depends_on_transform_step"])
	assert.Equal(23, pex.PipelineOutput["number"])
}

func (suite *ModTestSuite) TestParamAny() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"param_any": "hello as string",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.any_param", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("hello as string", pex.PipelineOutput["val"])

	// now re-run the pipeline with param_any as an int
	pipelineInput = modconfig.Input{
		"param_any": 42,
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.any_param", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal(42, pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestTypedParamAny() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"param_any": "hello as string",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.typed_any_param", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal("hello as string", pex.PipelineOutput["val"])

	// now re-run the pipeline with param_any as an int
	pipelineInput = modconfig.Input{
		"param_any": 42,
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.typed_any_param", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	if pex.Status != "finished" {
		assert.Fail("Pipeline execution not finished")
		return
	}

	assert.Equal(42, pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestCredentialReference() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_aws", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// This is not redacted because we're looking for either field name or field value, and neither will hit the redaction list
	assert.Equal("aws_static_foo", pex.PipelineOutput["val_access_key"])

	// Check if the environment function is created successfully
	envMap := pex.PipelineOutput["val"].(map[string]interface{})

	assert.Equal("aws_static_foo", envMap["AWS_ACCESS_KEY_ID"])
	assert.Equal("aws_static_key_key_key", envMap["AWS_SECRET_ACCESS_KEY"])

	// Now load the execution from file, it should be redacted
	time.Sleep(50 * time.Millisecond)

	ex, err := execution.NewExecution(suite.ctx, execution.WithEvent(pipelineCmd.Event))
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]

	// This is not redacted because we're looking for either field name or field value, and neither will hit the redaction list
	assert.Equal("aws_static_foo", pex.PipelineOutput["val_access_key"])

	// Check if the environment function is created successfully
	envMap = pex.PipelineOutput["val"].(map[string]interface{})

	assert.Equal(sanitize.RedactedStr, envMap["AWS_ACCESS_KEY_ID"])
	assert.Equal(sanitize.RedactedStr, envMap["AWS_SECRET_ACCESS_KEY"])
}

func (suite *ModTestSuite) TestCredentialRedactionFromMemoryAndFile() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sensitive_one", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// We're reading straight from memory, not redacted
	assert.Equal("abc", pex.PipelineOutput["val"].(map[string]interface{})["AWS_ACCESS_KEY_ID"])
	assert.Equal("def", pex.PipelineOutput["val"].(map[string]interface{})["AWS_SECRET_ACCESS_KEY"])
	assert.Equal("EAACEdEose0cBA1234FAKE1234", pex.PipelineOutput["val"].(map[string]interface{})["facebook_access_token"])
	assert.Equal("AKIAFAKEFAKEFAKEFAKE", pex.PipelineOutput["val"].(map[string]interface{})["pattern_match_aws_access_key_id"])

	// not redacted
	assert.Equal("AKFFFAKEFAKEFAKEFAKE", pex.PipelineOutput["val"].(map[string]interface{})["close_but_no_cigar"])
	assert.Equal("two", pex.PipelineOutput["val"].(map[string]interface{})["one"])

	// Now load the execution from file, it should be redacted
	time.Sleep(50 * time.Millisecond)

	ex, err := execution.NewExecution(suite.ctx, execution.WithEvent(pipelineCmd.Event))
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]

	// Since we are reading this from file, they should be redacted
	assert.Equal(sanitize.RedactedStr, pex.PipelineOutput["val"].(map[string]interface{})["AWS_ACCESS_KEY_ID"])
	assert.Equal(sanitize.RedactedStr, pex.PipelineOutput["val"].(map[string]interface{})["AWS_SECRET_ACCESS_KEY"])
	assert.Equal(sanitize.RedactedStr, pex.PipelineOutput["val"].(map[string]interface{})["facebook_access_token"])
	assert.Equal(sanitize.RedactedStr, pex.PipelineOutput["val"].(map[string]interface{})["pattern_match_aws_access_key_id"])

	// not redacted
	assert.Equal("AKFFFAKEFAKEFAKEFAKE", pex.PipelineOutput["val"].(map[string]interface{})["close_but_no_cigar"])
	assert.Equal("two", pex.PipelineOutput["val"].(map[string]interface{})["one"])
}

func (suite *ModTestSuite) TestExcludeSHARedaction() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.github_sha_exclude_redaction", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("https://github.com/turbot/flowpipe/commit/7a2c8fd9789a9b6rc8f29c41b42036823e2fceab", pex.PipelineOutput["sha"].(string))
}

func (suite *ModTestSuite) XTestRunMultiplePipelinesAtTheSameTimeWithDifferentInput() {
	// TODO
}

func (suite *ModTestSuite) XTestRunMultiplePipelinesRunningAtTheSameTimeUseSleepToEnsureRun() {
	// TODO
}

func (suite *ModTestSuite) XTestBufferTokenTooLargeFromFile() {
	// TODO
}

func (suite *ModTestSuite) XTestCredentialWithOptionalParamFromFile() {
	// TODO
}

func (suite *ModTestSuite) TestCredentialWithOptionalParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	os.Setenv("SLACK_TOKEN", "test.1.2.3")
	// This was crashing because of the optional param
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_slack", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// Reading from memory, not redacted
	assert.Equal("test.1.2.3", pex.PipelineOutput["slack_token"])

	// Now load the execution from file, it should be redacted
	time.Sleep(50 * time.Millisecond)

	ex, err := execution.NewExecution(suite.ctx, execution.WithEvent(pipelineCmd.Event))
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]

	assert.Equal(sanitize.RedactedStr, pex.PipelineOutput["slack_token"])

	//
	pipelineInput = modconfig.Input{}
	os.Setenv("GITLAB_TOKEN", "glpat-gsfio3wtyr92364ifkw")

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_gitlab", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// From memory not redacted
	assert.Equal("glpat-gsfio3wtyr92364ifkw", pex.PipelineOutput["gitlab_token"])

	// Now load the execution from file, it should be redacted
	time.Sleep(50 * time.Millisecond)

	ex, err = execution.NewExecution(suite.ctx, execution.WithEvent(pipelineCmd.Event))
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]
	assert.Equal(sanitize.RedactedStr, pex.PipelineOutput["gitlab_token"])

	//
	pipelineInput = modconfig.Input{}
	os.Setenv("ABUSEIPDB_API_KEY", "bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d")

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_abuseipdb", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("bfc6f1c42dsfsdfdxxxx26977977b2xxxsfsdda98f313c3d389126de0d", pex.PipelineOutput["abuseipdb_api_key"])

	//
	pipelineInput = modconfig.Input{}
	os.Setenv("CLICKUP_TOKEN", "pk_616_L5H36X3CXXXXXXXWEAZZF0NM5")

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_clickup", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("pk_616_L5H36X3CXXXXXXXWEAZZF0NM5", pex.PipelineOutput["clickup_token"])

	//
	pipelineInput = modconfig.Input{}
	os.Setenv("VAULT_TOKEN", "hsv-fkshfgskhf")

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_vault", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("hsv-fkshfgskhf", pex.PipelineOutput["vault_token"])
}

func (suite *ModTestSuite) TestMultipleCredential() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// Set env variable
	os.Setenv("SLACK_TOKEN", "slack123")
	os.Setenv("GITLAB_TOKEN", "gitlab123")
	os.Setenv("CLICKUP_TOKEN", "pk_616_L5H36X3CXXXXXXXWEAZZF0NM5")

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.multiple_credentials", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal("jgdjslgjdljgldjglsdjl", pex.PipelineOutput["slack_token_val"])
	assert.Equal("slack123", pex.PipelineOutput["slack_default_token"])
	assert.Equal("glpat-ksgfhwekty389398hdgkhgkhdgk", pex.PipelineOutput["gitlab_token_val"])
	assert.Equal("gitlab123", pex.PipelineOutput["gitlab_default_token"])
	assert.Equal("pk_616_L5H36X3CXXXXXXXWEAZZF0NM5", pex.PipelineOutput["clickup_token_val"])
}

func (suite *ModTestSuite) TestBadContainerStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.with_bad_container", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("failed", pex.Status)
	assert.NotNil(pex.Errors)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Detail, "InvalidClientTokenId")
}

func (suite *ModTestSuite) TestBadContainerStepWithIsErrorFunc() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.with_bad_container_with_is_error", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1500*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// The pipeline step should faile due to the invalid credentials, but the pipeline should continue
	assert.Equal(1, len(pex.StepStatus["pipeline.create_s3_bucket"]["0"].StepExecutions))
	assert.Equal("failed", pex.StepStatus["pipeline.create_s3_bucket"]["0"].StepExecutions[0].Output.Status)
	assert.Equal(460, pex.StepStatus["pipeline.create_s3_bucket"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)

	// The second step checks if the first step failed and it should be skipped
	assert.Equal(1, len(pex.StepStatus["pipeline.delete_s3_bucket"]["0"].StepExecutions))
	assert.Equal("skipped", pex.StepStatus["pipeline.delete_s3_bucket"]["0"].StepExecutions[0].Output.Status)
}

func (suite *ModTestSuite) TestContainerStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal(float64(60000), input[schema.AttributeTypeTimeout].(float64))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(float64(64), input[schema.AttributeTypeMemoryReservation].(float64))
	assert.Equal(float64(256), input[schema.AttributeTypeMemorySwap].(float64))
	assert.Equal(float64(10), input[schema.AttributeTypeMemorySwappiness].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]container.OutputLine); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]container.OutputLine)
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].Stream)
	assert.Equal("Line 1\n", lines[0].Line)

	assert.Equal("stdout", lines[1].Stream)
	assert.Equal("Line 2\n", lines[1].Line)

	assert.Equal("stdout", lines[2].Stream)
	assert.Equal("Line 3\n", lines[2].Line)
}

func (suite *ModTestSuite) TestContainerStepWithLoop() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.container_with_loop", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.pipe"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal(float64(60000), input[schema.AttributeTypeTimeout].(float64))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(float64(64), input[schema.AttributeTypeMemoryReservation].(float64))
	assert.Equal(float64(256), input[schema.AttributeTypeMemorySwap].(float64))
	assert.Equal(float64(10), input[schema.AttributeTypeMemorySwappiness].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.pipe"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	// the second run should have a different memory

	// first loop (after initial execution)
	input = pex.StepStatus["container.pipe"]["0"].StepExecutions[1].Input
	assert.Equal(float64(150), input[schema.AttributeTypeMemory].(float64))

	output = pex.StepStatus["container.pipe"]["0"].StepExecutions[1].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])

	// second loop
	input = pex.StepStatus["container.pipe"]["0"].StepExecutions[2].Input
	assert.Equal(float64(151), input[schema.AttributeTypeMemory].(float64))

	output = pex.StepStatus["container.pipe"]["0"].StepExecutions[2].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
}

func (suite *ModTestSuite) TestContainerStepWithLoopUpdateCmd() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.container_with_loop_update_cmd", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.pipe"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal(float64(60000), input[schema.AttributeTypeTimeout].(float64))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(float64(64), input[schema.AttributeTypeMemoryReservation].(float64))
	assert.Equal(float64(256), input[schema.AttributeTypeMemorySwap].(float64))
	assert.Equal(float64(10), input[schema.AttributeTypeMemorySwappiness].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.pipe"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	// the second run should have a different memory

	// first loop (after initial execution)
	input = pex.StepStatus["container.pipe"]["0"].StepExecutions[1].Input
	assert.Equal(float64(150), input[schema.AttributeTypeMemory].(float64))

	output = pex.StepStatus["container.pipe"]["0"].StepExecutions[1].Output
	assert.Equal("finished", output.Status)
	assert.Equal("0\n", output.Data["stdout"])

	// second loop
	input = pex.StepStatus["container.pipe"]["0"].StepExecutions[2].Input
	assert.Equal(float64(151), input[schema.AttributeTypeMemory].(float64))

	output = pex.StepStatus["container.pipe"]["0"].StepExecutions[2].Output
	assert.Equal("finished", output.Status)
	assert.Equal("1\n", output.Data["stdout"])
}

func (suite *ModTestSuite) TestContainerStepWithParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step_with_param", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal(float64(60000), input[schema.AttributeTypeTimeout].(float64))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(float64(64), input[schema.AttributeTypeMemoryReservation].(float64))
	assert.Equal(float64(256), input[schema.AttributeTypeMemorySwap].(float64))
	assert.Equal(float64(10), input[schema.AttributeTypeMemorySwappiness].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]container.OutputLine); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]container.OutputLine)
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].Stream)
	assert.Equal("Line 1\n", lines[0].Line)

	assert.Equal("stdout", lines[1].Stream)
	assert.Equal("Line 2\n", lines[1].Line)

	assert.Equal("stdout", lines[2].Stream)
	assert.Equal("Line 3\n", lines[2].Line)
}

func (suite *ModTestSuite) TestContainerStepWithParamOverride() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"user":    "root",
		"timeout": 120000,
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step_with_param_override", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal(float64(120000), input[schema.AttributeTypeTimeout].(float64))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]container.OutputLine); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]container.OutputLine)
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].Stream)
	assert.Equal("Line 1\n", lines[0].Line)

	assert.Equal("stdout", lines[1].Stream)
	assert.Equal("Line 2\n", lines[1].Line)

	assert.Equal("stdout", lines[2].Stream)
	assert.Equal("Line 3\n", lines[2].Line)
}

func (suite *ModTestSuite) TestContainerStepMissingImage() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step_missing_image", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("failed", pex.Status)
	assert.NotNil(pex.Errors)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Detail, "Container input must define 'image'")
}

func (suite *ModTestSuite) TestContainerStepInvalidMemory() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step_invalid_memory", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("failed", pex.Status)
	assert.NotNil(pex.Errors)
	assert.Equal(1, len(pex.Errors))
	assert.Equal(int(500), pex.Errors[0].Error.Status)
	assert.Contains(pex.Errors[0].Error.Detail, "Minimum memory limit allowed is 6MB")
}

func (suite *ModTestSuite) TestContainerStepWithStringTimeout() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step_with_string_timeout", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal("60s", input[schema.AttributeTypeTimeout].(string))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(float64(64), input[schema.AttributeTypeMemoryReservation].(float64))
	assert.Equal(float64(256), input[schema.AttributeTypeMemorySwap].(float64))
	assert.Equal(float64(10), input[schema.AttributeTypeMemorySwappiness].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]container.OutputLine); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]container.OutputLine)
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].Stream)
	assert.Equal("Line 1\n", lines[0].Line)

	assert.Equal("stdout", lines[1].Stream)
	assert.Equal("Line 2\n", lines[1].Line)

	assert.Equal("stdout", lines[2].Stream)
	assert.Equal("Line 3\n", lines[2].Line)
}

func (suite *ModTestSuite) TestContainerStepWithParamStringTimeout() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_container_step_with_string_timeout_with_param", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	// Validate inputs
	input := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Input
	assert.Equal("alpine:3.7", input[schema.AttributeTypeImage].(string))
	assert.Equal("root", input[schema.AttributeTypeUser].(string))
	assert.Equal("60s", input[schema.AttributeTypeTimeout].(string))
	assert.Equal(float64(128), input[schema.AttributeTypeMemory].(float64))
	assert.Equal(float64(64), input[schema.AttributeTypeMemoryReservation].(float64))
	assert.Equal(float64(256), input[schema.AttributeTypeMemorySwap].(float64))
	assert.Equal(float64(10), input[schema.AttributeTypeMemorySwappiness].(float64))
	assert.Equal(false, input[schema.AttributeTypeReadOnly].(bool))

	if _, ok := input[schema.AttributeTypeCmd].([]interface{}); !ok {
		assert.Fail("Cmd should be an array of strings")
	}
	assert.Equal(3, len(input[schema.AttributeTypeCmd].([]interface{})))

	if _, ok := input[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
		assert.Fail("Cmd should be a map")
	}
	assert.Equal("bar", input[schema.AttributeTypeEnv].(map[string]interface{})["FOO"].(string))

	output := pex.StepStatus["container.container_test_1"]["0"].StepExecutions[0].Output
	assert.Equal("finished", output.Status)
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Data["stdout"])
	assert.Equal("", output.Data["stderr"])
	assert.Equal(0, output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]container.OutputLine); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]container.OutputLine)
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].Stream)
	assert.Equal("Line 1\n", lines[0].Line)

	assert.Equal("stdout", lines[1].Stream)
	assert.Equal("Line 2\n", lines[1].Line)

	assert.Equal("stdout", lines[2].Stream)
	assert.Equal("Line 3\n", lines[2].Line)
}

func (suite *ModTestSuite) XTestBufferTokenTooLargeMemory() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// With in memory execution, we shouldn't be hitting the buffer too large issue anymore
	_, _, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.big_data", 500*time.Millisecond, pipelineInput)

	assert.Nil(err)

	_, pipelineCmd, _ := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.echo_one", 100*time.Millisecond, pipelineInput)
	_, pipelineCmd2, _ := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.echo_one", 100*time.Millisecond, pipelineInput)

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 20, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("Hello World from Depend A", pex.PipelineOutput["echo_one_output"])

	// value should be: ${step.echo.var_one.text} + ${var.var_depend_a_one}
	assert.Equal("Hello World from Depend A: this is the value of var_one + this is the value of var_one", pex.PipelineOutput["echo_one_output_val_var_one"])

	_, pex2, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd2.Event, pipelineCmd2.PipelineExecutionID, 100*time.Millisecond, 20, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("Hello World from Depend A", pex2.PipelineOutput["echo_one_output"])

	// value should be: ${step.echo.var_one.text} + ${var.var_depend_a_one}
	assert.Equal("Hello World from Depend A: this is the value of var_one + this is the value of var_one", pex2.PipelineOutput["echo_one_output_val_var_one"])
}

func (suite *ModTestSuite) TestBadHttpNotIgnored() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.bad_http_not_ignored", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.NotNil(pex.Errors)
	assert.Equal(1, len(pex.Errors))
	assert.Equal(int(404), pex.Errors[0].Error.Status)
	// the first step failed
	assert.Equal(1, len(pex.StepStatus["http.my_step_1"]["0"].Failed))
	assert.Equal(0, len(pex.StepStatus["http.my_step_1"]["0"].Finished))
	// the second step won't be started because the error is not ignored so the pipeline will fail
	// before the second step start
	assert.Equal(0, len(pex.StepStatus["transform.bad_http"]))
}

func (suite *ModTestSuite) TestInaccessibleFail() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.inaccessible_fail", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.NotNil(pex.Errors)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Detail, "This object does not have an attribute named \"value\"")

	// The error is that value does not exist, this
	// value = step.http.will_fail.value
	//
	// is the source of the failure because step.http.will_fail.value does not exist.
	//
	// the step.http.will_fail itself is not "failed" because there's ignore error directive
	assert.Equal(int(500), pex.Errors[0].Error.Status)

	// the first step has ignore error, so it will technically be "finished"
	assert.Equal(0, len(pex.StepStatus["http.will_fail"]["0"].Failed))
	assert.Equal(1, len(pex.StepStatus["http.will_fail"]["0"].Finished))

	// The step status for transform.will_not_run should be nil because we failed in generating the input for the step
	assert.Nil(pex.StepStatus["transform.will_not_run"])
}

func (suite *ModTestSuite) TestInaccessibleOk() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.inaccessible_ok", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 500*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))

	// the first step has ignore error, so it will technically be "finished"
	assert.Equal(0, len(pex.StepStatus["http.will_fail"]["0"].Failed))
	assert.Equal(1, len(pex.StepStatus["http.will_fail"]["0"].Finished))

	// the second step is actualy OK because it's referring to:
	//
	// step.http.will_fail
	//
	// which is OK
	assert.Equal(0, len(pex.StepStatus["transform.will_not_run"]["0"].Failed))
	assert.Equal(1, len(pex.StepStatus["transform.will_not_run"]["0"].Finished))
}

func (suite *ModTestSuite) TestNestedPipelineParamMismatched() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.parent_pipeline_param_mismatch", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Equal("unknown parameter specified 'invalid_param'", pex.Errors[0].Error.Detail)
}

func (suite *ModTestSuite) TestPipelineOutputShouldNotBeCalculatedOnError() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_error", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Equal(404, pex.Errors[0].Error.Status)
	assert.Nil(pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestPipelineIgnoredErrorOutputShouldBeCalculated() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_error_ignored", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal("should be calculated", pex.PipelineOutput["val"])

	// make sure that the step did fail and there's an error
	assert.Equal(1, len(pex.StepStatus["http.does_not_exist"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["http.does_not_exist"]["0"].StepExecutions[0].Output.Errors))
	assert.Equal(404, pex.StepStatus["http.does_not_exist"]["0"].StepExecutions[0].Output.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestPipelineFailedOutputCalculation() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline fail its output calculation, so it should be marked as "failed" with as many output calculated as possible
	//
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.failed_output_calc", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(2, len(pex.Errors))
	assert.Equal("echo that works", pex.PipelineOutput["val_ok"])
	assert.Equal("this works", pex.PipelineOutput["val_ok_two"])
}

func (suite *ModTestSuite) TestNestedPipelineWithEmptyOutput() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the nested pipeline has no output so the error calculation fail
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.parent_with_child_with_no_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("ok", pex.PipelineOutput["val"].(map[string]interface{})["call_child"])
}

func (suite *ModTestSuite) TestStepOutputShouldNotCalculateIfError() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the nested pipeline has no output so the error calculation fail
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.step_output_should_not_calculate_if_error", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.StepStatus["http.bad"]["0"].StepExecutions))

	// pipeline output should NOT be calculated, the pipeline failed due to step error
	assert.Nil(pex.PipelineOutput["val"])
	assert.Nil(pex.StepStatus["http.bad"]["0"].StepExecutions[0].StepOutput["val"])
}

func (suite *ModTestSuite) TestStepOutputCalculateIfErrorBecauseErrorIsIgnored() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the nested pipeline has no output so the error calculation fail
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.step_output_should_be_calculated_because_step_error_is_ignored", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("finished", pex.Status)
	assert.Equal(1, len(pex.StepStatus["http.bad"]["0"].StepExecutions))

	// pipeline output should NOT be calculated, the pipeline failed due to step error
	assert.Equal("pipeline: should be calculated", pex.PipelineOutput["val"])
	assert.Equal("step: should be calculated", pex.StepStatus["http.bad"]["0"].StepExecutions[0].StepOutput["val"])
}

func (suite *ModTestSuite) TestErrorWithThrowInvalidMessage() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	//
	// This pipeline step has a throw that the IF condition is not met, so it should not throw the error,
	// however there was a bug that if the IF condition is not met, Flowpipe will still try to calculate the
	// rest of the throw block
	//
	// What this test is trying to ensure is that if the IF condition in the throw block is NOT met, do not try to resolve
	// the "message" attribute
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_invalid_message", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Equal("404 Not Found", pex.Errors[0].Error.Detail)
}

func (suite *ModTestSuite) TestEmptySlice() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.empty_slice", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	valOutput, ok := pex.PipelineOutput["val"].([]interface{})
	if !ok {
		assert.Fail("val should be a slice")
		return
	}
	assert.Equal(0, len(valOutput))

	emptyOutput, ok := pex.PipelineOutput["empty_output"].([]interface{})
	if !ok {
		assert.Fail("empty_output should be a slice")
		return
	}
	assert.Equal(0, len(emptyOutput))

	emptyOutput, ok = pex.PipelineOutput["empty_input_number"].([]interface{})
	if !ok {
		assert.Fail("empty_output should be a slice")
		return
	}
	assert.Equal(0, len(emptyOutput))
}

func (suite *ModTestSuite) TestRedact() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.redact", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal("this should not be redacted: 9d9bdaa9-fa12-436b-bce8-9e783695b3ff", pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestEmptyArrayResponseFromHttpServer() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.empty_slice_http", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))

	valOutput, ok := pex.PipelineOutput["val"].([]interface{})
	if !ok {
		assert.Fail("val should be a slice")
		return
	}

	assert.Equal(0, len(valOutput))
}

func (suite *ModTestSuite) TestReferToArguments() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the step refer to an argument (input) of another step, we used to not have the input in the
	// eval context
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.refer_to_arguments", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))

	assert.Equal("http://api.open-notify.org/astros.json", pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestSimpleErrorIgnoredMultiSteps() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the step refer to an argument (input) of another step, we used to not have the input in the
	// eval context
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_error_ignored_multi_steps", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))

	// Ensure that there's no retry
	assert.Equal("should be calculated", pex.PipelineOutput["val"])
	assert.Equal("should exist", pex.PipelineOutput["val_two"])

	assert.Equal(0, pex.StepStatus["http.does_not_exist"]["0"].FailCount(), "error is ignored, should not be in failed state")
	assert.Equal(1, pex.StepStatus["http.does_not_exist"]["0"].FinishCount(), "error is ignored, should be in finished state")
}

func (suite *ModTestSuite) TestErrorRetryFailedCalculatingOutputBlock() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the step refer to an argument (input) of another step, we used to not have the input in the
	// eval context
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_failed_calculating_output_block", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))

	// Ensure that there's no retry
	assert.Equal(1, len(pex.StepStatus["transform.one"]["0"].StepExecutions))
	assert.Equal(fpconstants.StateFailed, pex.StepStatus["transform.one"]["0"].StepExecutions[0].Output.Status)
	assert.Equal(fpconstants.FailureModeFatal, pex.StepStatus["transform.one"]["0"].StepExecutions[0].Output.FailureMode)
}

func (suite *ModTestSuite) TestErrorRetryFailedCalculatingOutputBlockIgnoredErrorShouldNotBeFollowed() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// This pipeline used to fail because the step refer to an argument (input) of another step, we used to not have the input in the
	// eval context
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_failed_calculating_output_block_ignored_error_should_not_be_followed", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))

	// Ensure that there's no retry
	assert.Equal(1, len(pex.StepStatus["transform.one"]["0"].StepExecutions))
	assert.Equal(fpconstants.StateFailed, pex.StepStatus["transform.one"]["0"].StepExecutions[0].Output.Status)
	assert.Equal(fpconstants.FailureModeFatal, pex.StepStatus["transform.one"]["0"].StepExecutions[0].Output.FailureMode)

	assert.Equal(0, len(pex.StepStatus["transform.two"]), "transform.two should not be executed. It depends on transform.one. Although transform.one has ignore=error directive, the output block calculation failed. As per issue #419 we've decided that this type of failure ignores ignore=true directive")
}

func (suite *ModTestSuite) TestErrorWithThrowFailingToCalculateThrow() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	//
	// This pipeline step has a throw that the IF condition is met but then it failed to calculate the throw block
	//
	// The step where the throw is also erroring out and there's an ignore = true directive. However the failure of calculating the throw bypasses the
	// ignore = true directive and fail the step
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_failing_to_calculate_throw", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)

	// There are 2 errors, the error from the step primitive (HTTP 404) and the error calculating the throw block
	assert.Equal(2, len(pex.Errors))
	assert.Equal("404 Not Found", pex.Errors[0].Error.Detail)
	assert.Equal(500, pex.Errors[1].Error.Status)
}

func (suite *ModTestSuite) TestErrorWithThrowDoesNotIgnore() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_does_not_ignore", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)

	assert.Equal(1, len(pex.Errors))
	assert.Equal(500, pex.Errors[0].Error.Status)
}

func (suite *ModTestSuite) TestErrorWithThrowDoesNotRetry() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_with_throw_does_not_retry", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)

	assert.Equal(1, len(pex.Errors))
	assert.Equal(500, pex.Errors[0].Error.Status)

	// make sure that there's only 1 execution and the retry isn't happening
	assert.Equal(1, len(pex.StepStatus["transform.good_step"]["0"].StepExecutions))
	assert.Equal(1, len(pex.StepStatus["transform.foo"]["0"].StepExecutions))
}

func (suite *ModTestSuite) TestLoopBlockEvaluationError() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_block_evaluation_error", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)

	assert.Equal(1, len(pex.Errors))
	assert.Equal(500, pex.Errors[0].Error.Status)

	// make sure that there's only 1 execution and the retry isn't happening
	assert.Equal(1, len(pex.StepStatus["transform.one"]["0"].StepExecutions))
	assert.Equal(0, len(pex.StepStatus["transform.two"]), "transform.two should not be executed even if there's ignore = true directive in transform.one because the error is in the loop block")
}

func (suite *ModTestSuite) TestErrorRetryEvaluationBlock() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.error_retry_evaluation_block", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)

	assert.Equal(2, len(pex.Errors), "there should be 2 errors. The first error is the HTTP 404 error and the second error is the error trying to render the retry block")
	assert.Equal(404, pex.Errors[0].Error.Status)
	assert.Equal(500, pex.Errors[1].Error.Status)

	// make sure that there's only 1 execution and the retry isn't happening
	assert.Equal(1, len(pex.StepStatus["http.one"]["0"].StepExecutions))
}

func (suite *ModTestSuite) TestSimpleErrorIgnoredWithIfDoesNotMatch() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_error_ignored_with_if_does_not_match", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)

	assert.Equal(1, len(pex.Errors), "1 error no retry, error ignore = true directive is ignored because the if statement does not match")
	assert.Equal(404, pex.Errors[0].Error.Status)

	// make sure that there's only 1 execution and the retry isn't happening
	assert.Equal(1, len(pex.StepStatus["http.does_not_exist"]["0"].StepExecutions))
	assert.Nil(pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestSimpleErrorIgnoredWithIfMatches() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_error_ignored_with_if_matches", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors), "no error in pipeline, error in step is ignored")

	// make sure that there's only 1 execution and the retry isn't happening
	assert.Equal(1, len(pex.StepStatus["http.does_not_exist"]["0"].StepExecutions))
	assert.Equal("should be calculated", pex.PipelineOutput["val"])
}

func (suite *ModTestSuite) TestTier3PipelinesNotRunnableDirectly() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "mod_depend_b.pipeline.echo_from_depend_b", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "definition not found: mod_depend_b.pipeline.echo_from_depend_b")
}

func (suite *ModTestSuite) TestLoopWithResultArguments() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.request_body_loop", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(3, len(pex.StepExecutions))
}

func (suite *ModTestSuite) TestSqliteQuery() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqlite_query", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(5, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("John", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
	assert.Equal("Jane", pex.PipelineOutput["val"].([]interface{})[1].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestSqliteQueryPathAlternateB() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqlite_query_path_alternate_b", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(5, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("John", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
	assert.Equal("Jane", pex.PipelineOutput["val"].([]interface{})[1].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestSqliteQueryPathAlternateC() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqlite_query_path_alternate_c", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(5, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("John", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
	assert.Equal("Jane", pex.PipelineOutput["val"].([]interface{})[1].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestSqliteQueryWithParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqllite_query_wity_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(1, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("Jane", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestSqliteQueryWithParam2() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqllite_query_wity_param_2", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(2, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("Jane", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
	assert.Equal("John", pex.PipelineOutput["val"].([]interface{})[1].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestSqliteQueryWithParam2b() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"name_2": "Jill",
	}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqllite_query_wity_param_2", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(2, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("Jane", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
	assert.Equal("Jill", pex.PipelineOutput["val"].([]interface{})[1].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestDuckDBQuery() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.duckdb_query", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)

	// There should be 3 step executions
	assert.Equal(1, len(pex.StepExecutions))
	assert.Equal(3, len(pex.PipelineOutput["val"].([]interface{})))
	assert.Equal("John", pex.PipelineOutput["val"].([]interface{})[0].(map[string]interface{})["name"].(string))
	assert.Equal("Adam", pex.PipelineOutput["val"].([]interface{})[1].(map[string]interface{})["name"].(string))
}

func (suite *ModTestSuite) TestSqliteQueryTimeout() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.sqlite_query_with_timeout", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "Timeout: Query execution exceeded timeout")
}

func (suite *ModTestSuite) TestPipelineStepLoopWithArgs() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_pipeline_loop_with_args", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("finished", pex.Status)
	assert.Equal(4, len(pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions))

	// Iteration 0
	output := pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[0].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! iteration index 0", output["greet_world"].(string))

	// Iteration 1
	output = pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[1].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! loop index_0 0", output["greet_world"].(string))

	// Iteration 2
	output = pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[2].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! loop index_1 1", output["greet_world"].(string))

	// Iteration 3
	output = pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[3].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! loop index_2 2", output["greet_world"].(string))
}

func (suite *ModTestSuite) TestPipelineStepLoopWithArgsLiteral() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// The loop block has result.request_body which is an argument, we were only adding the output attributes rather than the argument attributes
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.simple_pipeline_loop_with_arg_literal", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("finished", pex.Status)
	assert.Equal(4, len(pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions))

	// Iteration 0
	output := pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[0].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! iteration index 0", output["greet_world"].(string))

	// Iteration 1
	output = pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[1].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! loop index 1", output["greet_world"].(string))

	// Iteration 2
	output = pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[2].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! loop index 1", output["greet_world"].(string))

	// Iteration 3
	output = pex.StepStatus["pipeline.repeat_pipeline_loop_test"]["0"].StepExecutions[3].Output.Data["output"].(map[string]interface{})
	assert.Equal("Hello world! loop index 1", output["greet_world"].(string))
}

// Input Step Validation Tests
func (suite *ModTestSuite) TestInputWithNoOptionsButtonType() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_with_no_options_button_type", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "Input type 'button' requires options, no options were defined")
}

func (suite *ModTestSuite) TestInputWithNoOptionsTextType() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_with_no_options_text_type", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("started", pex.Status)
	assert.Equal(0, len(pex.Errors))
}

func (suite *ModTestSuite) TestInputWithSlackNotifierNoChannelToken() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_with_slack_notifier_no_channel_set", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "slack notifications require a channel when using token auth, channel was not set")
}

// TODO: This actually calls slack... figure out a mocked approach or not test success path?
func (suite *ModTestSuite) XTestInputWithSlackNotifierNoChannelWebhook() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_with_slack_notifier_no_channel_set_wh", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(0, len(pex.Errors))
	// assert.Equal(1, len(pex.Errors))
	// assert.Contains(pex.Errors[0].Error.Error(), "slack server error: 404 Not Found")
}

func (suite *ModTestSuite) TestInputWithEmailNotifierNoRecipients() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.input_with_email_notifier_no_recipients", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "email notifications require recipients; one of 'to', 'cc' or 'bcc' need to be set")
}

func (suite *ModTestSuite) TestMessageStepWithThrow() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.message_step_with_throw", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 40*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "throw here")
	assert.Equal(461, pex.Errors[0].Error.Status)
}

// TODO: This actually calls slack... figure out a mocked approach or not test success path?
func (suite *ModTestSuite) TestMessageStepBadSlack() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.message_step_bad_slack_call", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	// assert.Equal(1, len(pex.Errors))
	assert.Contains(pex.Errors[0].Error.Error(), "slack server error: 404 Not Found")
}

func (suite *ModTestSuite) TestMessageStepBadSlackIgnored() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.message_step_bad_slack_call_ignored", 100*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	// pipeline finished successfully but the step failed
	assert.Equal("failed", pex.StepStatus["message.message"]["0"].StepExecutions[0].Status)
}

func TestModTestingSuite(t *testing.T) {
	suite.Run(t, &ModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
