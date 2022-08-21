package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecOK(t *testing.T) {
	assert := assert.New(t)
	hr := Exec{}
	input := Input(map[string]interface{}{"command": "echo 'test'"})
	output, err := hr.Run(context.Background(), input)
	assert.Nil(err)
	assert.Equal("test", output["stdout"])
	assert.Equal("", output["stderr"])
	assert.Equal(0, output["exit_code"])
}

func TestExecProgramNotFound(t *testing.T) {
	assert := assert.New(t)
	hr := Exec{}
	input := Input(map[string]interface{}{"command": "my_non_existent_cli 'test'"})
	output, err := hr.Run(context.Background(), input)
	assert.Nil(err)
	assert.Equal("", output["stdout"])
	assert.Contains(output["stderr"], "command not found")
	assert.Equal(127, output["exit_code"])
}
