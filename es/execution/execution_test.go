package execution

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/util"
)

func TestNewExecution(t *testing.T) {
	assert := assert.New(t)
	cfg, err := config.NewConfig(config.WithLogDir("es/state/tests"))
	assert.Nil(err)
	sessionID := "foo"
	ctx := context.Background()
	ctx = util.ContextWithSessionID(ctx, sessionID)
	ctx = config.Set(ctx, cfg)
	ex, err := NewExecution(ctx)
	assert.Nil(err)
	assert.NotEmpty(ex)
}

func TestLoadJSON(t *testing.T) {
	assert := assert.New(t)
	ex, err := NewExecution(context.Background())
	assert.Nil(err)
	assert.NotEmpty(ex)
	err = ex.LoadJSON("tests/test-load-execution.json")
	assert.Nil(err)
	assert.NotEmpty(ex.ID)
}

func TestLoadJSONNotFound(t *testing.T) {
	assert := assert.New(t)
	ex, err := NewExecution(context.Background())
	assert.Nil(err)
	assert.NotEmpty(ex)
	err = ex.LoadJSON("tests/file-does-not-exist.json")
	assert.NotNil(err)
	assert.ErrorContains(err, "no such file or directory")
}

func TestExecutionLoad(t *testing.T) {
	assert := assert.New(t)
	cfg, err := config.NewConfig(config.WithLogDir("tests"))
	assert.Nil(err)
	// Setup the session context
	sessionID := "foo"
	ctx := context.Background()
	ctx = util.ContextWithSessionID(ctx, sessionID)
	ctx = config.Set(ctx, cfg)
	// Setup the execution
	e := &event.Event{
		ExecutionID: "test-load",
	}
	ex, err := NewExecution(ctx, WithID(e.ExecutionID))
	assert.Nil(err)
	assert.NotEmpty(ex)
	// Load the execution
	err = ex.LoadProcess(e)
	assert.Nil(err)
	expectedExecution, err := NewExecution(ctx)
	assert.Nil(err)
	err = expectedExecution.LoadJSON("tests/test-load-execution.json")
	assert.Nil(err)
	assert.Equal(expectedExecution, ex)
}
