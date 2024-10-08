package estest

// Basic imports
import (
	"context"
	"net/http"
	"os"
	"path"
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
)

type ModLongRunningTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	server                *http.Server
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *ModLongRunningTestSuite) SetupSuite() {

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
func (suite *ModLongRunningTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}

	suite.server.Shutdown(suite.ctx) //nolint:errcheck // just a test case
	suite.TearDownSuiteRunCount++
}

func (suite *ModLongRunningTestSuite) BeforeTest(suiteName, testName string) {

}
func (suite *ModLongRunningTestSuite) AfterTest(suiteName, testName string) {
}

func (suite *ModLongRunningTestSuite) TestSleepWithLoop() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_sleep", 5*time.Second, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 5*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["sleep.sleep"]["0"].StepExecutions))

	// Testing this loop config:
	// loop {
	//   until = loop.index > 2
	//   duration = "${loop.index}s"
	// }
	assert.Equal("0s", pex.StepStatus["sleep.sleep"]["0"].StepExecutions[1].Input["duration"])
	assert.Equal("1s", pex.StepStatus["sleep.sleep"]["0"].StepExecutions[2].Input["duration"])
	assert.Equal("2s", pex.StepStatus["sleep.sleep"]["0"].StepExecutions[3].Input["duration"])
}

func (suite *ModLongRunningTestSuite) TestHttpWithLoop() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_http", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 5*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["http.http"]["0"].StepExecutions))

	assert.Equal("initial - 0", pex.StepStatus["http.http"]["0"].StepExecutions[1].Input["request_body"])
	assert.Equal("initial - 0 - 1", pex.StepStatus["http.http"]["0"].StepExecutions[2].Input["request_body"])
	assert.Equal("initial - 0 - 1 - 2", pex.StepStatus["http.http"]["0"].StepExecutions[3].Input["request_body"])
}

func (suite *ModLongRunningTestSuite) TestTransformWithLoop() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_transform", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 5*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["transform.transform"]["0"].StepExecutions))

	assert.Equal("initial value - 0", pex.StepStatus["transform.transform"]["0"].StepExecutions[1].Input["value"])
	assert.Equal("initial value - 0 - 1", pex.StepStatus["transform.transform"]["0"].StepExecutions[2].Input["value"])
	assert.Equal("initial value - 0 - 1 - 2", pex.StepStatus["transform.transform"]["0"].StepExecutions[3].Input["value"])
}

func (suite *ModLongRunningTestSuite) TestTransformWithLoopMap() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_transform_map", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 5*time.Millisecond, 40, "finished")
	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["transform.transform"]["0"].StepExecutions))

	assert.Equal(map[string]interface{}{
		"name": "victor - 0 - 1",
		"age":  float64(31),
	}, pex.StepStatus["transform.transform"]["0"].StepExecutions[2].Input["value"])

	assert.Equal(map[string]interface{}{
		"name": "victor - 0 - 1 - 2",
		"age":  float64(33),
	}, pex.StepStatus["transform.transform"]["0"].StepExecutions[3].Input["value"])
}

func (suite *ModLongRunningTestSuite) TestContainerWithLoop() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_container", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error waiting for execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["container.container"]["0"].StepExecutions))

	assert.Equal("bar - 0", pex.StepStatus["container.container"]["0"].StepExecutions[1].Output.Data["stdout"])
	assert.Equal("bar - 1", pex.StepStatus["container.container"]["0"].StepExecutions[2].Output.Data["stdout"])
	assert.Equal("bar - 2", pex.StepStatus["container.container"]["0"].StepExecutions[3].Output.Data["stdout"])
}

func (suite *ModLongRunningTestSuite) TestMessageWithLoop() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_message", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 10*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error waiting for execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["message.message"]["0"].StepExecutions))

	assert.Equal("2", pex.PipelineOutput["val"].(map[string]interface{})["3"].(map[string]interface{})["text"])

	// TODO: this won't work until we fix https://github.com/turbot/flowpipe/issues/812 to help testing
	// assert.Equal("foo - 0", pex.StepStatus["message.message"]["0"].StepExecutions[1].Output.Data["text"])
	// assert.Equal("foo - 0 - 1", pex.StepStatus["message.message"]["0"].StepExecutions[2].Output.Data["text"])
	// assert.Equal("foo - 0 - 1 - 2", pex.StepStatus["message.message"]["0"].StepExecutions[3].Output.Data["text"])
}

func (suite *ModLongRunningTestSuite) TestMessageWithLoop2() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_message_2", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 10*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail("Error waiting for execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
	assert.Equal(0, len(pex.Errors))
	assert.Equal(4, len(pex.StepStatus["message.message"]["0"].StepExecutions))

	assert.Equal("2", pex.PipelineOutput["val"].(map[string]interface{})["3"].(map[string]interface{})["text"])
}

func (suite *ModLongRunningTestSuite) TestMessageWithLoopFailed() {
	assert := assert.New(suite.T())
	pipelineInput := modconfig.Input{}
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod.pipeline.loop_message_failed", 40*time.Millisecond, pipelineInput)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 5*time.Millisecond, 40, "failed")
	assert.Equal("failed", pex.Status)
	assert.Equal(1, len(pex.Errors))
}

func TestModLongRunningTestingSuite(t *testing.T) {
	suite.Run(t, &ModLongRunningTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
