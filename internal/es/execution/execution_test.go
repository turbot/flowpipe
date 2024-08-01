package execution

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/es/event"
)

func TestExecutionLoadFromDB(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	evt := &event.Event{
		ExecutionID: "exec_cqlecr4204vm48hs8lp0",
	}

	ex, err := NewExecution(ctx, WithEvent(evt))
	if err != nil {
		assert.FailNow("Error creating execution", err)
	}

	assert.Equal(1, len(ex.PipelineExecutions))

	pe := ex.PipelineExecutions["pexec_cqlecr4204vm48hs8lpg"]

	assert.Equal("pexec_cqlecr4204vm48hs8lpg", pe.ID)
}
