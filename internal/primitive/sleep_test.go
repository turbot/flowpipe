package primitive

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

func TestSleepOK(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := flowpipe.Input(map[string]interface{}{"duration": "1s"})

	output, err := q.Run(ctx, input)
	assert.Nil(err)

	flowpipeMetadata := output.Flowpipe
	startTime := flowpipeMetadata[schema.AttributeTypeStartedAt].(time.Time)
	finishTime := flowpipeMetadata[schema.AttributeTypeFinishedAt].(time.Time)
	diff := finishTime.Sub(startTime)
	assert.Equal(float64(1), math.Floor(diff.Seconds()), "output does not match the provided duration")
}

func TestSleepWithDurationInNumber(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := flowpipe.Input(map[string]interface{}{"duration": int64(1000)})

	output, err := q.Run(ctx, input)
	assert.Nil(err)

	flowpipeMetadata := output.Flowpipe
	startTime := flowpipeMetadata[schema.AttributeTypeStartedAt].(time.Time)
	finishTime := flowpipeMetadata[schema.AttributeTypeFinishedAt].(time.Time)

	diff := finishTime.Sub(startTime)
	assert.Equal(float64(1), math.Floor(diff.Seconds()), "output does not match the provided duration")
}

func TestSleepInvalidDuration(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := flowpipe.Input(map[string]interface{}{"duration": "5"})

	_, err := q.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("invalid sleep duration 5", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestSleepNegativeDuration(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := flowpipe.Input(map[string]interface{}{"duration": int64(-1)})

	_, err := q.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("The attribute 'duration' must be a positive whole number", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}
