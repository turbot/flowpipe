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

type ModTwoTestSuite struct {
	suite.Suite
	*FlowpipeTestSuite

	server                *http.Server
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
}

// The SetupSuite method will be run by testify once, at the very
// start of the testing suite, before any tests are run.
func (suite *ModTwoTestSuite) SetupSuite() {

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

	pipelineDirPath := path.Join(cwd, "test_suite_mod_2")

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
func (suite *ModTwoTestSuite) TearDownSuite() {
	// Wait for a bit to allow the Watermill to finish running the pipelines
	time.Sleep(3 * time.Second)

	err := suite.esService.Stop()
	if err != nil {
		panic(err)
	}

	suite.server.Shutdown(suite.ctx) //nolint:errcheck // just a test case
	suite.TearDownSuiteRunCount++

	time.Sleep(1 * time.Second)
}

func (suite *ModTwoTestSuite) BeforeTest(suiteName, testName string) {

}

func (suite *ModTwoTestSuite) AfterTest(suiteName, testName string) {
}

func (suite *ModTwoTestSuite) TestEnumParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

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
}

func (suite *ModTwoTestSuite) TestValidParam() {
	assert := assert.New(suite.T())

	/* param definition
	   param "string_param" {
	          type = string
	          default = "value1"
	          enum = ["value1", "value2", "value3"]
	          tags = {
	              "tag3" = "value3"
	              "tag4" = "value4"
	          }
	      }

	*/
	pipelineInput := modconfig.Input{
		"string_param": "value2",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	/*
			Number param

			    param "num_param" {
		        type = number
		        default = 1
		        enum = [1, 2, 3]
		        tags = {
		            "tag5" = "value5"
		            "tag6" = "value6"
		        }
		    }
	*/

	pipelineInput = modconfig.Input{
		"num_param": 2,
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)

	/*

			    param "list_of_string_param" {
		        type = list(string)
		        default = ["value1", "value2"]
		        enum = ["value1", "value2", "value3"]
		        tags = {
		            "tag7" = "value7"
		            "tag8" = "value8"
		        }
		    }
	*/

	pipelineInput = modconfig.Input{
		"list_of_string_param": []string{"value2", "value3"},
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "finished")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("finished", pex.Status)
}

func (suite *ModTwoTestSuite) TestInvalidParam() {
	assert := assert.New(suite.T())

	/* param definition
	    param "string_param" {
	           type = string
	           default = "value1"
	           enum = ["value1", "value2", "value3"]
	           tags = {
	               "tag3" = "value3"
	               "tag4" = "value4"
	           }
	       }

	   	we will supply a value that is not in the num
	*/
	pipelineInput := modconfig.Input{
		"string_param": "value4",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err := getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("failed", pex.Status)
	assert.Equal("Bad Request: invalid value for param string_param", pex.Errors[0].Error.Error())

	/*
			Number param

			    param "num_param" {
		        type = number
		        default = 1
		        enum = [1, 2, 3]
		        tags = {
		            "tag5" = "value5"
		            "tag6" = "value6"
		        }
		    }
	*/

	pipelineInput = modconfig.Input{
		"num_param": 5,
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("failed", pex.Status)
	assert.Equal("Bad Request: invalid value for param num_param", pex.Errors[0].Error.Error())

	/*

			    param "list_of_string_param" {
		        type = list(string)
		        default = ["value1", "value2"]
		        enum = ["value1", "value2", "value3"]
		        tags = {
		            "tag7" = "value7"
		            "tag8" = "value8"
		        }
		    }
	*/

	pipelineInput = modconfig.Input{
		"list_of_string_param": []string{"value3", "value4"},
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.enum_param", 100*time.Millisecond, pipelineInput)

	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	_, pex, err = getPipelineExAndWait(suite.FlowpipeTestSuite, pipelineCmd.Event, pipelineCmd.PipelineExecutionID, 100*time.Millisecond, 100, "failed")
	if err != nil {
		assert.Fail("Error getting pipeline execution", err)
		return
	}

	assert.Equal("failed", pex.Status)
	assert.Equal("Bad Request: invalid value for param list_of_string_param", pex.Errors[0].Error.Error())
}

func TestModTwoTestingSuite(t *testing.T) {
	suite.Run(t, &ModTwoTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
