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
	"github.com/turbot/pipe-fittings/utils"

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

	// Test suite 3 contains a pipes metadata in one of the connection. This connection is referenced in the variable.
	//
	// Currently (2024-10) all connections in the variable are resolved regardless they are used by the pipeline execution or not. This is
	// different than the connection dependency in the step.

	os.Setenv("FLOWPIPE_PIPES_TOKEN", "foo")

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
	os.Unsetenv("FLOWPIPE_PIPES_TOKEN")

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

	executionCmd := event.NewExecutionQueueForPipeline("", name)

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

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 100*time.Millisecond, 50, "finished")
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

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 100*time.Millisecond, 50, "failed")
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

	sourceDbFilename := "./test_suite_mod_3/query_source_modified.db"
	_, err := os.Stat(sourceDbFilename)
	if !os.IsNotExist(err) {
		err = os.Remove(sourceDbFilename)
		if err != nil {
			assert.Fail("Error removing test db", err)
			return
		}
	}

	// copy the clean db to the modified db
	err = utils.CopyFile("./test_suite_mod_3/query_source_clean.db", sourceDbFilename)
	if err != nil {
		assert.Fail("Error copying test db", err)
		return
	}

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

func (suite *ModThreeTestSuite) TestSteampipeQueryUsingPipesMockServer() {
	assert := assert.New(suite.T())

	os.Setenv("FLOWPIPE_PIPES_TOKEN", "foo")

	// This simple query step uses a var for the database argument
	// that var is connection.steampipe.default which there is an override
	// pipes metadata to call the pipes mock server (localhost:7104)
	//
	// this tests the resolution of the database connection in var with Pipes server
	name := "test_suite_mod_3.pipeline.query_steampipe"

	executionCmd := event.NewExecutionQueueForPipeline("", name)

	if err := suite.esService.Send(executionCmd); err != nil {
		assert.Fail(fmt.Sprintf("error sending pipeline command: %v", err))
		return
	}

	time.Sleep(100 * time.Millisecond)

	ex, err := getExAndWait(suite.FlowpipeTestSuite, executionCmd.Event.ExecutionID, 10*time.Millisecond, 50, "failed")

	os.Unsetenv("FLOWPIPE_PIPES_TOKEN")
	if err != nil {
		assert.Fail(fmt.Sprintf("error getting execution: %v", err))
		return
	}
	if ex.Status != "failed" {
		assert.Fail(fmt.Sprintf("execution status is %s", ex.Status))
	}

	assert.Equal(1, len(ex.PipelineExecutions), "Expected 1 pipeline execution")
	for _, pex := range ex.PipelineExecutions {
		assert.Equal(1, len(pex.Errors))
		// This proves that the pipes mock server was called and resolved for step
		assert.Equal("Internal Error: Error initializing the database: Bad Request: Invalid database connection string: conn_string_from_mock_server", pex.Errors[0].Error.Detail)
		assert.Equal("query.steampipe", pex.Errors[0].Step)
	}
}

func (suite *ModThreeTestSuite) TestExecutionQueryTriggerModDb() {
	assert := assert.New(suite.T())

	sourceDbFilename := "./test_suite_mod_3/query_source_modified.db"
	_, err := os.Stat(sourceDbFilename)
	if !os.IsNotExist(err) {
		err = os.Remove(sourceDbFilename)
		if err != nil {
			assert.Fail("Error removing test db", err)
			return
		}
	}

	// copy the clean db to the modified db
	err = utils.CopyFile("./test_suite_mod_3/query_source_clean.db", sourceDbFilename)
	if err != nil {
		assert.Fail("Error copying test db", err)
		return
	}

	name := "test_suite_mod_3.trigger.query.simple_sqlite_no_db"

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
