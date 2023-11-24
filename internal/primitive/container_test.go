package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/container"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
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
		schema.AttributeTypeCmd:               []interface{}{"sh", "-c", "echo 'Line 1'; echo 'Line 2'; echo 'Line 3'"},
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
	assert.Equal("Line 1\nLine 2\nLine 3\n", output.Get("stdout"))
	assert.Equal("", output.Get("stderr"))

	assert.NotNil(output.Get("lines"))

	if _, ok := output.Get("lines").([]container.OutputLine); !ok {
		assert.Fail("Expected lines to be []container.OutputLine")
	}
	lines := output.Get("lines").([]container.OutputLine)
	assert.Equal(3, len(lines))

	assert.Equal("stdout", lines[0].Stream)
	assert.Equal("Line 1\n", lines[0].Line)

	assert.Equal("stdout", lines[1].Stream)
	assert.Equal("Line 2\n", lines[1].Line)

	assert.Equal("stdout", lines[2].Stream)
	assert.Equal("Line 3\n", lines[2].Line)
}

func TestContainerStepMissingImage(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Container{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeCmd:     []interface{}{"echo", "hello world"},
		schema.AttributeTypeEnv:     map[string]interface{}{"FOO": "bar"},
		schema.AttributeTypeTimeout: int64(120),
		schema.LabelName:            "container_test",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Container input must define 'image'", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestContainerStepInvalidImage(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Container{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeImage:   123,
		schema.AttributeTypeCmd:     []interface{}{"echo", "hello world"},
		schema.AttributeTypeEnv:     map[string]interface{}{"FOO": "bar"},
		schema.AttributeTypeTimeout: int64(120),
		schema.LabelName:            "container_test",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Container attribute 'image' must be a string", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestContainerStepInvalidMemory(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Container{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeImage:   "alpine:3.7",
		schema.AttributeTypeCmd:     []interface{}{"echo", "hello world"},
		schema.AttributeTypeEnv:     map[string]interface{}{"FOO": "bar"},
		schema.AttributeTypeTimeout: int64(120),
		schema.LabelName:            "container_test",
		schema.AttributeTypeMemory:  int64(1),
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)

	output.HasErrors()
	assert.Equal(1, len(output.Errors))
	assert.Contains(output.Errors[0].Error.Detail, "Minimum memory limit allowed is 6MB")
	assert.Equal(500, output.Errors[0].Error.Status)
}

func TestContainerStepInvalidCmd(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Container{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeImage:   "alpine:3.7",
		schema.AttributeTypeCmd:     "echo",
		schema.AttributeTypeEnv:     map[string]interface{}{"FOO": "bar"},
		schema.AttributeTypeTimeout: int64(120),
		schema.LabelName:            "container_test",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Container attribute 'cmd' must be an array of strings", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestContainerStepInvalidEntrypoint(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Container{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeImage:      "alpine:3.7",
		schema.AttributeTypeEntryPoint: "echo",
		schema.AttributeTypeEnv:        map[string]interface{}{"FOO": "bar"},
		schema.AttributeTypeTimeout:    int64(120),
		schema.LabelName:               "container_test",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Container attribute 'entrypoint' must be an array of strings", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}
