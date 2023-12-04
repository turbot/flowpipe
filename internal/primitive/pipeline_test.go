package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
)

func TestPipelineOK(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := RunPipeline{}
	input := modconfig.Input(map[string]interface{}{"pipeline": "my_pipeline", "args": map[string]interface{}{}})

	output, err := q.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("my_pipeline", output.Get("pipeline").(string), "wrong pipeline name")
}
