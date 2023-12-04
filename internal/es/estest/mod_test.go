package estest

// Basic imports
import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/cache"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
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

	// sets app specific constants defined in pipe-fittings
	localcmdconfig.SetAppSpecificConstants()

	app_specific.InstallDir = app_specific.DefaultInstallDir

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pipelineDirPath := path.Join(cwd, "test_suite_mod")

	viper.GetViper().Set(constants.ArgModLocation, pipelineDirPath)

	// clear the output dir before each test
	outputPath := filepaths.EventStoreDir()

	// Check if the directory exists
	_, err = os.Stat(outputPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(outputPath)
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

	err = docker.Initialize(ctx)
	if err != nil {
		panic(err)
	}

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

func (suite *ModTestSuite) XTestLoopWithForEach() {
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
	assert.Equal(3, len(pex.StepStatus["echo.repeat"]))
	assert.Equal(4, len(pex.StepStatus["echo.repeat"]["0"].StepExecutions))
	assert.Equal(4, len(pex.StepStatus["echo.repeat"]["1"].StepExecutions))
	assert.Equal(4, len(pex.StepStatus["echo.repeat"]["2"].StepExecutions))

	// Print out the step status if the step executions is not exactly 4
	if len(pex.StepStatus["echo.repeat"]["0"].StepExecutions) != 4 ||
		len(pex.StepStatus["echo.repeat"]["1"].StepExecutions) != 4 ||
		len(pex.StepStatus["echo.repeat"]["2"].StepExecutions) != 4 {
		s, err := prettyjson.Marshal(pex.StepStatus["echo.repeat"])

		if err != nil {
			assert.Fail("Error marshalling pipeline output", err)
			return
		}

		fmt.Println(string(s)) //nolint:forbidigo // test
	}

	assert.Equal("iteration: 0 - oasis", pex.StepStatus["echo.repeat"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("iteration: 1 - oasis", pex.StepStatus["echo.repeat"]["0"].StepExecutions[1].Output.Data["text"])
	assert.Equal("iteration: 2 - oasis", pex.StepStatus["echo.repeat"]["0"].StepExecutions[2].Output.Data["text"])
	assert.Equal("iteration: 3 - oasis", pex.StepStatus["echo.repeat"]["0"].StepExecutions[3].Output.Data["text"])

	assert.Equal("iteration: 0 - blur", pex.StepStatus["echo.repeat"]["1"].StepExecutions[0].Output.Data["text"])
	assert.Equal("iteration: 1 - blur", pex.StepStatus["echo.repeat"]["1"].StepExecutions[1].Output.Data["text"])
	assert.Equal("iteration: 2 - blur", pex.StepStatus["echo.repeat"]["1"].StepExecutions[2].Output.Data["text"])
	assert.Equal("iteration: 3 - blur", pex.StepStatus["echo.repeat"]["1"].StepExecutions[3].Output.Data["text"])

	assert.Equal("iteration: 0 - radiohead", pex.StepStatus["echo.repeat"]["2"].StepExecutions[0].Output.Data["text"])
	assert.Equal("iteration: 1 - radiohead", pex.StepStatus["echo.repeat"]["2"].StepExecutions[1].Output.Data["text"])
	assert.Equal("iteration: 2 - radiohead", pex.StepStatus["echo.repeat"]["2"].StepExecutions[2].Output.Data["text"])
	assert.Equal("iteration: 3 - radiohead", pex.StepStatus["echo.repeat"]["2"].StepExecutions[3].Output.Data["text"])
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

func (suite *ModTestSuite) TestCallingPipelineInDependentMod() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

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

	// s, err := prettyjson.Marshal(pex.StepStatus)
	// if err != nil {
	// 	assert.Fail("Error marshalling pipeline output", err)
	// 	return
	// }
	// fmt.Println(string(s)) //nolint:forbidigo // test

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

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.http_post_url_encoded", 500*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating executio¡n", err)
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

func (suite *ModTestSuite) TestErroWithIf() {
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
	assert.Equal("", pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Status)
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
	assert.Equal("", pex.StepStatus["pipeline.http"]["1"].StepExecutions[0].Output.Status)
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

	// TODO: should ignored error bubbles up to the pipeline?
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
	assert.Equal(float64(23), pex.PipelineOutput["number"])
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

	assert.Equal(float64(42), pex.PipelineOutput["val"])
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

	assert.Equal(float64(42), pex.PipelineOutput["val"])
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

	assert.Equal("<redacted>", envMap["AWS_ACCESS_KEY_ID"])
	assert.Equal("<redacted>", envMap["AWS_SECRET_ACCESS_KEY"])
}

func (suite *ModTestSuite) TestCredentialRedaction() {
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

	// This is not redacted because we're looking for either field name or field value, and neither will hit the redaction list
	assert.Equal("<redacted>", pex.PipelineOutput["val"].(map[string]interface{})["AWS_ACCESS_KEY_ID"])
	assert.Equal("<redacted>", pex.PipelineOutput["val"].(map[string]interface{})["AWS_SECRET_ACCESS_KEY"])
	assert.Equal("<redacted>", pex.PipelineOutput["val"].(map[string]interface{})["facebook_access_token"])
	assert.Equal("<redacted>", pex.PipelineOutput["val"].(map[string]interface{})["pattern_match_aws_access_key_id"])

	// not redacted
	assert.Equal("AKFFFAKEFAKEFAKEFAKE", pex.PipelineOutput["val"].(map[string]interface{})["close_but_no_cigar"])
	assert.Equal("two", pex.PipelineOutput["val"].(map[string]interface{})["one"])
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

	assert.Equal("<redacted>", pex.PipelineOutput["slack_token"])

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
	assert.Equal("<redacted>", pex.PipelineOutput["gitlab_token"])

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
	assert.Equal("<redacted>", pex.PipelineOutput["abuseipdb_api_key"])

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
	assert.Equal("<redacted>", pex.PipelineOutput["clickup_token"])
}

func (suite *ModTestSuite) TestBasicCredential() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.cred_basic", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("foo", pex.PipelineOutput["val_username"])
	assert.Equal("bar", pex.PipelineOutput["val_password"])
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

	//
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
	assert.Equal(float64(60), input[schema.AttributeTypeTimeout].(float64))
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
	assert.Equal(float64(0), output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]interface{}); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]interface{})
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 1\n", lines[0].(map[string]interface{})["line"].(string))

	assert.Equal("stdout", lines[1].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 2\n", lines[1].(map[string]interface{})["line"].(string))

	assert.Equal("stdout", lines[2].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 3\n", lines[2].(map[string]interface{})["line"].(string))
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
	assert.Equal(float64(60), input[schema.AttributeTypeTimeout].(float64))
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
	assert.Equal(float64(0), output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]interface{}); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]interface{})
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 1\n", lines[0].(map[string]interface{})["line"].(string))

	assert.Equal("stdout", lines[1].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 2\n", lines[1].(map[string]interface{})["line"].(string))

	assert.Equal("stdout", lines[2].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 3\n", lines[2].(map[string]interface{})["line"].(string))
}

func (suite *ModTestSuite) TestContainerStepWithParamOverride() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"user":    "root",
		"timeout": 120,
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
	assert.Equal(float64(120), input[schema.AttributeTypeTimeout].(float64))
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
	assert.Equal(float64(0), output.Data["exit_code"])

	if _, ok := output.Data["lines"].([]interface{}); !ok {
		assert.Fail("Container ID should be a list of strings")
	}
	lines := output.Data["lines"].([]interface{})
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 1\n", lines[0].(map[string]interface{})["line"].(string))

	assert.Equal("stdout", lines[1].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 2\n", lines[1].(map[string]interface{})["line"].(string))

	assert.Equal("stdout", lines[2].(map[string]interface{})["stream"].(string))
	assert.Equal("Line 3\n", lines[2].(map[string]interface{})["line"].(string))
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

func (suite *ModTestSuite) TestBufferTokenTooLarge() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, _, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.big_data", 500*time.Millisecond, pipelineInput)

	// It should fail straight away in the loadProcess
	assert.NotNil(err)
	assert.Contains(err.Error(), "Event log entry too large. Max size is")
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

func TestModTestingSuite(t *testing.T) {
	suite.Run(t, &ModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
