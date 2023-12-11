package primitive

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func TestSleepOK(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := modconfig.Input(map[string]interface{}{"duration": "1s"})

	output, err := q.Run(ctx, input)
	assert.Nil(err)

	startTime := output.Get("started_at").(time.Time)
	finishTime := output.Get("finished_at").(time.Time)
	diff := finishTime.Sub(startTime)
	assert.Equal(float64(1), math.Floor(diff.Seconds()), "output does not match the provided duration")
}

func TestSleepWithDurationInNumber(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := modconfig.Input(map[string]interface{}{"duration": int64(1000)})

	output, err := q.Run(ctx, input)
	assert.Nil(err)

	startTime := output.Get("started_at").(time.Time)
	finishTime := output.Get("finished_at").(time.Time)
	diff := finishTime.Sub(startTime)
	assert.Equal(float64(1), math.Floor(diff.Seconds()), "output does not match the provided duration")
}

func TestSleepInvalidDuration(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	q := Sleep{}
	input := modconfig.Input(map[string]interface{}{"duration": "5"})

	_, err := q.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("invalid sleep duration 5", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}
