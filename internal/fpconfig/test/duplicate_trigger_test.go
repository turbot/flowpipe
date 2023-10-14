package pipeline_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/pipeparser"
)

func TestDuplicateTriggers(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/duplicate_triggers.fp")

	if err == nil {
		assert.Fail("expected error not found")
		return
	}

	assert.Contains(err.Error(), "duplicate unresolved block name 'trigger.my_hourly_trigger'")
}

func TestDuplicateTriggersDiffPipeline(t *testing.T) {
	assert := assert.New(t)

	_, _, err := pipeparser.LoadPipelines(context.TODO(), "./test_pipelines/invalid_pipelines/duplicate_triggers_diff_pipeline.fp")

	if err == nil {
		assert.Fail("expected error not found")
		return
	}

	assert.Contains(err.Error(), "Mod defines more than one resource named 'local.trigger.schedule.my_hourly_trigger'")
}
