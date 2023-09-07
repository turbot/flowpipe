package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/perr"

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

	err, ok := errorsAndWarnings.Error.(perr.ErrorModel)
	if !ok {
		assert.Fail("should be a pcerr.ErrorModel")
		return
	}

	assert.Equal(perr.ErrorCodeDependencyFailure, err.Type, "wrong error type")
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
