package pipeline_test

import (
	"context"
	"github.com/turbot/flowpipe/internal/flowpipeconfig"
	"github.com/turbot/flowpipe/internal/tests/test_init"
	"os"
	"path"
	"testing"

	"github.com/turbot/pipe-fittings/utils"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FlowpipeConfigEqualityTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FlowpipeConfigEqualityTestSuite) SetupSuite() {

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

	pipelineDirPath := path.Join(cwd, "pipelines")

	viper.GetViper().Set("pipeline.dir", pipelineDirPath)
	viper.GetViper().Set("output.dir", outputPath)
	viper.GetViper().Set("log.dir", outputPath)

	// Create a single, global context for the application
	ctx := context.Background()

	suite.ctx = ctx

	// set app specific constants
	test_init.SetAppSpecificConstants()

	suite.SetupSuiteRunCount++
}

type flowpipeConfigEqualityTestCase struct {
	title   string
	base    string
	compare string
	equal   bool
}

var flowpipeConfigEqualityTestCases = []flowpipeConfigEqualityTestCase{
	{
		title:   "test: base == base",
		base:    "./config_notifier_base",
		compare: "./config_notifier_base",
		equal:   true,
	},
	{
		title:   "test: base != base_b",
		base:    "./config_notifier_base",
		compare: "./config_notifier_base_b",
		equal:   false,
	},
	{
		title:   "test: base_b == base_b",
		base:    "./config_notifier_base_b",
		compare: "./config_notifier_base_b",
		equal:   true,
	},
	{
		title:   "test: base_b != base_c",
		base:    "./config_notifier_base_b",
		compare: "./config_notifier_base_c",
		equal:   false,
	},
	{
		title:   "test: base_c == base_c",
		base:    "./config_notifier_base_c",
		compare: "./config_notifier_base_c",
		equal:   true,
	},
	{
		title:   "test: base_c != base_d",
		base:    "./config_notifier_base_c",
		compare: "./config_notifier_base_d",
		equal:   false,
	},
	{
		title:   "test: base_d == base_d",
		base:    "./config_notifier_base_d",
		compare: "./config_notifier_base_d",
		equal:   true,
	},
	{
		title:   "test: base_e == base_e",
		base:    "./config_notifier_base_e",
		compare: "./config_notifier_base_e",
		equal:   true,
	},
	{
		title:   "test: base_e != base_f",
		base:    "./config_notifier_base_e",
		compare: "./config_notifier_base_f",
		equal:   false,
	},
}

const (
	TARGET_DIR = "./config_notifier_target"
)

func (suite *FlowpipeConfigEqualityTestSuite) TestFlowpipeConfigEquality() {

	for _, tc := range flowpipeConfigEqualityTestCases {
		suite.T().Run(tc.title, func(t *testing.T) {
			assert := assert.New(t)
			utils.EmptyDir(TARGET_DIR)         //nolint:errcheck // test only
			utils.CopyDir(tc.base, TARGET_DIR) //nolint:errcheck // test only

			flowpipeConfigA, err := flowpipeconfig.LoadFlowpipeConfig([]string{TARGET_DIR})
			assert.Nil(err.Error)

			utils.EmptyDir(TARGET_DIR)            //nolint:errcheck // test only
			utils.CopyDir(tc.compare, TARGET_DIR) //nolint:errcheck // test only

			flowpipeConfigB, err := flowpipeconfig.LoadFlowpipeConfig([]string{TARGET_DIR})
			assert.Nil(err.Error)

			assert.Equal(tc.equal, flowpipeConfigA.Equals(flowpipeConfigB))
		})
	}
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FlowpipeConfigEqualityTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFlowpipeConfigEqualityTestSuite(t *testing.T) {
	suite.Run(t, new(FlowpipeConfigEqualityTestSuite))
}
