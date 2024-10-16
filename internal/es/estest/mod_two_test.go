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

func (suite *ModTwoTestSuite) TestValidCustomParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"aws_conn": map[string]string{
			"short_name":    "example_2",
			"name":          "aws.example_2",
			"type":          "aws",
			"resource_type": "connection",
		},
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.conn_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("example2_access_key", pex.PipelineOutput["val"])

	pipelineInput = modconfig.Input{
		"aws_conn": map[string]string{
			"short_name":    "example_3",
			"name":          "aws.example_3",
			"type":          "aws",
			"resource_type": "connection",
		},
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.conn_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("example3_access_key", pex.PipelineOutput["val"])

}

func (suite *ModTwoTestSuite) TestInvalidCustomParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"aws_conn": "foo",
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.conn_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("Internal Error: aws_conn: Invalid type for param aws_conn: The param type is not compatible with the given value", pex.Errors[0].Error.Error())

	// Wrong connection, expect slack not aws
	pipelineInput = modconfig.Input{
		"aws_conn": map[string]string{
			"short_name":    "example_2",
			"name":          "slack.example_2",
			"type":          "slack",
			"resource_type": "connection",
		},
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.conn_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("Internal Error: aws_conn: Invalid type for param aws_conn: The param type is not compatible with the given value", pex.Errors[0].Error.Error())

	// connection not found
	pipelineInput = modconfig.Input{
		"aws_conn": map[string]string{
			"name":          "aws.example_50",
			"short_name":    "example_50",
			"type":          "aws",
			"resource_type": "connection",
		},
	}

	_, pipelineCmd, err = runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.conn_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("Internal Error: aws_conn: No connection found for the given connection name: example_50", pex.Errors[0].Error.Error())
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

func (suite *ModTwoTestSuite) TestNotifierParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"notifier": map[string]interface{}{
			"resource_type": "notifier",
			"name":          "backend",
		},
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("backend", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestNotifierVarParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_var_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("admin", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestNotifierVar() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_var", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("admin", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestNotifierParamChildPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"notifier": map[string]interface{}{
			"resource_type": "notifier",
			"name":          "backend",
		},
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_param_parent", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("backend", pex.PipelineOutput["val"].(map[string]any)["value"].(map[string]any)["title"])
}

func (suite *ModTwoTestSuite) TestListNotifierParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"notifiers": []map[string]interface{}{
			{
				"resource_type": "notifier",
				"name":          "backend",
			},
			{
				"resource_type": "notifier",
				"name":          "admin",
			},
		},
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_list_param", 100*time.Millisecond, pipelineInput)

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
	//assert.Equal("backend", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestListNotifierDefaultParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_list_param", 100*time.Millisecond, pipelineInput)

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
}

func (suite *ModTwoTestSuite) TestInvalidNotifierParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"notifier": map[string]interface{}{
			"resource_type": "notifier",
			"name":          "does_not_exist",
		},
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_param", 100*time.Millisecond, pipelineInput)

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
	assert.True(len(pex.Errors) > 0)
	assert.Equal("Bad Request: notifier not found: does_not_exist", pex.Errors[0].Error.Error())
}

func (suite *ModTwoTestSuite) TestNotifierDefaultParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.notifier_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("frontend", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestSteampipeConn() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.steampipe_conn", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("default_conn_string", pex.PipelineOutput["val"].(string))
}

func (suite *ModTwoTestSuite) TestSteampipeConnWithParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.steampipe_conn_with_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("default_conn_string", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestSteampipeConnectionVarParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.connection_var_param", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("default_conn_string", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestSteampipeConnectionVar() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.connection_var", 100*time.Millisecond, pipelineInput)

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
	assert.Equal("default_conn_string", pex.PipelineOutput["val"].(map[string]any)["value"])
}

func (suite *ModTwoTestSuite) TestConnectionParamChildPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{
		"Connection": map[string]interface{}{
			"resource_type": "Connection",
			"name":          "backend",
		},
	}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.connection_param_parent", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("default_conn_string", pex.PipelineOutput["val"].(map[string]any)["value"].(string))
}

func (suite *ModTwoTestSuite) TestConnectionReferenceFromAnotherStep() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.from_another_step", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("example2_access_key", pex.PipelineOutput["val"].(map[string]any)["value"].(map[string]any)["access_key"])
}

func (suite *ModTwoTestSuite) TestConnectionReferenceFromParam() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.from_param", 100*time.Millisecond, pipelineInput)

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

	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pex.PipelineOutput["val"].(map[string]any)["value"].(map[string]any)["access_key"])
}

func (suite *ModTwoTestSuite) TestConnectionReferenceWithForEach() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.foreach_with_conn_simple", 100*time.Millisecond, pipelineInput)

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

	pipelineOutput := pex.PipelineOutput["val"]
	assert.Equal(3, len(pipelineOutput.(map[string]any)))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pipelineOutput.(map[string]any)["0"].(map[string]any)["value"].(map[string]any)["access_key"])
	assert.Equal("example2_access_key", pipelineOutput.(map[string]any)["1"].(map[string]any)["value"].(map[string]any)["access_key"])
	assert.Equal("example3_access_key", pipelineOutput.(map[string]any)["2"].(map[string]any)["value"].(map[string]any)["access_key"])
}

func (suite *ModTwoTestSuite) TestConnectionReferenceWithForEachInLiteral() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.foreach_with_conn_literal", 100*time.Millisecond, pipelineInput)

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

	pipelineOutput := pex.PipelineOutput["val"]
	assert.Equal(3, len(pipelineOutput.(map[string]any)))
	assert.Equal("Foo: bar and ASIAQGDFAKEKGUI5MCEU", pipelineOutput.(map[string]any)["0"].(map[string]any)["value"].(string))
	assert.Equal("Foo: bar and example2_access_key", pipelineOutput.(map[string]any)["1"].(map[string]any)["value"].(string))
	assert.Equal("Foo: bar and example3_access_key", pipelineOutput.(map[string]any)["2"].(map[string]any)["value"].(string))
}

func (suite *ModTwoTestSuite) TestConnectionReferenceWithForEachInObject() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.foreach_with_conn_object", 100*time.Millisecond, pipelineInput)

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

	pipelineOutput := pex.PipelineOutput["val"]
	assert.Equal(3, len(pipelineOutput.(map[string]any)))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pipelineOutput.(map[string]any)["0"].(map[string]any)["value"].(map[string]any)["akey"].(map[string]any)["access_key"])
	assert.Equal("example2_access_key", pipelineOutput.(map[string]any)["1"].(map[string]any)["value"].(map[string]any)["akey"].(map[string]any)["access_key"])
	assert.Equal("example3_access_key", pipelineOutput.(map[string]any)["2"].(map[string]any)["value"].(map[string]any)["akey"].(map[string]any)["access_key"])
}

func (suite *ModTwoTestSuite) TestForEachConnectedNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.parent_foreach_connection", 100*time.Millisecond, pipelineInput)

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

	pipelineOutput := pex.PipelineOutput["val"]
	assert.Equal(2, len(pipelineOutput.(map[string]any)))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pipelineOutput.(map[string]any)["0"].(map[string]any)["args"].(map[string]any)["conn"].(map[string]any)["access_key"])
	assert.Equal("example2_access_key", pipelineOutput.(map[string]any)["1"].(map[string]any)["args"].(map[string]any)["conn"].(map[string]any)["access_key"])
}

func (suite *ModTwoTestSuite) TestComplexForEachConnectedNestedPipeline() {
	assert := assert.New(suite.T())

	pipelineInput := modconfig.Input{}

	_, pipelineCmd, err := runPipeline(suite.FlowpipeTestSuite, "test_suite_mod_2.pipeline.parent_foreach_connection", 100*time.Millisecond, pipelineInput)

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

	pipelineOutput := pex.PipelineOutput["val"]
	assert.Equal(2, len(pipelineOutput.(map[string]any)))
	assert.Equal("ASIAQGDFAKEKGUI5MCEU", pipelineOutput.(map[string]any)["0"].(map[string]any)["args"].(map[string]any)["conn"].(map[string]any)["access_key"])
	assert.Equal("example2_access_key", pipelineOutput.(map[string]any)["1"].(map[string]any)["args"].(map[string]any)["conn"].(map[string]any)["access_key"])
}

func TestModTwoTestingSuite(t *testing.T) {
	suite.Run(t, &ModTwoTestSuite{
		FlowpipeTestSuite: &FlowpipeTestSuite{},
	})
}
