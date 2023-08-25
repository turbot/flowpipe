package pipeline_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/filepaths"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/versionmap"
	filehelpers "github.com/turbot/go-kit/files"
)

type FpTestSuite struct {
	suite.Suite
	SetupSuiteRunCount    int
	TearDownSuiteRunCount int
	ctx                   context.Context
}

func (suite *FpTestSuite) SetupSuite() {

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

	ctx = fplog.ContextWithLogger(ctx)
	ctx, err = config.ContextWithConfig(ctx)
	if err != nil {
		panic(err)
	}

	suite.ctx = ctx

	filepaths.PipesComponentWorkspaceDataDir = ".flowpipe"
	filepaths.PipesComponentModsFileName = "mod.hcl"

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FpTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

func (suite *FpTestSuite) TestModDependencies() {
	assert := assert.New(suite.T())

	workspaceLock, err := versionmap.LoadWorkspaceLock("./mod_dep_one")

	assert.Nil(err, "error loading workspace lock")

	parseCtx := parse.NewModParseContext(
		suite.ctx,
		workspaceLock,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/*.hcl"},
		})

	mod, errorsAndWarnings := pipeparser.LoadModWithFileName("./mod_dep_one", filepaths.PipesComponentModsFileName, parseCtx)
	if errorsAndWarnings != nil && errorsAndWarnings.Error != nil {
		assert.Fail("error loading mod file", errorsAndWarnings.Error.Error())
		return
	}

	pipelines := mod.ResourceMaps.Pipelines

	assert.NotNil(mod, "mod is nil")
	jsonForPipeline := pipelines["mod_parent.pipeline.json"]
	if jsonForPipeline == nil {
		assert.Fail("json pipeline not found")
		return
	}

	childMod := mod.ResourceMaps.Mods["mod_child_a@v1.0.0"]
	assert.NotNil(childMod)

	thisPipelineIsInTheChildPipeline := childMod.ResourceMaps.Pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	assert.NotNil(thisPipelineIsInTheChildPipeline)

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFpTestSuite(t *testing.T) {
	suite.Run(t, new(FpTestSuite))
}
