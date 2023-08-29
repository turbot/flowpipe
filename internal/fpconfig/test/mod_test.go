package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/pcerr"

	filehelpers "github.com/turbot/go-kit/files"
)

func TestModWithBadTrigger(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/bad_trigger.hcl"},
		})

	_, errorsAndWarnings := pipeparser.LoadModWithFileName("./test_mods", "bad_trigger.hcl", parseCtx)

	if errorsAndWarnings != nil && errorsAndWarnings.Error == nil {
		assert.Fail("should have an error")
		return
	}

	err, ok := errorsAndWarnings.Error.(pcerr.ErrorModel)
	if !ok {
		assert.Fail("should be a pcerr.ErrorModel")
		return
	}

	assert.Equal(pcerr.ErrorCodeDependencyFailure, err.Type, "wrong error type")
}

func TestModLoadFp(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/good_mod.hcl"},
		})

	mod, errorsAndWarnings := pipeparser.LoadModWithFileName("./test_mods", "good_mod.hcl", parseCtx)

	if errorsAndWarnings != nil && errorsAndWarnings.Error != nil {
		assert.Fail("error loading mod file", errorsAndWarnings.Error.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")

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
	assert.Equal("test_mod.trigger.my_hourly_trigger", triggers["test_mod.trigger.my_hourly_trigger"].FullName, "wrong trigger name")
}

func TestModReferences(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/mod_references.hcl"},
		})

	mod, errorsAndWarnings := pipeparser.LoadModWithFileName("./test_mods", "mod_references.hcl", parseCtx)

	if errorsAndWarnings != nil && errorsAndWarnings.Error != nil {
		assert.Fail("error loading mod file", errorsAndWarnings.Error.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")

	// check if all pipelines are there
	pipelines := mod.ResourceMaps.Pipelines
	assert.NotNil(pipelines, "pipelines is nil")
	assert.NotNil(pipelines["pipeline_with_references.pipeline.foo"])
	assert.NotNil(pipelines["pipeline_with_references.pipeline.foo_two"])
}

func TestBadStepReference(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/bad_step_reference.hcl"},
		})

	_, errorsAndWarnings := pipeparser.LoadModWithFileName("./test_mods", "bad_step_reference.hcl", parseCtx)

	if errorsAndWarnings == nil && errorsAndWarnings.Error == nil {
		assert.Fail("should have an error")
		return
	}

	assert.Contains(errorsAndWarnings.Error.Error(), `invalid depends_on 'echozzzz.bar' - step 'echo.baz' does not exist for pipeline pipeline_with_references.pipeline.foo`, "wrong error message")
}

func TestBadStepReferenceTwo(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/bad_step_reference_two.hcl"},
		})

	_, errorsAndWarnings := pipeparser.LoadModWithFileName("./test_mods", "bad_step_reference_two.hcl", parseCtx)

	if errorsAndWarnings == nil && errorsAndWarnings.Error == nil {
		assert.Fail("should have an error")
		return
	}

	assert.Contains(errorsAndWarnings.Error.Error(), `invalid depends_on 'echo.barrs' - step 'echo.baz' does not exist for pipeline pipeline_with_references.pipeline.foo`, "wrong error message")
}

func TestBadPipelineReference(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		"./test_mod/",
		0,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files,
			Include: []string{"**/bad_pipeline_reference.hcl"},
		})

	_, errorsAndWarnings := pipeparser.LoadModWithFileName("./test_mods", "bad_pipeline_reference.hcl", parseCtx)

	if errorsAndWarnings == nil && errorsAndWarnings.Error == nil {
		assert.Fail("should have an error")
		return
	}

	assert.Contains(errorsAndWarnings.Error.Error(), `MISSING: pipeline.foo_two_invalid`, "wrong error message")
}
