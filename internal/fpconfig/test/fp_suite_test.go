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
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/filepaths"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/workspace"
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
	filepaths.PipesComponentDefaultVarsFileName = "flowpipe.pvars"
	filepaths.PipesComponentDefaultInstallDir = "~/.flowpipe"

	constants.PipesComponentModDataExtension = ".hcl"
	constants.PipesComponentVariablesExtension = ".pvars"
	constants.PipesComponentAutoVariablesExtension = ".auto.pvars"
	constants.PipesComponentEnvInputVarPrefix = "P_VAR_"

	suite.SetupSuiteRunCount++
}

// The TearDownSuite method will be run by testify once, at the very
// end of the testing suite, after all tests have been run.
func (suite *FpTestSuite) TearDownSuite() {
	suite.TearDownSuiteRunCount++
}

func (suite *FpTestSuite) TestGoodMod() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./good_mod", []string{".hcl"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.Equal(len(pipelines), 3, "wrong number of pipelines")

	jsonForPipeline := pipelines["test_mod.pipeline.json_for"]
	if jsonForPipeline == nil {
		assert.Fail("json_for pipeline not found")
		return
	}

	// check if all steps are there
	assert.Equal(2, len(jsonForPipeline.Steps), "wrong number of steps")
	assert.Equal(jsonForPipeline.Steps[0].GetName(), "json", "wrong step name")
	assert.Equal(jsonForPipeline.Steps[0].GetType(), "echo", "wrong step type")
	assert.Equal(jsonForPipeline.Steps[1].GetName(), "json_for", "wrong step name")
	assert.Equal(jsonForPipeline.Steps[1].GetType(), "echo", "wrong step type")

	// check if all triggers are there
	triggers := mod.ResourceMaps.Triggers
	assert.Equal(1, len(triggers), "wrong number of triggers")
	assert.Equal("test_mod.trigger.schedule.my_hourly_trigger", triggers["test_mod.trigger.schedule.my_hourly_trigger"].FullName, "wrong trigger name")
}

func (suite *FpTestSuite) TestModReferences() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_references", []string{".hcl"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")
	assert.Equal(2, len(pipelines), "wrong number of pipelines")
	assert.NotNil(pipelines["pipeline_with_references.pipeline.foo"])
	assert.NotNil(pipelines["pipeline_with_references.pipeline.foo_two"])
}

func (suite *FpTestSuite) TestStepOutputParsing() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_with_step_output", []string{".hcl"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")
	assert.Equal(1, len(pipelines), "wrong number of pipelines")

	assert.Equal(2, len(pipelines["test_mod.pipeline.with_step_output"].Steps), "wrong number of steps")
	assert.False(pipelines["test_mod.pipeline.with_step_output"].Steps[0].IsResolved())
	assert.False(pipelines["test_mod.pipeline.with_step_output"].Steps[1].IsResolved())

}

func (suite *FpTestSuite) TestModDependencies() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_dep_one", []string{".hcl"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines

	assert.NotNil(mod, "mod is nil")
	jsonForPipeline := pipelines["mod_parent.pipeline.json"]
	if jsonForPipeline == nil {
		assert.Fail("json pipeline not found")
		return
	}

	fooPipeline := pipelines["mod_parent.pipeline.foo"]
	if fooPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	fooTwoPipeline := pipelines["mod_parent.pipeline.foo_two"]
	if fooTwoPipeline == nil {
		assert.Fail("foo_two pipeline not found")
		return
	}

	referToChildPipeline := pipelines["mod_parent.pipeline.refer_to_child"]
	if referToChildPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	referToChildBPipeline := pipelines["mod_parent.pipeline.refer_to_child_b"]
	if referToChildBPipeline == nil {
		assert.Fail("refer_to_child_b pipeline not found")
		return
	}

	childModA := mod.ResourceMaps.Mods["mod_child_a@v1.0.0"]
	assert.NotNil(childModA)

	thisPipelineIsInTheChildPipelineModA := childModA.ResourceMaps.Pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	assert.NotNil(thisPipelineIsInTheChildPipelineModA)

	// check for the triggers
	triggers := mod.ResourceMaps.Triggers
	myHourlyTrigger := triggers["mod_parent.trigger.schedule.my_hourly_trigger"]
	if myHourlyTrigger == nil {
		assert.Fail("my_hourly_trigger not found")
		return
	}

}

func (suite *FpTestSuite) TestModDependenciesSimple() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_dep_simple", []string{".hcl"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines
	jsonForPipeline := pipelines["mod_parent.pipeline.json"]
	if jsonForPipeline == nil {
		assert.Fail("json pipeline not found")
		return
	}

	fooPipeline := pipelines["mod_parent.pipeline.foo"]
	if fooPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	assert.Equal(2, len(fooPipeline.Steps), "wrong number of steps")
	assert.Equal("baz", fooPipeline.Steps[0].GetName())
	assert.Equal("bar", fooPipeline.Steps[1].GetName())

	referToChildPipeline := pipelines["mod_parent.pipeline.refer_to_child"]
	if referToChildPipeline == nil {
		assert.Fail("foo pipeline not found")
		return
	}

	childPipeline := pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child"]
	if childPipeline == nil {
		assert.Fail("this_pipeline_is_in_the_child pipeline not found")
		return
	}

	childPipelineWithVar := pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child_using_variable"]
	if childPipelineWithVar == nil {
		assert.Fail("this_pipeline_is_in_the_child pipeline not found")
		return
	}

	assert.Equal("foo: this is the value of var_one", childPipelineWithVar.Steps[0].(*modconfig.PipelineStepEcho).Text)

	childPipelineWithVarPassedFromParent := pipelines["mod_child_a.pipeline.this_pipeline_is_in_the_child_using_variable_passed_from_parent"]
	if childPipelineWithVarPassedFromParent == nil {
		assert.Fail("this_pipeline_is_in_the_child pipeline not found")
		return
	}

	assert.Equal("foo: var_two from parent .pvars file", childPipelineWithVarPassedFromParent.Steps[0].(*modconfig.PipelineStepEcho).Text)
}

func (suite *FpTestSuite) TestModDependenciesBackwardCompatible() {
	assert := assert.New(suite.T())

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./backward_compatible_mod", []string{".hcl", ".sp"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines

	// TODO: need to fix this
	assert.Equal(11, len(pipelines), "wrong number of pipelines")

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

func (suite *FpTestSuite) TestModVariable() {
	assert := assert.New(suite.T())

	os.Setenv("P_VAR_var_six", "set from env var")

	w, errorAndWarning := workspace.LoadWithParams(suite.ctx, "./mod_variable", []string{".hcl", ".sp"})

	assert.NotNil(w)
	assert.Nil(errorAndWarning.Error)

	mod := w.Mod
	if mod == nil {
		assert.Fail("mod is nil")
		return
	}

	pipelines := mod.ResourceMaps.Pipelines
	pipelineOne := pipelines["test_mod.pipeline.one"]
	if pipelineOne == nil {
		assert.Fail("pipeline one not found")
		return
	}

	assert.Equal("prefix text here and this is the value of var_one and suffix", pipelineOne.Steps[0].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("prefix text here and value from var file and suffix", pipelineOne.Steps[1].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("prefix text here and var_three from var file and suffix", pipelineOne.Steps[2].(*modconfig.PipelineStepEcho).Text)

	assert.True(pipelineOne.Steps[0].IsResolved())
	assert.True(pipelineOne.Steps[1].IsResolved())
	assert.True(pipelineOne.Steps[2].IsResolved())

	// step echo.one_echo should not be resolved, it has reference to echo.one step
	assert.False(pipelineOne.Steps[3].IsResolved())

	assert.Equal("using value from locals: value of locals_one", pipelineOne.Steps[4].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("using value from locals: 10", pipelineOne.Steps[5].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("using value from locals: value of key_two", pipelineOne.Steps[6].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("using value from locals: value of key_two", pipelineOne.Steps[7].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("using value from locals: 33", pipelineOne.Steps[8].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("var_four value is: value from auto.vars file", pipelineOne.Steps[9].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("var_five value is: value from two.auto.vars file", pipelineOne.Steps[10].(*modconfig.PipelineStepEcho).Text)
	assert.Equal("var_six value is: set from env var", pipelineOne.Steps[11].(*modconfig.PipelineStepEcho).Text)

	githubIssuePipeline := pipelines["test_mod.pipeline.github_issue"]
	if githubIssuePipeline == nil {
		assert.Fail("github_issue pipeline not found")
		return
	}

	assert.Equal(1, len(githubIssuePipeline.Params))
	assert.NotNil(githubIssuePipeline.Params["gh_repo"])
	assert.Equal("hello-world", githubIssuePipeline.Params["gh_repo"].Default.AsString())

	githubGetIssueWithNumber := pipelines["test_mod.pipeline.github_get_issue_with_number"]
	if githubGetIssueWithNumber == nil {
		assert.Fail("github_get_issue_with_number pipeline not found")
		return
	}

	assert.Equal(2, len(githubGetIssueWithNumber.Params))
	assert.Equal("cty.String", githubGetIssueWithNumber.Params["github_token"].Type.GoString())
	assert.Equal("cty.Number", githubGetIssueWithNumber.Params["github_issue_number"].Type.GoString())

}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestFpTestSuite(t *testing.T) {
	suite.Run(t, new(FpTestSuite))
}
