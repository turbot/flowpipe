package es_test

// Basic imports
import (
	"context"
	"fmt"
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

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	_, pipelineCmd, err := suite.runPipeline("expr_depend_and_function", 2300*time.Millisecond)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	ex, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 3, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	pipelineDefn, err := ex.PipelineDefinition(pipelineCmd.PipelineExecutionID)
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

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < 3 && !pex.IsComplete(); i++ {
		time.Sleep(100 * time.Millisecond)

		err = ex.LoadProcess(pipelineCmd.Event)
		if err != nil {
			assert.Fail("Error loading process", err)
			return
		}
		pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]
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

	assert.Equal(8, len(echoStepsOutput))
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

	assert.Equal(2, len(echoStepsOutput["echo_for_if"].([]*types.StepOutput)))
	// First one is OK, the second step should be skipped
	assert.Equal("finished", echoStepsOutput["echo_for_if"].([]*types.StepOutput)[0].Status)
	assert.Equal("skipped", echoStepsOutput["echo_for_if"].([]*types.StepOutput)[1].Status)

	assert.Equal(3, len(pex.PipelineOutput))
	assert.Equal("sleep 1 output: 1s", pex.PipelineOutput["one"])
	assert.Equal("Sleep 1 Output: 1s", pex.PipelineOutput["one_function"])
	assert.Equal("2s", pex.PipelineOutput["indexed"])
}

func (suite *EsTestSuite) TestIfConditionsOnSteps() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := suite.runPipeline("if", 100*time.Millisecond)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	ex, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

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

func (suite *EsTestSuite) TestErrorHandlingOnPipelines() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := suite.runPipeline("bad_http_not_ignored", 100*time.Millisecond)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, err := suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "failed")
	if err == nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Pipeline should not have completed", err)
		return
	}

	assert.False(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", ex.AllStepOutputs["http"]["my_step_1"].(*types.StepOutput).Status)
	assert.NotNil(ex.AllStepOutputs["http"]["my_step_1"].(*types.StepOutput).Errors)
	assert.Equal(float64(404), ex.AllStepOutputs["http"]["my_step_1"].(*types.StepOutput).OutputVariables["status_code"])
	assert.Nil(ex.AllStepOutputs["echo"]["bad_http"])

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_ignored pipeline
	_, cmd, err = suite.runPipeline("bad_http_ignored", 100*time.Millisecond)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)

	output := pex.PipelineOutput["one"]
	if output == nil {
		assert.Fail("Pipeline output not found")
		return
	}

	assert.Equal("foo", output.(string))

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_ignored_get_error_code pipeline
	_, cmd, err = suite.runPipeline("bad_http_ignored_get_error_code", 100*time.Millisecond)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, err = suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)

	output = pex.PipelineOutput["one"]
	if output == nil {
		assert.Fail("Pipeline output not found")
		return
	}

	assert.Equal(float64(404), ex.AllStepOutputs["http"]["my_step_1"].(*types.StepOutput).OutputVariables["status_code"])
	assert.Equal("404", ex.AllStepOutputs["echo"]["bad_http"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal("404", output.(string))

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_with_for pipeline
	_, cmd, err = suite.runPipeline("bad_http_with_for", 1*time.Second)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, err = suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)

	assert.Equal(float64(404), ex.AllStepOutputs["http"]["http_step"].([]*types.StepOutput)[0].OutputVariables["status_code"])
	assert.Equal(float64(404), ex.AllStepOutputs["http"]["http_step"].([]*types.StepOutput)[1].OutputVariables["status_code"])
	assert.Equal(float64(200), ex.AllStepOutputs["http"]["http_step"].([]*types.StepOutput)[2].OutputVariables["status_code"])

	assert.Equal("skipped", ex.AllStepOutputs["echo"]["http_step"].([]*types.StepOutput)[0].Status)
	assert.Equal("skipped", ex.AllStepOutputs["echo"]["http_step"].([]*types.StepOutput)[1].Status)
	assert.Equal("finished", ex.AllStepOutputs["echo"]["http_step"].([]*types.StepOutput)[2].Status)
	assert.Nil(ex.AllStepOutputs["echo"]["http_step"].([]*types.StepOutput)[0].OutputVariables["text"])
	assert.Nil(ex.AllStepOutputs["echo"]["http_step"].([]*types.StepOutput)[1].OutputVariables["text"])
	assert.Equal("200", ex.AllStepOutputs["echo"]["http_step"].([]*types.StepOutput)[2].OutputVariables["text"])

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test
}

func (suite *EsTestSuite) getPipelineExAndWait(event *event.Event, pipelineExecutionID string, waitTime time.Duration, waitRetry int, expectedState string) (*execution.Execution, *execution.PipelineExecution, error) {
	// check if the execution id has been completed, check 3 times
	ex, err := execution.NewExecution(suite.ctx)
	if err != nil {
		return nil, nil, err
	}

	err = ex.LoadProcess(event)
	if err != nil {
		return nil, nil, err
	}

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		return nil, nil, fmt.Errorf("Pipeline execution " + pipelineExecutionID + " not found")
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < waitRetry && !pex.IsComplete() && pex.Status != expectedState; i++ {
		time.Sleep(waitTime)

		err = ex.LoadProcess(event)
		if err != nil {
			return nil, nil, fmt.Errorf("Error loading process: %w", err)
		}
		if pex == nil {
			return nil, nil, fmt.Errorf("Pipeline execution " + pipelineExecutionID + " not found")
		}
		pex = ex.PipelineExecutions[pipelineExecutionID]
	}

	if !pex.IsComplete() {
		return ex, pex, fmt.Errorf("not completed")
	}

	return ex, pex, nil

}
func (suite *EsTestSuite) runPipeline(name string, initialWaitTime time.Duration) (*execution.Execution, *event.PipelineQueue, error) {

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(suite.ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                name,
	}

	if err := suite.esService.Send(pipelineCmd); err != nil {
		return nil, nil, fmt.Errorf("Error sending pipeline command: %w", err)

	}

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	time.Sleep(initialWaitTime)

	// check if the execution id has been completed, check 3 times
	ex, err := execution.NewExecution(suite.ctx)
	if err != nil {
		return nil, nil, err
	}

	err = ex.LoadProcess(pipelineCmd.Event)
	if err != nil {
		return nil, nil, err
	}

	return ex, pipelineCmd, nil
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEsTestingSuite(t *testing.T) {
	suite.Run(t, new(EsTestSuite))
}
