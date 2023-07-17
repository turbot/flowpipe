package pipeline

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/turbot/flowpipe/internal/fplog"
// )

// func TestSimpleForAndParam(t *testing.T) {
// 	assert := assert.New(t)

// 	ctx := context.Background()
// 	ctx = fplog.ContextWithLogger(ctx)

// 	pipelines, err := LoadPipelines(ctx, "./test_pipelines/for.fp")
// 	assert.Nil(err, "error found")

// 	assert.GreaterOrEqual(len(pipelines), 1, "wrong number of pipelines")

// 	if pipelines["for_loop"] == nil {
// 		assert.Fail("for_loop pipeline not found")
// 		return
// 	}
// }
