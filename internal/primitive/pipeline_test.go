package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

func TestPipelineOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	q := RunPipeline{}
	input := types.Input(map[string]interface{}{"pipeline": "my_pipeline", "args": map[string]interface{}{}})

	output, err := q.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("my_pipeline", output.Get("pipeline").(string), "wrong pipeline name")
}
