package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineOK(t *testing.T) {

	assert := assert.New(t)

	cron, err := IntervalToCronExpression("abc1234", "5m")
	assert.Nil(err)
	assert.Equal("0-59/5 * * * *", cron)

	cron, err = IntervalToCronExpression("abc1237", "5m")
	assert.Nil(err)
	assert.Equal("1-59/5 * * * *", cron)

	cron, err = IntervalToCronExpression("abc1237", "30m")
	assert.Nil(err)
	assert.Equal("7-59/30 * * * *", cron)

	cron, err = IntervalToCronExpression("abc1237", "2h")
	assert.Nil(err)
	assert.Equal("15 0-23/2 * * *", cron)

	cron, err = IntervalToCronExpression("aaaaaaabbbbbbb", "2h")
	assert.Nil(err)
	assert.Equal("2 0-23/2 * * *", cron)

	cron, err = IntervalToCronExpression("aaaaaaabbbbbbb", "4h")
	assert.Nil(err)
	assert.Equal("2 0-23/4 * * *", cron)

	cron, err = IntervalToCronExpression("aaaaccccbbb", "4h")
	assert.Nil(err)
	assert.Equal("53 2-23/4 * * *", cron)

	cron, err = IntervalToCronExpression("aaaaccccbbbeeee", "4h")
	assert.Nil(err)
	assert.Equal("31 1-23/4 * * *", cron)
}
