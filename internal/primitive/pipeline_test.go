package primitive

import (
	"context"
	"github.com/turbot/flowpipe/internal/resources"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineOK(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := RunPipeline{}
	input := resources.Input(map[string]interface{}{"pipeline": "my_pipeline", "args": map[string]interface{}{}})

	output, err := q.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("my_pipeline", output.Get("pipeline").(string), "wrong pipeline name")
}
