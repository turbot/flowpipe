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
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/utils"
)

type ModTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *ModTestSuite) SetupSuite() {

	err := os.Setenv("RUN_MODE", "TEST_ES")
	if err != nil {
		panic(err)
	}

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

	pipelineDirPath := path.Join(cwd, "default_mod")

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
	esService.StartedAt = utils.TimeNow()

	suite.esService = esService

	// Give some time for Watermill to fully start
	time.Sleep(2 * time.Second)

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *ModTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}
	suite.TearDownSuiteRunCount++
}

func (suite *ModTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *ModTestSuite) AfterTest(suiteName, testName string) {
	time.Sleep(2 * time.Second)
}

func (suite *ModTestSuite) XXTestPipelineWithParam() {
	assert := assert.New(suite.T())

	pipelineInput := &modconfig.Input{
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

	echoStepsOutput := pex.AllStepOutputs["echo"]
	if echoStepsOutput == nil {
		assert.Fail("echo step output not found")
		return
	}

	assert.Equal("finished", echoStepsOutput["simple"].(*modconfig.Output).Status)
	assert.Equal("bar", echoStepsOutput["simple"].(*modconfig.Output).Data["text"])
}

func TestModTestingSuite(t *testing.T) {
	suite.Run(t, &ModTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
