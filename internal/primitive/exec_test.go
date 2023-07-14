package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/types"
)

func TestExecOK(t *testing.T) {
	assert := assert.New(t)
	hr := Exec{}
	input := types.Input(map[string]interface{}{"command": "echo 'test'"})
	output, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)
	// We only return *_lines fields
	assert.Nil(output.Get("stdout"))
	assert.Nil(output.Get("stderr"))
	// Check stdout
	assert.NotNil(output.Get("stdout_lines"))
	assert.Equal("test", output.Get("stdout_lines").([]string)[0])
	// Check stderr
	assert.NotNil(output.Get("stderr_lines"))
	assert.Empty(output.Get("stderr_lines").([]string))
	// Check exit code
	assert.Equal(0, output.Get("exit_code"))
}

func TestExecProgramNotFound(t *testing.T) {
	assert := assert.New(t)
	hr := Exec{}
	input := types.Input(map[string]interface{}{"command": "my_non_existent_cli 'test'"})
	output, err := hr.Run(context.Background(), input)
	// No errors
	assert.Nil(err)
	// Check stdout
	assert.NotNil(output.Get("stdout_lines"))
	assert.Empty(output.Get("stdout_lines").([]string))
	// Check stderr
	assert.NotNil(output.Get("stderr_lines"))
	assert.Contains(output.Get("stderr_lines").([]string)[0], "my_non_existent_cli: not found")
	// Check exit code
	assert.Equal(127, output.Get("exit_code"))
}
