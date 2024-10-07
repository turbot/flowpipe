package estest

import (
	"context"
	"fmt"
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
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/constants"

	"github.com/turbot/pipe-fittings/error_helpers"
)

type ModThreeTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	server                *http.Server
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *ModThreeTestSuite) SetupSuite() {

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

	pipelineDirPath := path.Join(cwd, "test_suite_mod_3")

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
func (suite *ModThreeTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}

	suite.server.Shutdown(suite.ctx) //nolint:errcheck // just a test case
	suite.TearDownSuiteRunCount++
}

func (suite *ModThreeTestSuite) TestExecutionPipelineSimple() {
	assert := assert.New(suite.T())

	name := "test_suite_mod_3.pipeline.simple"

	pipelineCmd := &event.PipelineQueue{
		Name: name,
	}

	executionCmd := &event.ExecutionQueue{
		Event:         event.NewExecutionEvent(),
		PipelineQueue: pipelineCmd,
	}

	if err := suite.esService.Send(executionCmd); err != nil {
		assert.Fail(fmt.Sprintf("error sending pipeline command: %v", err))
		return
	}

	time.Sleep(100 * time.Millisecond)

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 10*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail(fmt.Sprintf("error getting execution: %v", err))
		return
	}

	if ex.Status != "finished" {
		assert.Fail(fmt.Sprintf("execution status is %s", ex.Status))
	}

}

func (suite *ModThreeTestSuite) TestExecutionEventsSimple() {
	assert := assert.New(suite.T())

	name := "test_suite_mod_3.trigger.schedule.s_simple"

	triggerCmd := &event.TriggerQueue{
		Name: name,
	}

	executionCmd := &event.ExecutionQueue{
		Event:        event.NewExecutionEvent(),
		TriggerQueue: triggerCmd,
	}

	if err := suite.esService.Send(executionCmd); err != nil {
		assert.Fail(fmt.Sprintf("error sending pipeline command: %v", err))
		return
	}

	time.Sleep(100 * time.Millisecond)

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 10*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail(fmt.Sprintf("error getting execution: %v", err))
		return
	}

	if ex.Status != "finished" {
		assert.Fail(fmt.Sprintf("execution status is %s", ex.Status))
	}
}

func (suite *ModThreeTestSuite) TestExecutionEventsSimpleErrorIgnored() {
	assert := assert.New(suite.T())

	name := "test_suite_mod_3.trigger.schedule.s_simple_error_ignored_with_if_matches"

	triggerCmd := &event.TriggerQueue{
		Name: name,
	}

	executionCmd := &event.ExecutionQueue{
		Event:        event.NewExecutionEvent(),
		TriggerQueue: triggerCmd,
	}

	if err := suite.esService.Send(executionCmd); err != nil {
		assert.Fail(fmt.Sprintf("error sending pipeline command: %v", err))
		return
	}

	time.Sleep(100 * time.Millisecond)

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 10*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail(fmt.Sprintf("error getting execution: %v", err))
		return
	}

	if ex.Status != "finished" {
		assert.Fail(fmt.Sprintf("execution status is %s", ex.Status))
	}

	assert.Equal(1, len(ex.PipelineExecutions), "Expected 1 pipeline execution")
	for _, pex := range ex.PipelineExecutions {
		assert.Equal("finished", pex.Status)
		assert.Equal("should be calculated", pex.PipelineOutput["val"])
	}
}

func (suite *ModThreeTestSuite) TestExecutionEventsSimpleFailure() {
	assert := assert.New(suite.T())

	name := "test_suite_mod_3.trigger.schedule.s_simple_failure"

	triggerCmd := &event.TriggerQueue{
		Name: name,
	}

	executionCmd := &event.ExecutionQueue{
		Event:        event.NewExecutionEvent(),
		TriggerQueue: triggerCmd,
	}

	if err := suite.esService.Send(executionCmd); err != nil {
		assert.Fail(fmt.Sprintf("error sending pipeline command: %v", err))
		return
	}

	time.Sleep(100 * time.Millisecond)

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 10*time.Millisecond, 50, "failed")
	if err != nil {
		assert.Fail(fmt.Sprintf("error getting execution: %v", err))
		return
	}

	if ex.Status != "failed" {
		assert.Fail(fmt.Sprintf("execution status is %s", ex.Status))
	}
}

func (suite *ModThreeTestSuite) TestExecutionQueryTrigger() {
	assert := assert.New(suite.T())

	name := "test_suite_mod_3.trigger.query.simple_sqlite"

	triggerCmd := &event.TriggerQueue{
		Name: name,
	}

	executionCmd := &event.ExecutionQueue{
		Event:        event.NewExecutionEvent(),
		TriggerQueue: triggerCmd,
	}

	if err := suite.esService.Send(executionCmd); err != nil {
		assert.Fail(fmt.Sprintf("error sending pipeline command: %v", err))
		return
	}

	time.Sleep(100 * time.Millisecond)

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 10*time.Millisecond, 50, "finished")
	if err != nil {
		assert.Fail(fmt.Sprintf("error getting execution: %v", err))
		return
	}

	if ex.Status != "finished" {
		assert.Fail(fmt.Sprintf("execution status is %s", ex.Status))
	}

	// There should only be 1 root pipeline
	assert.Equal(1, len(ex.RootPipelines), "Expected 1 root pipeline")
	// And there should only be 1 pipeline execution
	assert.Equal(1, len(ex.PipelineExecutions), "Expected 1 pipeline execution")

	pex := ex.PipelineExecutions[ex.RootPipelines[0]]
	if pex == nil {
		assert.Fail("root pipeline execution")
		return
	}

	assert.Equal(6, len(pex.PipelineOutput["inserted_rows"].([]any)), "Expected 6 inserted rows")

}

func TestModThreeTestingSuite(t *testing.T) {
	suite.Run(t, &ModThreeTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
