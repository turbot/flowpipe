package primitive

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

func TestSleepOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	q := Sleep{}
	input := types.Input(map[string]interface{}{"duration": "1s"})

	output, err := q.Run(ctx, input)
	assert.Nil(err)

	startTime := output.Get("started_at").(time.Time)
	finishTime := output.Get("finished_at").(time.Time)
	diff := finishTime.Sub(startTime)
	assert.Equal(float64(1), math.Floor(diff.Seconds()), "output does not match the provided duration")
}

func TestSleepInvalidDuration(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	q := Sleep{}
	input := types.Input(map[string]interface{}{"duration": "5"})

	_, err := q.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("invalid sleep duration 5", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}
