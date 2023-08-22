package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/parse"

	"github.com/turbot/go-kit/files"
	filehelpers "github.com/turbot/go-kit/files"
)

func SkipTestModFileLoad(t *testing.T) {
	assert := assert.New(t)

	mod, err := parse.LoadModfile("./test_mod/")

	if err != nil {
		assert.Fail("error loading mod file", err.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")
}

func SkipTestModLoadSp(t *testing.T) {
	assert := assert.New(t)

	parseCtx := parse.NewModParseContext(
		context.TODO(),
		nil,
		"./test_steampipe_mod/",
		parse.CreateDefaultMod,
		&filehelpers.ListOptions{
			// // listFlag specifies whether to load files recursively
			Flags: files.Files | files.Recursive,
			// Exclude: w.exclusions,
			// only load .sp files
			Include: filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension}),
		})

	mod, err := pipeparser.LoadMod("./test_steampipe_mod/", parseCtx)

	if err != nil {
		assert.Fail("error loading mod file", err.Error.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")
}

func TestModLoadFp(t *testing.T) {
	assert := assert.New(t)

	parseCtx := parse.NewModParseContext(
		context.TODO(),
		nil,
		"./test_mod/",
		parse.CreateDefaultMod,
		&filehelpers.ListOptions{
			// // listFlag specifies whether to load files recursively
			Flags: files.Files | files.Recursive,
			// Exclude: w.exclusions,
			// only load .sp files
			Include: filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension}),
		})

	mod, errorsAndWarnings := pipeparser.LoadMod("./test_mod/", parseCtx)

	if errorsAndWarnings != nil && errorsAndWarnings.Error != nil {
		assert.Fail("error loading mod file", errorsAndWarnings.Error.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")
}
