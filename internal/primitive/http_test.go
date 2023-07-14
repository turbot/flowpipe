package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

func TestHTTPRequestOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{"url": "https://steampipe.io/"})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "Steampipe")
}

func TestHTTPRequestJSONResponseOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{"url": "http://api.open-notify.org/astros.json"})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "success")
}

func TestHTTPRequestNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{"url": "https://steampipe.io/asdlkfjasdlfkjnotfound/"})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "Steampipe")
}
