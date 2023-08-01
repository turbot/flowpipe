package es_test

// Basic imports
import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EsTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	esService             *es.ESService
	ctx                   context.Context
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *EsTestSuite) SetupSuite() {
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

	pipelineDirPath := path.Join(cwd, "pipelines")

	viper.GetViper().Set("pipeline.dir", pipelineDirPath)
	viper.GetViper().Set("output.dir", outputPath)
	viper.GetViper().Set("log.dir", outputPath)

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

	esService, err := es.NewESService(ctx)
	if err != nil {
		panic(err)
	}
	err = esService.Start()
	if err != nil {
		panic(err)
	}
	esService.Status = "running"
	esService.StartedAt = util.TimeNow()

	suite.esService = esService

	// Give some time for Watermill to fully start
	time.Sleep(2 * time.Second)

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *EsTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}
	suite.TearDownSuiteRunCount++
}

func (suite *EsTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *EsTestSuite) AfterTest(suiteName, testName string) {

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *EsTestSuite) TestExpressionWithDependenciesFunctions() {
	assert := assert.New(suite.T())

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(suite.ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                "expr_depend_and_function",
	}

	if err := suite.esService.Send(pipelineCmd); err != nil {
		assert.Fail("Error sending pipeline command", err)
		return
	}

	pipelineExecutionID := pipelineCmd.PipelineExecutionID

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	time.Sleep(2200 * time.Millisecond)

	// check if the execution id has been completed, check 3 times
	ex, err := execution.NewExecution(suite.ctx)
	for i := 0; i < 3 && err != nil; i++ {
		time.Sleep(100 * time.Millisecond)
		ex, err = execution.NewExecution(suite.ctx)
	}

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	err = ex.LoadProcess(pipelineCmd.Event)
	if err != nil {
		assert.Fail("Error loading process", err)
		return
	}

	pipelineDefn, err := ex.PipelineDefinition(pipelineExecutionID)
	if err != nil || pipelineDefn == nil {
		assert.Fail("Pipeline definition not found", err)
	}

	explicitDependsStep := pipelineDefn.GetStep("echo.explicit_depends")
	if explicitDependsStep == nil {
		assert.Fail("echo.explicit_depends not found")
		return
	}

	dependsOn := explicitDependsStep.GetDependsOn()

	assert.Equal(2, len(dependsOn))
	assert.Contains(dependsOn, "echo.text_1")
	assert.Contains(dependsOn, "echo.text_2")

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		assert.Fail("Pipeline execution not found")
		return
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < 3 && !pex.IsComplete(); i++ {
		time.Sleep(100 * time.Millisecond)

		err = ex.LoadProcess(pipelineCmd.Event)
		if err != nil {
			assert.Fail("Error loading process", err)
			return
		}
		pex = ex.PipelineExecutions[pipelineExecutionID]
	}

	if !pex.IsComplete() {
		assert.Fail("Pipeline execution not complete")
		return
	}

	echoStepsOutput := ex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal(7, len(echoStepsOutput))
	assert.Equal("foo bar", echoStepsOutput["text_1"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal("lower case Bar Foo Bar Baz and here", echoStepsOutput["text_2"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal("output 2 Lower Case Bar Foo Bar Baz And Here title(output1) Foo Bar", echoStepsOutput["text_3"].(*types.StepOutput).OutputVariables["text"])

	// check output for the "time"/"for"/"sleep" steps
	assert.Equal("sleep 2 output: 2s", echoStepsOutput["echo_sleep_1"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal("sleep 1 output: 1s", echoStepsOutput["echo_sleep_2"].(*types.StepOutput).OutputVariables["text"])

	sleepStepsOutput := ex.AllStepOutputs["sleep"]
	if sleepStepsOutput == nil {
		assert.Fail("sleep step output not found")
		return
	}

	assert.Equal(1, len(sleepStepsOutput))
	sleep1StepOutputs := sleepStepsOutput["sleep_1"].([]*types.StepOutput)
	if sleep1StepOutputs == nil {
		assert.Fail("sleep_1 step output not found")
		return
	}

	assert.Equal(2, len(sleep1StepOutputs))
	assert.Equal("1s", sleep1StepOutputs[0].OutputVariables["duration"])
	assert.Equal("2s", sleep1StepOutputs[1].OutputVariables["duration"])

}

func (suite *EsTestSuite) TestIfConditionsOnSteps() {
	assert := assert.New(suite.T())

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(suite.ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                "if",
	}

	if err := suite.esService.Send(pipelineCmd); err != nil {
		assert.Fail("Error sending pipeline command", err)
		return
	}

	pipelineExecutionID := pipelineCmd.PipelineExecutionID

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	time.Sleep(100 * time.Millisecond)

	// check if the execution id has been completed, check 3 times
	ex, err := execution.NewExecution(suite.ctx)
	for i := 0; i < 3 && err != nil; i++ {
		time.Sleep(100 * time.Millisecond)
		ex, err = execution.NewExecution(suite.ctx)
	}

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	err = ex.LoadProcess(pipelineCmd.Event)
	if err != nil {
		assert.Fail("Error loading process", err)
		return
	}

	pipelineDefn, err := ex.PipelineDefinition(pipelineExecutionID)
	if err != nil || pipelineDefn == nil {
		assert.Fail("Pipeline definition not found", err)
	}

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		assert.Fail("Pipeline execution not found")
		return
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < 3 && !pex.IsComplete(); i++ {
		time.Sleep(100 * time.Millisecond)

		err = ex.LoadProcess(pipelineCmd.Event)
		if err != nil {
			assert.Fail("Error loading process", err)
			return
		}
		pex = ex.PipelineExecutions[pipelineExecutionID]
	}

	if !pex.IsComplete() {
		assert.Fail("Pipeline execution not complete")
		return
	}

	echoStepsOutput := ex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal(5, len(echoStepsOutput))

	assert.Equal("finished", echoStepsOutput["text_true"].(*types.StepOutput).Status)
	assert.Equal("skipped", echoStepsOutput["text_false"].(*types.StepOutput).Status)
	assert.Equal("finished", echoStepsOutput["text_1"].(*types.StepOutput).Status)
	assert.Equal("finished", echoStepsOutput["text_2"].(*types.StepOutput).Status)
	assert.Equal("skipped", echoStepsOutput["text_3"].(*types.StepOutput).Status)

	assert.Equal("foo", echoStepsOutput["text_true"].(*types.StepOutput).OutputVariables["text"])
	assert.Nil(echoStepsOutput["text_false"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal("foo", echoStepsOutput["text_1"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal("bar", echoStepsOutput["text_2"].(*types.StepOutput).OutputVariables["text"])
	assert.Nil(echoStepsOutput["text_3"].(*types.StepOutput).OutputVariables["text"])

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEsTestingSuite(t *testing.T) {
	suite.Run(t, new(EsTestSuite))
}
