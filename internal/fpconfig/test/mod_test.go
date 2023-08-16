package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/parse"

	filehelpers "github.com/turbot/go-kit/files"
)

func TestModFileLoad(t *testing.T) {
	assert := assert.New(t)

	mod, err := parse.LoadModfile("./test_mod/")

	if err != nil {
		assert.Fail("error loading mod file", err.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")
}

func SkipTestModLoad(t *testing.T) {
	assert := assert.New(t)

	parseCtx := parse.NewModParseContext(
		nil,
		"./test_mod/",
		parse.CreateDefaultMod,
		&filehelpers.ListOptions{
			// // listFlag specifies whether to load files recursively
			// Flags:   w.listFlag,
			// Exclude: w.exclusions,
			// only load .sp files
			Include: filehelpers.InclusionsFromExtensions([]string{constants.ModDataExtension}),
		})

	mod, err := pipeparser.LoadMod("./test_mod/", parseCtx)

	if err != nil {
		assert.Fail("error loading mod file", err.Error.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")
}
