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
	viper.GetViper().Set("pipeline.dir", "./pipelines")
	viper.GetViper().Set("output.dir", "./output")
	viper.GetViper().Set("log.dir", "./output")
	// viper.GetViper().Set("pipeline.dir", "./tmp")

	// Create a single, global context for the application
	ctx := context.Background()

	ctx = fplog.ContextWithLogger(ctx)
	ctx, err := config.ContextWithConfig(ctx)
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

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *EsTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(1 * time.Second)
	suite.TearDownSuiteRunCount++
}

func (suite *EsTestSuite) BeforeTest(suiteName, testName string) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// clear the tmp dir (output dir) before each test
	dirPath := path.Join(cwd, "output")

	// Check if the directory exists
	_, err = os.Stat(dirPath)
	if !os.IsNotExist(err) {
		// Remove the directory and its contents
		err = os.RemoveAll(dirPath)
		if err != nil {
			panic(err)
		}

	}

	// Check if the directory already exists
	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		err := os.Mkdir(dirPath, 0755) // 0755 sets read-write-execute permissions for owner and read-execute permissions for group and others
		if err != nil {
			panic(err)
		}
	}

}

func (suite *EsTestSuite) AfterTest(suiteName, testName string) {

}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *EsTestSuite) TestExpressionWithDependenciesFunctions() {
	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(suite.ctx),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                "expr_depend_and_function",
	}

	if err := suite.esService.Send(pipelineCmd); err != nil {
		assert.Fail(suite.T(), "Error sending pipeline command", err)
		return
	}

	pipelineExecutionID := pipelineCmd.PipelineExecutionID

	// give it a moment to let Watermill does its thing
	time.Sleep(100 * time.Millisecond)

	// check if the execution id has been completed
	ex, err := execution.NewExecution(suite.ctx)
	if err != nil {
		assert.Fail(suite.T(), "Error creating execution", err)
		return
	}
	err = ex.LoadProcess(pipelineCmd.Event)
	if err != nil {
		assert.Fail(suite.T(), "Error loading process", err)
		return
	}

	pex := ex.PipelineExecutions[pipelineExecutionID]
	if pex == nil {
		assert.Fail(suite.T(), "Pipeline execution not found")
		return
	}

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < 3 && !pex.IsComplete(); i++ {
		time.Sleep(100 * time.Millisecond)
	}

	if !pex.IsComplete() {
		assert.Fail(suite.T(), "Pipeline execution not complete")
		return
	}

	echoStepsOutput := ex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail(suite.T(), "echo step output not found")
		return
	}

	assert.Equal(suite.T(), 3, len(echoStepsOutput))
	assert.Equal(suite.T(), "foo bar", echoStepsOutput["text_1"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal(suite.T(), "lower case Bar Foo Bar Baz and here", echoStepsOutput["text_2"].(*types.StepOutput).OutputVariables["text"])
	assert.Equal(suite.T(), "output 2 Lower Case Bar Foo Bar Baz And Here title(output1) Foo Bar", echoStepsOutput["text_3"].(*types.StepOutput).OutputVariables["text"])
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(EsTestSuite))
}
