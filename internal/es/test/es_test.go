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
	"github.com/turbot/flowpipe/internal/service/manager"
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
	time.Sleep(2 * time.Second)
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *EsTestSuite) TestExpressionWithDependenciesFunctions() {
	assert := assert.New(suite.T())

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	_, pipelineCmd, err := suite.runPipeline("expr_depend_and_function", 2300*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	ex, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "finished")
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

	echoStepsOutput := pex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal(10, len(echoStepsOutput))
	assert.Equal("foo bar", echoStepsOutput["text_1"].(*types.Output).Data["text"])
	assert.Equal("lower case Bar Foo Bar Baz and here", echoStepsOutput["text_2"].(*types.Output).Data["text"])
	assert.Equal("output 2 Lower Case Bar Foo Bar Baz And Here title(output1) Foo Bar", echoStepsOutput["text_3"].(*types.Output).Data["text"])

	// check output for the "time"/"for"/"sleep" steps
	assert.Equal("sleep 2 output: 2s", echoStepsOutput["echo_sleep_1"].(*types.Output).Data["text"])
	assert.Equal("sleep 1 output: 1s", echoStepsOutput["echo_sleep_2"].(*types.Output).Data["text"])

	sleepStepsOutput := pex.AllStepOutputs["sleep"]
	if sleepStepsOutput == nil {
		assert.Fail("sleep step output not found")
		return
	}

	assert.Equal(1, len(sleepStepsOutput))
	sleep1StepOutputs := sleepStepsOutput["sleep_1"].([]*types.Output)
	if sleep1StepOutputs == nil {
		assert.Fail("sleep_1 step output not found")
		return
	}

	assert.Equal(2, len(sleep1StepOutputs))
	assert.Equal("1s", sleep1StepOutputs[0].Data["duration"])
	assert.Equal("2s", sleep1StepOutputs[1].Data["duration"])

	assert.Equal(2, len(echoStepsOutput["echo_for_if"].([]*types.Output)))
	// First one is OK, the second step should be skipped
	assert.Equal("finished", echoStepsOutput["echo_for_if"].([]*types.Output)[0].Status)
	assert.Equal("skipped", echoStepsOutput["echo_for_if"].([]*types.Output)[1].Status)

	assert.Equal(3, len(pex.PipelineOutput))
	assert.Equal("sleep 1 output: 1s", pex.PipelineOutput["one"])
	assert.Equal("Sleep 1 Output: 1s", pex.PipelineOutput["one_function"])
	assert.Equal("2s", pex.PipelineOutput["indexed"])

	// checking the "echo.literal_for" step
	assert.Equal(3, len(echoStepsOutput["literal_for"].([]*types.Output)))

	assert.Equal("name is bach", echoStepsOutput["literal_for"].([]*types.Output)[0].Data["text"])
	assert.Equal("name is beethoven", echoStepsOutput["literal_for"].([]*types.Output)[1].Data["text"])
	assert.Equal("name is mozart", echoStepsOutput["literal_for"].([]*types.Output)[2].Data["text"])

	// checking the "echo.literal_for_from_list" step
	assert.Equal(3, len(echoStepsOutput["literal_for_from_list"].([]*types.Output)))

	// TODO: "something" is re-ordering the for_each expression evaluation to an ordered list, I'm yet to find out what that is
	assert.Equal("prokofiev", echoStepsOutput["literal_for_from_list"].([]*types.Output)[0].Data["text"])
	assert.Equal("rachmaninoff", echoStepsOutput["literal_for_from_list"].([]*types.Output)[1].Data["text"])
	assert.Equal("shostakovitch", echoStepsOutput["literal_for_from_list"].([]*types.Output)[2].Data["text"])

}

func (suite *EsTestSuite) TestIfConditionsOnSteps() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := suite.runPipeline("if", 100*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 200, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	echoStepsOutput := pex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal(5, len(echoStepsOutput))

	assert.Equal("finished", echoStepsOutput["text_true"].(*types.Output).Status)
	assert.Equal("skipped", echoStepsOutput["text_false"].(*types.Output).Status)
	assert.Equal("finished", echoStepsOutput["text_1"].(*types.Output).Status)
	assert.Equal("finished", echoStepsOutput["text_2"].(*types.Output).Status)
	assert.Equal("skipped", echoStepsOutput["text_3"].(*types.Output).Status)

	assert.Equal("foo", echoStepsOutput["text_true"].(*types.Output).Data["text"])
	assert.Nil(echoStepsOutput["text_false"].(*types.Output).Data["text"])
	assert.Equal("foo", echoStepsOutput["text_1"].(*types.Output).Data["text"])
	assert.Equal("bar", echoStepsOutput["text_2"].(*types.Output).Data["text"])
	assert.Nil(echoStepsOutput["text_3"].(*types.Output).Data["text"])

}

func (suite *EsTestSuite) TestPipelineErrorBubbleUp() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := suite.runPipeline("bad_http_one_step", 200*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 200, "failed")
	if err != nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Status)
	assert.NotNil(pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Errors)
	assert.Equal(float64(404), pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Data["status_code"])
	assert.Nil(pex.AllStepOutputs["echo"]["bad_http"])

	assert.NotNil(pex.PipelineOutput["errors"])
	assert.Equal(float64(404), pex.PipelineOutput["errors"].([]interface{})[0].(map[string]interface{})["error_code"])
}

func (suite *EsTestSuite) TestParentChildPipeline() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := suite.runPipeline("parent_pipeline_with_args", 100*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 300, "finished")
	if err != nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)
	// TODO: this doesn't work yet, we need pass the pipeline status up? or does it has its own status?
	// assert.Equal("finished", pex.AllStepOutputs["pipeline"]["child_pipeline_with_args"].(*types.Output).Status)
	assert.Equal("child echo step: from parent 24", pex.AllStepOutputs["pipeline"]["child_pipeline_with_args"].(*types.Output).Data["child_output"])
	assert.Equal("child echo step: from parent 24", pex.PipelineOutput["parent_output"])
	assert.Nil(pex.PipelineOutput["does_not_exist"])

}

func (suite *EsTestSuite) TestErrorHandlingOnPipelines() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := suite.runPipeline("bad_http_not_ignored", 100*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 200, "failed")
	if err == nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	// This pipeline: bad_http_not_ignored should not complete because there's a step that it can't start
	// so in a way it's "not completed" but it has failed, since it will never be able to start that one step
	assert.False(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Status)
	assert.NotNil(pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Errors)
	assert.Equal(float64(404), pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Data["status_code"])
	assert.Nil(pex.AllStepOutputs["echo"]["bad_http"])

	// end pipeline test

	// bad_http_ignored pipeline
	_, cmd, err = suite.runPipeline("bad_http_ignored", 100*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, err := suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 100, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil || pipelineDefn == nil {
		assert.Fail("Pipeline definition not found", err)
	}

	badHttpIfErrorTrueStep := pipelineDefn.GetStep("echo.bad_http_if_error_true")
	if badHttpIfErrorTrueStep == nil {
		assert.Fail("echo.bad_http_if_error_true not found")
		return
	}
	assert.Contains(badHttpIfErrorTrueStep.GetDependsOn(), "http.my_step_1")

	badHttpIfErrorFalseStep := pipelineDefn.GetStep("echo.bad_http_if_error_false")
	if badHttpIfErrorFalseStep == nil {
		assert.Fail("echo.bad_http_if_error_false not found")
		return
	}
	assert.Contains(badHttpIfErrorFalseStep.GetDependsOn(), "http.my_step_1")

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)

	output := pex.PipelineOutput["one"]
	if output == nil {
		assert.Fail("Pipeline output not found")
		return
	}

	assert.Equal("foo", output.(string))

	assert.Equal("bar", pex.AllStepOutputs["echo"]["bad_http_if_error_true"].(*types.Output).Data["text"])

	// checking the is_error function working correctly
	assert.Equal("finished", pex.AllStepOutputs["echo"]["bad_http_if_error_true"].(*types.Output).Status)
	assert.Equal("skipped", pex.AllStepOutputs["echo"]["bad_http_if_error_false"].(*types.Output).Status)

	// checking the error_message function working correctly
	assert.Equal("404 Not Found", pex.AllStepOutputs["echo"]["error_message"].(*types.Output).Data["text"])

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_ignored_get_error_code pipeline
	_, cmd, err = suite.runPipeline("bad_http_ignored_get_error_code", 100*time.Millisecond, nil)

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

	output = pex.PipelineOutput["one"]
	if output == nil {
		assert.Fail("Pipeline output not found")
		return
	}

	assert.Equal(float64(404), pex.AllStepOutputs["http"]["my_step_1"].(*types.Output).Data["status_code"])
	assert.Equal("404", pex.AllStepOutputs["echo"]["bad_http"].(*types.Output).Data["text"])
	assert.Equal("404", output.(string))

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_with_for pipeline
	_, cmd, err = suite.runPipeline("bad_http_with_for", 1*time.Second, nil)

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

	assert.Equal(float64(404), pex.AllStepOutputs["http"]["http_step"].([]*types.Output)[0].Data["status_code"])
	assert.Equal(float64(404), pex.AllStepOutputs["http"]["http_step"].([]*types.Output)[1].Data["status_code"])
	assert.Equal(float64(200), pex.AllStepOutputs["http"]["http_step"].([]*types.Output)[2].Data["status_code"])

	assert.Equal("skipped", pex.AllStepOutputs["echo"]["http_step"].([]*types.Output)[0].Status)
	assert.Equal("skipped", pex.AllStepOutputs["echo"]["http_step"].([]*types.Output)[1].Status)
	assert.Equal("finished", pex.AllStepOutputs["echo"]["http_step"].([]*types.Output)[2].Status)
	assert.Nil(pex.AllStepOutputs["echo"]["http_step"].([]*types.Output)[0].Data["text"])
	assert.Nil(pex.AllStepOutputs["echo"]["http_step"].([]*types.Output)[1].Data["text"])
	assert.Equal("200", pex.AllStepOutputs["echo"]["http_step"].([]*types.Output)[2].Data["text"])

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_email_with_invalid_recipients pipeline
	_, cmd, err = suite.runPipeline("bad_email_with_invalid_recipients", 1*time.Second, nil)
	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.AllStepOutputs["email"]["test_email"].(*types.Output).Status)
	assert.NotNil(pex.AllStepOutputs["email"]["test_email"].(*types.Output).Errors)

	errors := pex.AllStepOutputs["email"]["test_email"].(*types.Output).Errors
	for _, e := range errors {
		assert.Contains(e.Message, "no such host")
	}

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_email_with_param pipeline
	_, cmd, err = suite.runPipeline("bad_email_with_param", 1*time.Second, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.AllStepOutputs["email"]["test_email"].(*types.Output).Status)
	assert.NotNil(pex.AllStepOutputs["email"]["test_email"].(*types.Output).Errors)

	errors = pex.AllStepOutputs["email"]["test_email"].(*types.Output).Errors
	for _, e := range errors {
		assert.Contains(e.Message, "no such host")
	}

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_email_with_expr
	_, cmd, err = suite.runPipeline("bad_email_with_expr", 1*time.Second, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, err = suite.getPipelineExAndWait(cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	pipelineDefn, err = ex.PipelineDefinition(cmd.PipelineExecutionID)
	if err != nil || pipelineDefn == nil {
		assert.Fail("Pipeline definition not found", err)
	}

	// Get the step details
	explicitDependsStep := pipelineDefn.GetStep("email.test_email")
	if explicitDependsStep == nil {
		assert.Fail("echo.explicit_depends not found")
		return
	}

	// Get the depends_on for the step
	dependsOn := explicitDependsStep.GetDependsOn()
	assert.Equal(2, len(dependsOn))
	assert.Contains(dependsOn, "echo.sender_address")
	assert.Contains(dependsOn, "echo.email_body")

	// Check if the depends_on step is finished and has the correct output
	echoStepOutput := pex.AllStepOutputs["echo"]
	if echoStepOutput == nil {
		assert.Fail("echo step output not found")
		return
	}
	assert.Equal("flowpipe@example.com", echoStepOutput["sender_address"].(*types.Output).Data["text"])
	assert.Equal("This is an email body", echoStepOutput["email_body"].(*types.Output).Data["text"])

	// Expected the pipeline to fail
	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.AllStepOutputs["email"]["test_email"].(*types.Output).Status)
	assert.NotNil(pex.AllStepOutputs["email"]["test_email"].(*types.Output).Errors)

	// The email step should fail because of the invalid smtp host
	errors = pex.AllStepOutputs["email"]["test_email"].(*types.Output).Errors
	for _, e := range errors {
		assert.Contains(e.Message, "no such host")
	}

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test
}

func (suite *EsTestSuite) TestHttp() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := suite.runPipeline("jsonplaceholder_expr", 500*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	echoStepsOutput := pex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal("finished", echoStepsOutput["output"].(*types.Output).Status)
	assert.Equal("201", echoStepsOutput["output"].(*types.Output).Data["text"])

	jsonBodyLoopOutputs := echoStepsOutput["body_json_loop"].([]*types.Output)
	assert.Equal(len(jsonBodyLoopOutputs), 4)
	assert.Equal("brian may", jsonBodyLoopOutputs[0].Data["text"])
	assert.Equal("freddie mercury", jsonBodyLoopOutputs[1].Data["text"])
	assert.Equal("roger taylor", jsonBodyLoopOutputs[2].Data["text"])
	assert.Equal("john deacon", jsonBodyLoopOutputs[3].Data["text"])
}

func (suite *EsTestSuite) TestParam() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := suite.runPipeline("param_test", 100*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 50*time.Millisecond, 500, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	echoStepsOutput := pex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal("finished", echoStepsOutput["simple"].(*types.Output).Status)
	assert.Equal("foo", echoStepsOutput["simple"].(*types.Output).Data["text"])

	assert.Equal("finished", echoStepsOutput["map_echo"].(*types.Output).Status)
	assert.Equal("felix", echoStepsOutput["map_echo"].(*types.Output).Data["text"])

	assert.Equal(7, len(echoStepsOutput["for_with_list"].([]*types.Output)))

	assert.Equal("finished", echoStepsOutput["for_with_list"].([]*types.Output)[0].Status)
	assert.Equal("Green Day", echoStepsOutput["for_with_list"].([]*types.Output)[0].Data["text"])

	assert.Equal("finished", echoStepsOutput["for_with_list"].([]*types.Output)[6].Status)
	assert.Equal("The All-American Rejects", echoStepsOutput["for_with_list"].([]*types.Output)[6].Data["text"])

	assert.Equal("finished", echoStepsOutput["map_diff_types_string"].(*types.Output).Status)
	assert.Equal("string", echoStepsOutput["map_diff_types_string"].(*types.Output).Data["text"])

	assert.Equal("finished", echoStepsOutput["map_diff_types_number"].(*types.Output).Status)
	assert.Equal("1", echoStepsOutput["map_diff_types_number"].(*types.Output).Data["text"])

	assert.Equal(3, len(echoStepsOutput["for_each_list_within_map"].([]*types.Output)))
	assert.Equal("a", echoStepsOutput["for_each_list_within_map"].([]*types.Output)[0].Data["text"])
	assert.Equal("b", echoStepsOutput["for_each_list_within_map"].([]*types.Output)[1].Data["text"])
	assert.Equal("c", echoStepsOutput["for_each_list_within_map"].([]*types.Output)[2].Data["text"])
}

func (suite *EsTestSuite) TestParamOverride() {
	assert := assert.New(suite.T())

	pipelineInput := &types.Input{
		"simple": "bar",
	}

	_, pipelineCmd, err := suite.runPipeline("param_override_test", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 10, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	echoStepsOutput := pex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal("finished", echoStepsOutput["simple"].(*types.Output).Status)
	assert.Equal("bar", echoStepsOutput["simple"].(*types.Output).Data["text"])
}

func (suite *EsTestSuite) TestChildPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := &types.Input{
		"simple": "bar",
	}

	_, pipelineCmd, err := suite.runPipeline("parent_pipeline", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := suite.getPipelineExAndWait(pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 300, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	assert.Equal("child echo step", pex.PipelineOutput["parent_output"])

	// TODO: - Check child pipeline status
	// TODO: - Add status on pipeline step
	// TODO: - add multiple childs
	// TODO: - add more levels (not just 1)
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

func (suite *EsTestSuite) runPipeline(name string, initialWaitTime time.Duration, args *types.Input) (*execution.Execution, *event.PipelineQueue, error) {

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(suite.ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                name,
	}

	if args != nil {
		pipelineCmd.Args = *args
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
