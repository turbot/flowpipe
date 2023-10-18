package primitive

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/turbot/flowpipe/internal/fplog"
// 	"github.com/turbot/pipe-fittings/modconfig"
// )

// func TestInputStep(t *testing.T) {
// 	ctx := context.Background()
// 	ctx = fplog.ContextWithLogger(ctx)

// 	assert := assert.New(t)
// 	hr := Input{
// 		ExecutionID:         "exec_cknkhj5gdurd7349d4v0",
// 		StepExecutionID:     "sexec_cknkhj5gdurd7349d510",
// 		PipelineExecutionID: "pexec_cknkhj5gdurd7349d4vg",
// 	}

// 	input := modconfig.Input(map[string]interface{}{
// 		"type": InputTypeSlack,
// 	})



// 	_, err := hr.Run(ctx, input)
// 	assert.Nil(err)
// 	// assert.Equal("200 OK", output.Get("status"))
// 	// assert.Equal(200, output.Get("status_code"))
// 	// assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
// 	// assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe")
// }
