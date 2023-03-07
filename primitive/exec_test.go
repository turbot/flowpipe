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
	// No errors
	assert.Nil(err)
	// We only return *_lines fields
	assert.Nil(output["stdout"])
	assert.Nil(output["stderr"])
	// Check stdout
	assert.NotNil(output["stdout_lines"])
	assert.Equal("test", output["stdout_lines"].([]string)[0])
	// Check stderr
	assert.NotNil(output["stderr_lines"])
	assert.Empty(output["stderr_lines"].([]string))
	// Check exit code
	assert.Equal(0, output["exit_code"])
}

func TestExecProgramNotFound(t *testing.T) {
	assert := assert.New(t)
	hr := Exec{}
	input := Input(map[string]interface{}{"command": "my_non_existent_cli 'test'"})
	output, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)
	// Check stdout
	assert.NotNil(output["stdout_lines"])
	assert.Empty(output["stdout_lines"].([]string))
	// Check stderr
	assert.NotNil(output["stderr_lines"])
	assert.Contains(output["stderr_lines"].([]string)[0], "command not found")
	// Check exit code
	assert.Equal(127, output["exit_code"])
}
