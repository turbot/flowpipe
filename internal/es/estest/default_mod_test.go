package estest

// Basic imports
import (
	"context"
	"net/http"
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
	fconstants "github.com/turbot/flowpipe/internal/constants"
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

	server                *http.Server
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

	suite.server = StartServer()

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

	suite.server.Shutdown(suite.ctx) //nolint:errcheck // just a test case
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

func (suite *DefaultModTestSuite) TestEchoOneCustomEventStoreLocation() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// make sure that ./event-store-test-dir/flowpipe.db does not exist
	_, err := os.Stat("./event-store-test-dir/flowpipe.db")
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.Remove("./event-store-test-dir/flowpipe.db")
		if err != nil {
			assert.FailNow("Error removing event store file", err)
		}
	}

	viper.SetDefault(constants.ArgDataDir, "./event-store-test-dir")

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

	// check if the event store file was created
	fi, err := os.Stat("./event-store-test-dir/flowpipe.db")
	assert.Nil(err)

	assert.Equal("flowpipe.db", fi.Name())
	assert.False(fi.IsDir())

	// now delete the event store file
	err = os.Remove("./event-store-test-dir/flowpipe.db")
	assert.Nil(err)

	// removed the default value
	viper.SetDefault(constants.ArgDataDir, "")
}

func (suite *DefaultModTestSuite) TestBasicAuth() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"user_email": "asdf",
		"token":      "12345",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.test_basic_auth", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("401 Unauthorized", pex.Errors[0].Error.Detail)

	// Now re-run with the correct credentials
	pipelineInput = modconfig.Input{
		"user_email": "testuser",
		"token":      "testpass",
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.test_basic_auth", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("finished", pex.Status)
	assert.Equal("Authenticated successfully", pex.PipelineOutput["val"])

	// re-run with bad creds, should fail again
	pipelineInput = modconfig.Input{
		"user_email": "testuser",
		"token":      "testpassxxxxx",
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.test_basic_auth", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
	assert.Equal("401 Unauthorized", pex.Errors[0].Error.Detail)
}

func (suite *DefaultModTestSuite) TestLoopWithFunction() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.simple_pipeline_loop_with_args_and_function", 100*time.Millisecond, pipelineInput)

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

func (suite *DefaultModTestSuite) TestNestedWithCreds() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.parent_with_creds", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("AAAA", pex.PipelineOutput["env"].(map[string]interface{})["AWS_ACCESS_KEY_ID"])
	assert.Equal("BBBB", pex.PipelineOutput["env"].(map[string]interface{})["AWS_SECRET_ACCESS_KEY"])
}

func (suite *DefaultModTestSuite) TestNestedModWithCreds() {
	assert := assert.New(suite.T())

	os.Setenv("GITHUB_TOKEN", "12345")

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.parent_call_nested_mod_with_cred", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("12345", pex.PipelineOutput["val"])
	assert.Equal("default", pex.PipelineOutput["val_merge"].(map[string]interface{})["cred_name"])
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

// Testing the "better" error message where Flowpipe tries to guess the error and present an improved error message with context.
func (suite *DefaultModTestSuite) TestNestedWithInvalidCred() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.parent_call_nested_mod_with_cred_with_invalid_cred", 100*time.Millisecond, pipelineInput)

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
	assert.Contains(pex.Errors[0].Error.Detail, "Missing credential: This object does not have an attribute named \"github\"")
}

func (suite *DefaultModTestSuite) TestNestedWithInvalidCredIncorrectErrorMessage() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.incorrect_better_error_message_from_id_attribute", 100*time.Millisecond, pipelineInput)

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
	assert.Contains(pex.Errors[0].Error.Detail, "Unsupported attribute: This object does not have an attribute named \"id\".")
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

func (suite *DefaultModTestSuite) TestStalledPipelineWithIf() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	// default_mod.pipeline.caller was failing due to issue https://github.com/turbot/flowpipe/issues/836
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.caller", 100*time.Millisecond, pipelineInput)

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
	// assert.Equal("starting", stepExecution.Status)
	assert.True(strings.HasPrefix(stepExecution.Input[fconstants.FormUrl].(string), "http://localhost:7103/form"), "form_url should start with http://localhost:7103/form but "+stepExecution.Input["form_url"].(string))
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

func (suite *DefaultModTestSuite) TestInputStepError() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.input_step_error_out", 1*time.Second, pipelineInput)
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
	assert.Equal("Internal Error: all 1 notifications failed:\nslack server error: 777 status code 777\n", pex.Errors[0].Error.Detail)
	assert.Equal(500, pex.Errors[0].Error.Status)
}

func (suite *DefaultModTestSuite) TestInputStepErrorIgnored() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.input_step_error_out_error_config", 1*time.Second, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// Pipeline finished
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(1, len(pex.StepStatus["input.test"]["0"].StepExecutions))

	// but the step execution actually failed, but the error was ignored
	assert.Equal("failed", pex.StepStatus["input.test"]["0"].StepExecutions[0].Status)
}

func (suite *DefaultModTestSuite) TestInputStepErrorRetried() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.input_step_error_out_retry", 1*time.Second, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// Pipeline failed
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))

	// retry max attempts = 3
	assert.Equal(3, len(pex.StepStatus["input.test"]["0"].StepExecutions))

	// but the step execution actually failed, but the error was ignored
	assert.Equal("failed", pex.StepStatus["input.test"]["0"].StepExecutions[0].Status)
}

func (suite *DefaultModTestSuite) TestIfLoop() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "default_mod.pipeline.if_loop", 1*time.Second, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	// Pipeline failed
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))

	// retry max attempts = 3
	assert.Equal(1, len(pex.StepStatus["message.test"]))
}

func TestDefaultModTestingSuite(t *testing.T) {
	suite.Run(t, &DefaultModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
