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
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/filepaths"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/utils"
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

func TestModTestingSuite(t *testing.T) {
	suite.Run(t, &ModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
