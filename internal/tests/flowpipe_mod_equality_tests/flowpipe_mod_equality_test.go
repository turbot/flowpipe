package pipeline_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/flowpipeconfig"
	fpparse "github.com/turbot/flowpipe/internal/parse"
	"github.com/turbot/flowpipe/internal/tests/test_init"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
)

type FlowpipeModEqualityTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeModEqualityTestSuite) SetupSuite() {

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

	// Create a single, global context for the application
	ctx := context.Background()

	suite.ctx = ctx

	// set app specific constants
	test_init.SetAppSpecificConstants()

	suite.SetupSuiteRunCount++
}

type modEqualityTestCase struct {
	title       string
	description string
	base        string
	compare     string
	equal       bool
}

var modEqualityTestCases = []modEqualityTestCase{
	{
		title:   "base_a == base_a",
		base:    "./base_a",
		compare: "./base_a",
		equal:   true,
	},
	{
		title:   "base_a != base_b",
		base:    "./base_a",
		compare: "./base_b",
		equal:   false,
	},
	{
		title:   "http_step_with_config == http_step_with_config",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config",
		equal:   true,
	},
	{
		title:   "http_step_with_config == http_step_with_config_line_change",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config_line_change",
		equal:   true,
	},
	{
		title:   "http_step_with_config == http_step_with_config_b",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config_b",
		equal:   false,
	},
	{
		title:   "http_step_with_config == http_step_with_config_c",
		base:    "./http_step_with_config",
		compare: "./http_step_with_config_c",
		equal:   false,
	},
	{
		title:       "http_step_with_config_c != http_step_with_config_c_basic_auth_line_change",
		description: "one line change in the basic auth section",
		base:        "./http_step_with_config_c",
		compare:     "./http_step_with_config_c_basic_auth_line_change",
		equal:       true,
	},
	{
		title:   "input_step_a == input_step_a",
		base:    "./input_step_a",
		compare: "./input_step_a",
		equal:   true,
	},
	{
		title:   "input_step_a != input_step_b",
		base:    "./input_step_a",
		compare: "./input_step_b",
		equal:   false,
	},
	{
		title:   "input_step_b == input_step_b",
		base:    "./input_step_b",
		compare: "./input_step_b",
		equal:   true,
	},
	{
		title:   "input_step_a != input_step_c",
		base:    "./input_step_a",
		compare: "./input_step_c",
		equal:   false,
	},
	{
		title:   "input_step_c == input_step_c",
		base:    "./input_step_c",
		compare: "./input_step_c",
		equal:   true,
	},
	{
		title:   "input_step_d != input_step_d",
		base:    "./input_step_d",
		compare: "./input_step_d",
		equal:   true,
	},
	{
		title:   "input_step_d != input_step_d_line_change",
		base:    "./input_step_d",
		compare: "./input_step_d_line_change",
		equal:   true,
	},
	{
		title:   "input_step_d != input_step_e",
		base:    "./input_step_d",
		compare: "./input_step_e",
		equal:   false,
	},
	{
		title:   "container_a == container_a",
		base:    "./container_a",
		compare: "./container_a",
		equal:   true,
	},
	{
		title:   "container_a == container_a_line_change",
		base:    "./container_a",
		compare: "./container_a_line_change",
		equal:   true,
	},
	{
		title:   "container_a != container_b",
		base:    "./container_a",
		compare: "./container_b",
		equal:   false,
	},
	{
		title:   "container_c == container_c",
		base:    "./container_c",
		compare: "./container_c",
		equal:   true,
	},
	{
		title:   "container_c != container_d",
		base:    "./container_c",
		compare: "./container_d",
		equal:   false,
	},
	{
		title:   "container_d == container_d",
		base:    "./container_d",
		compare: "./container_d",
		equal:   true,
	},
	{
		title:       "container_d != container_e",
		description: "cmd attribute has different map values, runtime reference",
		base:        "./container_d",
		compare:     "./container_e",
		equal:       false,
	},
	{
		title:   "container_f == container_f",
		base:    "./container_f",
		compare: "./container_f",
		equal:   true,
	},
	{
		title:       "container_f != container_g",
		description: "cmd attribute has different map values, not runtime reference",
		base:        "./container_f",
		compare:     "./container_g",
		equal:       false,
	},
	{
		title:   "param_a == param_a",
		base:    "./param_a",
		compare: "./param_a",
		equal:   true,
	},
	{
		title:   "param_a == param_a_line_change",
		base:    "./param_a",
		compare: "./param_a_line_change",
		equal:   true,
	},
	{
		title:       "param_a != param_b",
		description: "param b has a param with a different default value, same name",
		base:        "./param_a",
		compare:     "./param_b",
		equal:       false,
	},
	{
		title:   "param_c == param_c",
		base:    "./param_c",
		compare: "./param_c",
		equal:   true,
	},
	{
		title:       "param_c != param_d",
		description: "param d has a param with a different type",
		base:        "./param_c",
		compare:     "./param_d",
		equal:       false,
	},
	{
		title:   "foreach_a == foreach_a",
		base:    "./foreach_a",
		compare: "./foreach_a",
		equal:   true,
	},
	{
		title:   "foreach_a == foreach_a_line_change",
		base:    "./foreach_a",
		compare: "./foreach_a_line_change",
		equal:   true,
	},
	{
		title:       "foreach_a != foreach_b",
		description: "different element in for_each, same length",
		base:        "./foreach_a",
		compare:     "./foreach_b",
		equal:       false,
	},
	{
		title:   "throw_a == throw_a",
		base:    "./throw_a",
		compare: "./throw_a",
		equal:   true,
	},
	{
		title:   "throw_a == throw_a_line_change",
		base:    "./throw_a",
		compare: "./throw_a_line_change",
		equal:   true,
	},
	{
		title:       "throw_a != throw_b",
		description: "different message",
		base:        "./throw_a",
		compare:     "./throw_b",
		equal:       false,
	},
	{
		title:   "throw_b == throw_b",
		base:    "./throw_b",
		compare: "./throw_b",
		equal:   true,
	},
	{
		title:       "throw_b != throw_c",
		description: "different if",
		base:        "./throw_b",
		compare:     "./throw_c",
		equal:       false,
	},
	{
		title:   "throw_c == throw_c",
		base:    "./throw_c",
		compare: "./throw_c",
		equal:   true,
	},
	{
		title:   "output_a == output_a",
		base:    "./output_a",
		compare: "./output_a",
		equal:   true,
	},
	{
		title:       "output_a != output_b",
		description: "value attribute in output is different, also an expression",
		base:        "./output_a",
		compare:     "./output_b",
		equal:       false,
	},
	{
		title:   "output_c == output_c",
		base:    "./output_c",
		compare: "./output_c",
		equal:   true,
	},
	{
		title:       "output_a != output_c",
		description: "change a value in a ternery expression, change wasn't detected at some point",
		base:        "./output_a",
		compare:     "./output_c",
		equal:       false,
	},
	{
		title:   "loop_input_a == loop_input_a",
		base:    "./loop_input_a",
		compare: "./loop_input_a",
		equal:   true,
	},
	{
		title:   "loop_input_a == loop_input_a_line_change",
		base:    "./loop_input_a",
		compare: "./loop_input_a_line_change",
		equal:   true,
	},
	{
		title:       "loop_input_a != loop_input_b",
		description: "different change in the until attribute of the loop",
		base:        "./loop_input_a",
		compare:     "./loop_input_b",
		equal:       false,
	},
	{
		title:   "loop_input_b != loop_input_a_line_change",
		base:    "./loop_input_b",
		compare: "./loop_input_a_line_change",
		equal:   false,
	},
	{
		title:   "loop_sleep_a == loop_sleep_a",
		base:    "./loop_sleep_a",
		compare: "./loop_sleep_a",
		equal:   true,
	},
	{
		title:   "loop_sleep_a == loop_sleep_a_line_change",
		base:    "./loop_sleep_a",
		compare: "./loop_sleep_a_line_change",
		equal:   true,
	},
	{
		title:   "loop_sleep_a != loop_sleep_b",
		base:    "./loop_sleep_a",
		compare: "./loop_sleep_b",
		equal:   false,
	},
	{
		title:   "loop_sleep_b != loop_sleep_a_line_change",
		base:    "./loop_sleep_b",
		compare: "./loop_sleep_a_line_change",
		equal:   false,
	},
	{
		title:   "loop_sleep_b == loop_sleep_b",
		base:    "./loop_sleep_b",
		compare: "./loop_sleep_b",
		equal:   true,
	},
	{
		title:   "loop_sleep_b != loop_sleep_c",
		base:    "./loop_sleep_b",
		compare: "./loop_sleep_c",
		equal:   false,
	},
	{
		title:   "loop_sleep_c == loop_sleep_c",
		base:    "./loop_sleep_c",
		compare: "./loop_sleep_c",
		equal:   true,
	},
	{
		title:   "loop_sleep_c != loop_sleep_d",
		base:    "./loop_sleep_c",
		compare: "./loop_sleep_d",
		equal:   false,
	},
	{
		title:   "loop_http_a == loop_http_a",
		base:    "./loop_http_a",
		compare: "./loop_http_a",
		equal:   true,
	},
	{
		title:   "loop_http_a == loop_http_a_line_change",
		base:    "./loop_http_a",
		compare: "./loop_http_a_line_change",
		equal:   true,
	},
	{
		title:   "loop_http_a != loop_http_b",
		base:    "./loop_http_a",
		compare: "./loop_http_b",
		equal:   false,
	},
	{
		title:   "trigger_a == trigger_a",
		base:    "./trigger_a",
		compare: "./trigger_a",
		equal:   true,
	},
	// Because we're using hcl.Expression in the ArgsRaw attribute
	// {
	// 	title:   "trigger_a == trigger_a_line_change",
	// 	base:    "./trigger_a",
	// 	compare: "./trigger_a_line_change",
	// 	equal:   true,
	// },
	{
		title:   "trigger_a != trigger_b",
		base:    "./trigger_a",
		compare: "./trigger_b",
		equal:   false,
	},
	{
		title:   "trigger_c == trigger_c",
		base:    "./trigger_c",
		compare: "./trigger_c",
		equal:   true,
	},
	{
		// trigger_c: missing param_three compared to trigger_a
		title:   "trigger_a != trigger_c",
		base:    "./trigger_a",
		compare: "./trigger_c",
		equal:   false,
	},
	{
		title:   "trigger_d == trigger_d",
		base:    "./trigger_d",
		compare: "./trigger_d",
		equal:   true,
	},
	{
		// trigger_d: one of the param has a different value 42 vs 43
		title:   "trigger_a != trigger_d",
		base:    "./trigger_a",
		compare: "./trigger_d",
		equal:   false,
	},
	{
		title:   "trigger_e == trigger_e",
		base:    "./trigger_e",
		compare: "./trigger_e",
		equal:   true,
	},
	{
		// trigger_e: one of the param's default value is removed vs trigger_a
		title:   "trigger_a != trigger_e",
		base:    "./trigger_a",
		compare: "./trigger_e",
		equal:   false,
	},
	{
		title: "trigger_a != trigger_f_config_change",
		// trigger_f: config attribute has a different value
		base:    "./trigger_a",
		compare: "./trigger_f_config_change",
		equal:   false,
	},
	{
		title:   "trigger_http_a == trigger_http_a",
		base:    "./trigger_http_a",
		compare: "./trigger_http_a",
		equal:   true,
	},
	{
		title:   "trigger_http_a != trigger_http_b",
		base:    "./trigger_http_a",
		compare: "./trigger_http_b",
		equal:   false,
	},
	{
		title:   "trigger_http_b == trigger_http_b",
		base:    "./trigger_http_b",
		compare: "./trigger_http_b",
		equal:   true,
	},
	{
		title:   "trigger_http_c == trigger_http_c",
		base:    "./trigger_http_c",
		compare: "./trigger_http_c",
		equal:   true,
	},
	{
		title:   "trigger_http_a != trigger_http_c",
		base:    "./trigger_http_a",
		compare: "./trigger_http_c",
		equal:   false,
	},
	{
		title:   "trigger_http_d == trigger_http_d",
		base:    "./trigger_http_d",
		compare: "./trigger_http_d",
		equal:   true,
	},
	{
		title:   "trigger_http_a != trigger_http_d",
		base:    "./trigger_http_a",
		compare: "./trigger_http_d",
		equal:   false,
	},
	{
		title:   "trigger_query_a == trigger_query_a",
		base:    "./trigger_query_a",
		compare: "./trigger_query_a",
		equal:   true,
	},
	{
		title: "trigger_query_a != trigger_query_b",
		// trigger_query_b: updated SQL
		base:    "./trigger_query_a",
		compare: "./trigger_query_b",
		equal:   false,
	},
	{
		title:   "trigger_query_b == trigger_query_b",
		base:    "./trigger_query_b",
		compare: "./trigger_query_b",
		equal:   true,
	},
	{
		title: "trigger_query_a != trigger_query_c",
		// trigger_query_c: updated database attribute
		base:    "./trigger_query_a",
		compare: "./trigger_query_c",
		equal:   false,
	},
	{
		title: "trigger_query_a != trigger_query_d",
		// trigger_query_d: removed on of the captures
		base:    "./trigger_query_a",
		compare: "./trigger_query_d",
		equal:   false,
	},
	{
		title:   "trigger_query_e == trigger_query_e",
		base:    "./trigger_query_e",
		compare: "./trigger_query_e",
		equal:   true,
	},
	{
		title: "trigger_query_a != trigger_query_e",
		// change the pipeline in one of the captures
		base:    "./trigger_query_a",
		compare: "./trigger_query_e",
		equal:   false,
	},
}

const (
	TARGET_DIR = "./target_dir"
)

func (suite *FlowpipeModEqualityTestSuite) TestFlowpipeModEquality() {

	for _, tc := range modEqualityTestCases {
		suite.T().Run(tc.title, func(t *testing.T) {
			assert := assert.New(t)
			utils.EmptyDir(TARGET_DIR)         //nolint:errcheck // test only
			utils.CopyDir(tc.base, TARGET_DIR) //nolint:errcheck // test only

			flowpipeConfigA, ew := flowpipeconfig.LoadFlowpipeConfig([]string{TARGET_DIR})
			if ew.Error != nil {
				assert.FailNow(ew.Error.Error())
				return
			}

			notiferMapA, err := flowpipeConfigA.NotifierValueMap()
			if err != nil {
				assert.FailNow(err.Error())
				return
			}

			wA, errorAndWarning := workspace.Load(suite.ctx,
				TARGET_DIR,
				workspace.WithDecoderOptions(fpparse.WithCredentials(flowpipeConfigA.Credentials)),
				workspace.WithConfigValueMap("notifier", notiferMapA))

			assert.NotNil(wA)
			assert.Nil(errorAndWarning.Error)
			assert.Equal(0, len(errorAndWarning.Warnings))

			utils.EmptyDir(TARGET_DIR)            //nolint:errcheck // test only
			utils.CopyDir(tc.compare, TARGET_DIR) //nolint:errcheck // test only

			flowpipeConfigB, ew := flowpipeconfig.LoadFlowpipeConfig([]string{TARGET_DIR})
			if ew.Error != nil {
				assert.FailNow(ew.Error.Error())
				return
			}

			notiferMapB, err := flowpipeConfigB.NotifierValueMap()
			if err != nil {
				assert.FailNow(err.Error())
				return
			}

			wB, ew := workspace.Load(suite.ctx,
				TARGET_DIR,
				workspace.WithDecoderOptions(fpparse.WithCredentials(flowpipeConfigB.Credentials)),
				workspace.WithConfigValueMap("notifier", notiferMapB))

			assert.NotNil(wB)
			assert.Nil(ew.Error)
			assert.Equal(0, len(ew.Warnings))

			assert.Equal(tc.equal, wA.GetModResources().Equals(wB.GetModResources()))
		})
	}

}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeModEqualityTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeModEqualityTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeModEqualityTestSuite))
}
