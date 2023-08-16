package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser/parse"
)

func TestModLoad(t *testing.T) {
	assert := assert.New(t)

	mod, err := parse.LoadModfile("./test_mod/")

	if err != nil {
		assert.Fail("error loading mod file", err.Error())
		return
	}

	assert.NotNil(mod, "mod is nil")
}
