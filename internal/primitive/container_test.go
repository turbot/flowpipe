package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

func initializeCocker() {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)
	logger := fplog.Logger(ctx)

	// Start MailHog server as a separate process
	logger.Debug("Initializing Docker")

	err := docker.Initialize(ctx)
	if err != nil {
		logger.Error("Failed to start MailHog: ", err.Error())
	}
	logger.Debug("Docker initialized")
}

func TestSimpleContainerStep(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Container{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeImage:             "alpine:3.7",
		schema.AttributeTypeCmd:               []interface{}{"echo", "hello world"},
		schema.AttributeTypeEnv:               map[string]interface{}{"FOO": "bar"},
		schema.AttributeTypeTimeout:           int64(120),
		schema.LabelName:                      "container_test",
		schema.AttributeTypeMemory:            int64(128),
		schema.AttributeTypeMemoryReservation: int64(64),
		schema.AttributeTypeMemorySwap:        int64(256),
		schema.AttributeTypeMemorySwappiness:  int64(10),
		schema.AttributeTypeReadOnly:          false,
		schema.AttributeTypeUser:              "root",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(0, len(output.Errors))
	assert.NotNil(output.Get("container_id"))
	assert.Equal(0, output.Get("exit_code"))
	assert.Contains(output.Get("stdout"), "hello world")
	assert.Equal("", output.Get("stderr"))
}
