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
		ExecutionID: "exec_cmsp5a272ijn3jbg6850",
	}

	ex, err := NewExecution(ctx)
	if err != nil {
		assert.Fail("Error creating execution", err)
		return
	}

	err = ex.LoadProcessDB(evt)
	assert.Nil(err)
	assert.Equal(1, len(ex.PipelineExecutions))

	pe := ex.PipelineExecutions["pexec_cmsp5a272ijn3jbg685g"]

	assert.Equal("pexec_cmsp5a272ijn3jbg685g", pe.ID)
	assert.Equal(20, len(pe.PipelineOutput["inserted_rows"].([]interface{})))
}
