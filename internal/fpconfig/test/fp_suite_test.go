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

	childModA := mod.ResourceMaps.Mods["mod_child_a@v1.0.0"]
	assert.NotNil(childModA)

	thisPipelineIsInTheChildPipelineModA := childModA.ResourceMaps.Pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	assert.NotNil(thisPipelineIsInTheChildPipelineModA)

}

func (suite *FpTestSuite) TestModDependenciesBackwardCompatible() {
	assert := assert.New(suite.T())

	workspaceLock, err := versionmap.LoadWorkspaceLock("./backward_compatible_mod")

	assert.Nil(err, "error loading workspace lock")

	parseCtx := parse.NewModParseContext(
		suite.ctx,
		workspaceLock,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files | filehelpers.Recursive,
			Exclude: []string{"./.flowpipe/**/*.*"},
			Include: []string{"**/*.hcl", "**/*.sp"},
		})

	mod, errorsAndWarnings := pipeparser.LoadModWithFileName("./backward_compatible_mod", filepaths.PipesComponentModsFileName, parseCtx)
	if errorsAndWarnings != nil && errorsAndWarnings.Error != nil {
		assert.Fail("error loading mod file", errorsAndWarnings.Error.Error())
		return
	}

	pipelines := mod.ResourceMaps.Pipelines

	assert.Equal(6, len(pipelines), "wrong number of pipelines")

	assert.NotNil(mod, "mod is nil")
	jsonForPipeline := pipelines["mod_parent.pipeline.json"]
	if jsonForPipeline == nil {
		assert.Fail("json pipeline not found")
		return
	}

	parentPipelineHcl := pipelines["mod_parent.pipeline.parent_pipeline_hcl"]
	assert.NotNil(parentPipelineHcl)

	parentPipelineHclB := pipelines["mod_parent.pipeline.parent_pipeline_hcl_b"]
	assert.NotNil(parentPipelineHclB)

	parentPipelineHclNested := pipelines["mod_parent.pipeline.parent_pipeline_hcl_nested"]
	assert.NotNil(parentPipelineHclNested)

	// This should be nil, there was a bug that was causing the child pipelines to be loaded in the parent mod
	thisPipelineIsInTheChildParent := pipelines["mod_parent.pipeline.this_pipeline_is_in_the_child"]
	assert.Nil(thisPipelineIsInTheChildParent)

	nestedPipeInChildHclParent := pipelines["mod_parent.pipeline.nested_pipe_in_child_hcl"]
	assert.Nil(nestedPipeInChildHclParent)

	// SP file format
	parentPipelineSp := pipelines["mod_parent.pipeline.parent_pipeline_sp"]
	assert.NotNil(parentPipelineSp)

	parentPipelineSpNested := pipelines["mod_parent.pipeline.parent_pipeline_sp_nested"]
	assert.NotNil(parentPipelineSpNested)

	childModA := mod.ResourceMaps.Mods["mod_child_a@v1.0.0"]
	assert.NotNil(childModA)

	thisPipelineIsInTheChildPipelineModA := childModA.ResourceMaps.Pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	assert.NotNil(thisPipelineIsInTheChildPipelineModA)

	childModB := mod.ResourceMaps.Mods["mod_child_b@v2.0.0"]
	assert.NotNil(childModB)

	thisPipelineIsInTheChildPipelineModB := childModB.ResourceMaps.Pipelines["mod_child_b.pipeline.this_pipeline_is_in_the_child"]
	assert.NotNil(thisPipelineIsInTheChildPipelineModB)

	anotherChildPipelineModB := childModB.ResourceMaps.Pipelines["mod_child_b.pipeline.another_child_pipeline"]
	assert.NotNil(anotherChildPipelineModB)

	secondPipeInTheChildModB := childModB.ResourceMaps.Pipelines["mod_child_b.pipeline.second_pipe_in_the_child"]
	assert.NotNil(secondPipeInTheChildModB)

	nestedPipeInTheChildModB := childModB.ResourceMaps.Pipelines["mod_child_b.pipeline.nested_pipe_in_child_hcl"]
	assert.NotNil(nestedPipeInTheChildModB)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFpTestSuite(t *testing.T) {
	suite.Run(t, new(FpTestSuite))
}
