package estest

// Basic imports
import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/cache"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	putils "github.com/turbot/pipe-fittings/utils"
)

type DefaultModTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *DefaultModTestSuite) SetupSuite() {

	err := os.Setenv("RUN_MODE", "TEST_ES")
	if err != nil {
		panic(err)
	}
	err = os.Setenv("FP_VAR_var_from_env", "from env")
	if err != nil {
		panic(err)
	}

	// sets app specific constants defined in pipe-fittings
	viper.SetDefault("main.version", "0.0.0-test.0")
	viper.SetDefault(constants.ArgProcessRetention, 604800) // 7 days
	localcmdconfig.SetAppSpecificConstants()

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pipelineDirPath := path.Join(cwd, "default_mod")

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
func (suite *DefaultModTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}

	suite.TearDownSuiteRunCount++
}

func (suite *DefaultModTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *DefaultModTestSuite) AfterTest(suiteName, testName string) {
}

func (suite *DefaultModTestSuite) TestEchoOne() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.echo_one", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors))
	assert.Equal("Hello World from Depend A", pex.PipelineOutput["echo_one_output"])
	assert.Equal(1, len(pex.PipelineOutput))
}

func (suite *DefaultModTestSuite) TestLoopWithFunction() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.simple_pipeline_loop_with_args", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors))
	assert.Equal(1, len(pex.PipelineOutput))

	outputValues := pex.PipelineOutput["value"].(map[string]interface{})
	assert.Equal(4, len(outputValues))
	// TODO: test more here
}

func (suite *DefaultModTestSuite) TestNestedWithInvalidParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.parent_invalid_param", 100*time.Millisecond, pipelineInput)

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

	assert.Equal(1, len(pex.Errors))
	assert.Equal("unknown parameter specified 'credentials'", pex.Errors[0].Error.Detail)
}

func (suite *DefaultModTestSuite) TestCredInStepOutput() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.cred_in_step_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pex.PipelineOutput["val"])
}

func (suite *DefaultModTestSuite) TestCredInOutput() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.cred_in_output", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pex.PipelineOutput["val"])
}

func (suite *DefaultModTestSuite) TestDynamicCredResolution() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.dynamic_cred", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pex.PipelineOutput["val"])
}

func (suite *DefaultModTestSuite) TestDynamicCredResolutionNested() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.dynamic_cred_parent", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	assert.Equal(0, len(pex.Errors))
	assert.Equal("sso_key", pex.PipelineOutput["val_0"])
	assert.Equal("dundermifflin_key", pex.PipelineOutput["val_1"])
}

func (suite *DefaultModTestSuite) TestInputStepWithDefaultNotifier() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.integration_pipe_default_with_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}
	waitTime := 100 * time.Millisecond

	_, _, stepExecution, err := getPipelineExWaitForStepStarted(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, waitTime, 40, "input.my_step")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.NotNil(stepExecution)
	assert.Equal("starting", stepExecution.Status)
	assert.True(strings.HasPrefix(stepExecution.Input["form_url"].(string), "http://localhost:7103/form"), "form_url should start with http://localhost:7103/form but "+stepExecution.Input["form_url"].(string))
}

func (suite *DefaultModTestSuite) TestInputStepOptionResolution() {
	suite.testInputStepOptionResolution("default_mod.pipeline.input_opts_att_resolution")
	suite.testInputStepOptionResolution("default_mod.pipeline.input_opt_block_resolution")
}

func (suite *DefaultModTestSuite) testInputStepOptionResolution(pipelineName string) {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	waitTime := 100 * time.Millisecond
	stepName := "input.input_test"

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, pipelineName, waitTime, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, sex, err := getPipelineExWaitForStepStarted(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, waitTime, 40, "input.input_test")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("started", pex.Status)
	assert.Equal(3, len(pex.StepExecutions))
	assert.Equal(stepName, sex.Name)

	if opts, ok := sex.Input["options"].([]any); ok {
		assert.Equal(2, len(opts))
		if opt0, ok := opts[0].(map[string]any); ok {
			assert.Equal("yes", opt0["value"].(string))
		} else {
			assert.Fail("Error parsing first option to map[string]any")
			return
		}
		if opt1, ok := opts[1].(map[string]any); ok {
			assert.Equal("no", opt1["value"].(string))
		} else {
			assert.Fail("Error parsing second option to map[string]any")
			return
		}
	} else {
		assert.Fail("Error obtaining options from step input")
		return
	}
}

func (suite *DefaultModTestSuite) TestSleepStepReferenceToFlowpipeMetadata() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// has reference to the built in flowpipe attribute
	//
	// step "transform" "check_start" {
	//     value = step.sleep.sleep.flowpipe.started_at
	// }

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.sleep_with_flowpipe_attributes", 1*time.Second, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	start, err := time.Parse(putils.RFC3339WithMS, pex.PipelineOutput["val_start"].(string))
	if err != nil {
		assert.Fail("Error parsing start time", err)
		return
	}

	end, err := time.Parse(putils.RFC3339WithMS, pex.PipelineOutput["val_end"].(string))
	if err != nil {
		assert.Fail("Error parsing end time", err)
		return
	}

	// make sure that end is after start
	assert.True(end.After(start))

	// make sure that end is at least 1 second after start
	assert.True(end.Sub(start) > 800*time.Millisecond)
}

func (suite *DefaultModTestSuite) TestSleepStepReferenceToFlowpipeMetadataInPipelineStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.parent_of_nested", 1*time.Second, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)

	start, err := time.Parse(putils.RFC3339WithMS, pex.PipelineOutput["val_start"].(string))
	if err != nil {
		assert.Fail("Error parsing start time", err)
		return
	}

	end, err := time.Parse(putils.RFC3339WithMS, pex.PipelineOutput["val_end"].(string))
	if err != nil {
		assert.Fail("Error parsing end time", err)
		return
	}

	// make sure that end is after start
	assert.True(end.After(start))

	// make sure that end is at least 1 second after start
	assert.True(end.Sub(start) > 900*time.Millisecond)
}

func TestDefaultModTestingSuite(t *testing.T) {
	suite.Run(t, &DefaultModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
