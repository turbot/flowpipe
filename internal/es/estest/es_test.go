package estest

// Basic imports
import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EsTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *EsTestSuite) SetupSuite() {

	err := os.Setenv("RUN_MODE", "TEST_ES")
	if err != nil {
		panic(err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)

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

	viper.GetViper().Set(constants.ArgModLocation, pipelineDirPath)
	viper.GetViper().Set(constants.ArgOutputDir, outputPath)
	viper.GetViper().Set(constants.ArgLogDir, outputPath)

	// Create a single, global context for the application
	ctx := fplog.ContextWithLogger(context.Background())

	suite.ctx = ctx

	// We use the cache to store the pipelines
	cache.InMemoryInitialize(nil)

	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx, manager.WithESService()).Start()
	error_helpers.FailOnError(err)

	suite.esService = m.ESService

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

func (suite *EsTestSuite) TestExpressionWithDependenciesFunctions() {
	assert := assert.New(suite.T())

	// give it a moment to let Watermill does its thing, we need just over 2 seconds because we have a sleep step for 2 seconds
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "expr_depend_and_function", 2300*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	ex, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
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

	executionVariables, err := pex.GetExecutionVariables()
	if err != nil {
		assert.Fail("Error getting execution variables", err)
		return
	}
	assert.NotNil(executionVariables)

	echoStepsOutput := executionVariables["step"].AsValueMap()["echo"].AsValueMap()

	if len(echoStepsOutput) != 10 {
		assert.Fail("Invalid number of steps", len(echoStepsOutput))
		return
	}
	assert.Equal(10, len(echoStepsOutput))

	assert.Equal("foo bar", echoStepsOutput["text_1"].AsValueMap()["text"].AsString())
	assert.Equal("lower case Bar Foo Bar Baz and here", echoStepsOutput["text_2"].AsValueMap()["text"].AsString())
	assert.Equal("output 2 Lower Case Bar Foo Bar Baz And Here title(output1) Foo Bar", echoStepsOutput["text_3"].AsValueMap()["text"].AsString())

	// check output for the "time"/"for"/"sleep" steps
	assert.Equal("sleep 2 output: 2s", echoStepsOutput["echo_sleep_1"].AsValueMap()["text"].AsString())
	assert.Equal("sleep 1 output: 1s", echoStepsOutput["echo_sleep_2"].AsValueMap()["text"].AsString())

	sleepStepsOutput := executionVariables["step"].AsValueMap()["sleep"].AsValueMap()
	if sleepStepsOutput == nil {
		assert.Fail("sleep step output not found")
		return
	}

	assert.Equal(1, len(sleepStepsOutput))
	sleep1StepOutputs := sleepStepsOutput["sleep_1"].AsValueMap()
	if sleep1StepOutputs == nil {
		assert.Fail("sleep_1 step output not found")
		return
	}

	assert.Equal(2, len(sleep1StepOutputs))
	assert.Equal("1s", sleep1StepOutputs["0"].AsValueMap()["duration"].AsString())
	assert.Equal("2s", sleep1StepOutputs["1"].AsValueMap()["duration"].AsString())

	assert.Equal(2, len(echoStepsOutput["echo_for_if"].AsValueMap()))
	// First one is OK, the second step should be skipped
	assert.True(len(echoStepsOutput["echo_for_if"].AsValueMap()["0"].AsValueMap()) > 0)
	assert.True(len(echoStepsOutput["echo_for_if"].AsValueMap()["1"].AsValueMap()) == 0)

	assert.Equal(3, len(pex.PipelineOutput))
	assert.Equal("sleep 1 output: 1s", pex.PipelineOutput["one"])
	assert.Equal("Sleep 1 Output: 1s", pex.PipelineOutput["one_function"])
	assert.Equal("2s", pex.PipelineOutput["indexed"])

	// checking the "echo.literal_for" step
	assert.Equal(3, len(echoStepsOutput["literal_for"].AsValueMap()))

	assert.Equal("name is bach", echoStepsOutput["literal_for"].AsValueMap()["0"].AsValueMap()["text"].AsString())
	assert.Equal("name is beethoven", echoStepsOutput["literal_for"].AsValueMap()["1"].AsValueMap()["text"].AsString())
	assert.Equal("name is mozart", echoStepsOutput["literal_for"].AsValueMap()["2"].AsValueMap()["text"].AsString())

	// checking the "echo.literal_for_from_list" step
	assert.Equal(3, len(echoStepsOutput["literal_for_from_list"].AsValueMap()))

	expectedNames := []string{"shostakovitch", "prokofiev", "rachmaninoff"}
	foundNames := []string{
		echoStepsOutput["literal_for_from_list"].AsValueMap()["shostakovitch"].AsValueMap()["text"].AsString(),
		echoStepsOutput["literal_for_from_list"].AsValueMap()["prokofiev"].AsValueMap()["text"].AsString(),
		echoStepsOutput["literal_for_from_list"].AsValueMap()["rachmaninoff"].AsValueMap()["text"].AsString(),
	}

	less := func(a, b string) bool { return a < b }
	equalIgnoreOrder := cmp.Diff(expectedNames, foundNames, cmpopts.SortSlices(less)) == ""
	if !equalIgnoreOrder {
		assert.Fail("literal_for_from_list output not correct")
		return
	}
}

// TODO: VH 2021-10-11 - this test is failing, we need to fix it
func (suite *EsTestSuite) TestIfConditionsOnSteps() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "if", 500*time.Millisecond, nil)
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

	executionVariables, err := pex.GetExecutionVariables()
	if err != nil {
		assert.Fail("Error getting execution variables", err)
		return
	}
	assert.NotNil(executionVariables)

	echoStepsOutput := executionVariables["step"].AsValueMap()["echo"].AsValueMap()
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal(5, len(echoStepsOutput))

	// TODO: we have to check this in the StepStatus now rather than the AllStepOutput attribute (that was removed)
	// assert.Equal("finished", echoStepsOutput["text_true"].(*modconfig.Output).Status)
	// assert.Equal("skipped", echoStepsOutput["text_false"].(*modconfig.Output).Status)
	// assert.Equal("finished", echoStepsOutput["text_1"].(*modconfig.Output).Status)
	// assert.Equal("finished", echoStepsOutput["text_2"].(*modconfig.Output).Status)
	// assert.Equal("skipped", echoStepsOutput["text_3"].(*modconfig.Output).Status)

	assert.Equal("foo", echoStepsOutput["text_true"].AsValueMap()["text"].AsString())
	assert.Equal(0, len(echoStepsOutput["text_false"].AsValueMap()))
	assert.Equal("foo", echoStepsOutput["text_1"].AsValueMap()["text"].AsString())
	assert.Equal("bar", echoStepsOutput["text_2"].AsValueMap()["text"].AsString())
	assert.Equal(0, len(echoStepsOutput["text_3"].AsValueMap()))
}

func (suite *EsTestSuite) TestPipelineErrorBubbleUp() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := runPipeline(suite.FlowpipeTestSuite, "bad_http_one_step", 200*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 40, "failed")
	if err != nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Status)
	assert.NotNil(pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Errors)
	assert.Equal(float64(404), pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Data["status_code"])

	assert.NotNil(pex.PipelineOutput["errors"])
	assert.Equal(int(404), pex.PipelineOutput["errors"].([]modconfig.StepError)[0].Error.Status)
}

func (suite *EsTestSuite) TestParentChildPipeline() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := runPipeline(suite.FlowpipeTestSuite, "parent_pipeline_with_args", 100*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 50, "finished")
	if err != nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)
	// TODO: this doesn't work yet, we need pass the pipeline status up? or does it has its own status?
	// assert.Equal("finished", pex.AllStepOutputs["pipeline"]["child_pipeline_with_args"].(*modconfig.Output).Status)
	// assert.Equal("child echo step: from parent 24", pex.AllNativeStepOutputs["pipeline"]["child_pipeline_with_args"].(*modconfig.Output).Data["output"].(map[string]interface{})["child_output"])
	assert.Equal("child echo step: from parent 24", pex.PipelineOutput["parent_output"])
	assert.Nil(pex.PipelineOutput["does_not_exist"])

}

func (suite *EsTestSuite) TestErrorHandlingOnPipelines() {

	// bad_http_not_ignored pipeline
	assert := assert.New(suite.T())
	_, cmd, err := runPipeline(suite.FlowpipeTestSuite, "bad_http_not_ignored", 500*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 60, "failed")
	if err == nil || (err != nil && err.Error() != "not completed") {
		assert.Fail("Invalid pipeline status", err)
		return
	}

	// This pipeline: bad_http_not_ignored should not complete because there's a step that it can't start
	// so in a way it's "not completed" but it has failed, since it will never be able to start that one step
	assert.False(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Status)
	assert.NotNil(pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Errors)
	assert.Equal(float64(404), pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Data["status_code"])

	// end pipeline test

	// bad_http_ignored pipeline
	_, cmd, err = runPipeline(suite.FlowpipeTestSuite, "bad_http_ignored", 100*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 40, "finished")
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

	assert.Equal("bar", pex.StepStatus["echo.bad_http_if_error_true"]["0"].StepExecutions[0].Output.Data["text"])

	// checking the is_error function working correctly
	assert.Equal("finished", pex.StepStatus["echo.bad_http_if_error_true"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("skipped", pex.StepStatus["echo.bad_http_if_error_false"]["0"].StepExecutions[0].Output.Status)

	// checking the error_message function working correctly
	assert.Equal("404 Not Found", pex.StepStatus["echo.error_message"]["0"].StepExecutions[0].Output.Data["text"])

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_ignored_get_error_code pipeline
	_, cmd, err = runPipeline(suite.FlowpipeTestSuite, "bad_http_ignored_get_error_code", 100*time.Millisecond, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "finished")
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

	assert.Equal(float64(404), pex.StepStatus["http.my_step_1"]["0"].StepExecutions[0].Output.Data["status_code"])
	assert.Equal("404", pex.StepStatus["echo.bad_http"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal("404", output.(string))

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_http_with_for pipeline
	_, cmd, err = runPipeline(suite.FlowpipeTestSuite, "bad_http_with_for", 1*time.Second, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("finished", pex.Status)

	assert.Equal(float64(404), pex.StepStatus["http.http_step"]["0"].StepExecutions[0].Output.Data["status_code"])
	assert.Equal(float64(404), pex.StepStatus["http.http_step"]["1"].StepExecutions[0].Output.Data["status_code"])
	assert.Equal(float64(200), pex.StepStatus["http.http_step"]["2"].StepExecutions[0].Output.Data["status_code"])

	if pex.StepStatus["echo.http_step"] == nil {
		filename := fmt.Sprintf("%s.jsonl", cmd.Event.ExecutionID)
		p := filepath.Join(viper.GetString(constants.ArgLogDir), filename)

		// Open the file
		file, err := os.Open(p)
		if err != nil {
			assert.Fail("Error opening file:", err)
			return
		}
		defer file.Close()

		// Read the file
		byteValue, err := io.ReadAll(file)
		if err != nil {
			assert.Fail("Error reading file:", err)
			return
		}

		// Assuming the JSON is in a format like {"key": "value"}
		var result map[string]interface{}

		// Unmarshal the JSON data into the map
		err = json.Unmarshal(byteValue, &result)
		if err != nil {
			assert.Fail("Error parsing JSON:", err)
			return
		}

		// Print the data
		fmt.Println(result) //nolint:forbidigo // test

		assert.Fail("echo.http_step not found in StepStatus: " + cmd.Event.ExecutionID)
		return
	}

	assert.Equal("skipped", pex.StepStatus["echo.http_step"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("skipped", pex.StepStatus["echo.http_step"]["1"].StepExecutions[0].Output.Status)
	assert.Equal("finished", pex.StepStatus["echo.http_step"]["2"].StepExecutions[0].Output.Status)

	assert.Equal(0, len(pex.StepStatus["echo.http_step"]["0"].StepExecutions[0].Output.Data))
	assert.Equal(0, len(pex.StepStatus["echo.http_step"]["1"].StepExecutions[0].Output.Data))
	assert.Equal("200", pex.StepStatus["echo.http_step"]["2"].StepExecutions[0].Output.Data["text"])

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_email_with_invalid_recipients pipeline
	_, cmd, err = runPipeline(suite.FlowpipeTestSuite, "bad_email_with_invalid_recipients", 1*time.Second, nil)
	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Status)
	assert.NotNil(pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Errors)

	errors := pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Errors
	for _, e := range errors {
		assert.Contains(e.Error.Detail, "no such host")
	}

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test

	// bad_email_with_param pipeline
	_, cmd, err = runPipeline(suite.FlowpipeTestSuite, "bad_email_with_param", 1*time.Second, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 500*time.Millisecond, 5, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.True(pex.IsComplete())
	assert.Equal("failed", pex.Status)

	assert.Equal("failed", pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Status)
	assert.NotNil(pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Errors)

	errors = pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Errors
	for _, e := range errors {
		assert.Contains(e.Error.Detail, "no such host")
	}

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test
}

func (suite *EsTestSuite) TestBadEmail() {
	assert := assert.New(suite.T())

	// bad_email_with_expr
	_, cmd, err := runPipeline(suite.FlowpipeTestSuite, "bad_email_with_expr", 1*time.Second, nil)

	if err != nil {
		assert.Fail("Error running pipeline", err)
		return
	}

	ex, pex, _ := getPipelineExAndWait(suite.FlowpipeTestSuite, cmd.Event, cmd.PipelineExecutionID, 100*time.Millisecond, 50, "failed")

	if pex.Status != "failed" {
		assert.Fail("Pipeline should have failed")
		return
	}

	pipelineDefn, err := ex.PipelineDefinition(cmd.PipelineExecutionID)
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
	assert.Equal("flowpipe@example.com", pex.StepStatus["echo.sender_address"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("This is an email body", pex.StepStatus["echo.email_body"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal("failed", pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Status)
	assert.NotNil(pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Errors)

	// The email step should fail because of the invalid smtp host
	errors := pex.StepStatus["email.test_email"]["0"].StepExecutions[0].Output.Errors
	for _, e := range errors {
		assert.Contains(e.Error.Detail, "no such host")
	}

	// reset ex (so we don't forget if we copy & paste the block)
	ex = nil
	// end pipeline test
}

func (suite *EsTestSuite) TestHttp() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "jsonplaceholder_expr", 500*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 5, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	assert.Equal("finished", pex.StepStatus["echo.output"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("201", pex.StepStatus["echo.output"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal(len(pex.StepStatus["echo.body_json_loop"]), 4)
	assert.Equal("brian may", pex.StepStatus["echo.body_json_loop"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("freddie mercury", pex.StepStatus["echo.body_json_loop"]["1"].StepExecutions[0].Output.Data["text"])
	assert.Equal("roger taylor", pex.StepStatus["echo.body_json_loop"]["2"].StepExecutions[0].Output.Data["text"])
	assert.Equal("john deacon", pex.StepStatus["echo.body_json_loop"]["3"].StepExecutions[0].Output.Data["text"])
}

func (suite *EsTestSuite) TestParam() {
	assert := assert.New(suite.T())

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "param_test", 200*time.Millisecond, nil)
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

	assert.Equal("finished", pex.StepStatus["echo.simple"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("foo", pex.StepStatus["echo.simple"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal("finished", pex.StepStatus["echo.map_echo"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("felix", pex.StepStatus["echo.map_echo"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal(7, len(pex.StepStatus["echo.for_with_list"]))

	assert.Equal("finished", pex.StepStatus["echo.for_with_list"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("Green Day", pex.StepStatus["echo.for_with_list"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal("finished", pex.StepStatus["echo.for_with_list"]["6"].StepExecutions[0].Output.Status)
	assert.Equal("The All-American Rejects", pex.StepStatus["echo.for_with_list"]["6"].StepExecutions[0].Output.Data["text"])

	assert.Equal("finished", pex.StepStatus["echo.map_diff_types_string"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("string", pex.StepStatus["echo.map_diff_types_string"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal("finished", pex.StepStatus["echo.map_diff_types_number"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("1", pex.StepStatus["echo.map_diff_types_number"]["0"].StepExecutions[0].Output.Data["text"])

	assert.Equal(3, len(pex.StepStatus["echo.for_each_list_within_map"]))
	assert.Equal("a", pex.StepStatus["echo.for_each_list_within_map"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("b", pex.StepStatus["echo.for_each_list_within_map"]["1"].StepExecutions[0].Output.Data["text"])
	assert.Equal("c", pex.StepStatus["echo.for_each_list_within_map"]["2"].StepExecutions[0].Output.Data["text"])

	assert.Equal(7, len(pex.StepStatus["echo.for_with_list_and_index"]))
	assert.Equal("0: Green Day", pex.StepStatus["echo.for_with_list_and_index"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("1: New Found Glory", pex.StepStatus["echo.for_with_list_and_index"]["1"].StepExecutions[0].Output.Data["text"])
	assert.Equal("2: Sum 41", pex.StepStatus["echo.for_with_list_and_index"]["2"].StepExecutions[0].Output.Data["text"])
}

func (suite *EsTestSuite) TestParamOverride() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"simple": "bar",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "param_override_test", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 10, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	assert.Equal("finished", pex.StepStatus["echo.simple"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("bar", pex.StepStatus["echo.simple"]["0"].StepExecutions[0].Output.Data["text"])
}

func (suite *EsTestSuite) TestParamOptional() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"simple": "bar",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_param_optional", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 10, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	pipelineParamNull := pex.PipelineOutput["test_output_2"]
	if pipelineParamNull == nil {
		assert.Fail("pipeline output not found")
		return
	}

	assert.Equal("optional and null", pipelineParamNull)
}

func (suite *EsTestSuite) TestParamOverrideWithCtyTypes() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"simple": "bar",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "param_override_test", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 1*time.Second, 10, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	assert.Equal("finished", pex.StepStatus["echo.simple"]["0"].StepExecutions[0].Output.Status)
	assert.Equal("bar", pex.StepStatus["echo.simple"]["0"].StepExecutions[0].Output.Data["text"])
}

func (suite *EsTestSuite) TestChildPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"simple": "bar",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "parent_pipeline", 300*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 300, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	assert.Equal("child echo step", pex.PipelineOutput["parent_output"])

	// TODO: - Check child pipeline status
	// TODO: - Add status on pipeline step
	// TODO: - add multiple childs
}

func (suite *EsTestSuite) TestStepOutput() {
	assert := assert.New(suite.T())
	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "step_output", 500*time.Millisecond, nil)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 2, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	assert.Equal("baz", pex.StepStatus["echo.begin"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("foo", pex.StepStatus["echo.start_step"]["0"].StepExecutions[0].Output.Data["text"])
	assert.Equal("baz", pex.StepStatus["echo.end_step"]["0"].StepExecutions[0].Output.Data["text"])

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEsTestingSuite(t *testing.T) {
	suite.Run(t, &EsTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
