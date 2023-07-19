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
	input := types.Input(map[string]interface{}{"url": "https://steampipe.io/", "method": "get"})
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
	input := types.Input(map[string]interface{}{"url": "http://api.open-notify.org/astros.json", "method": "get"})
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
	input := types.Input(map[string]interface{}{"url": "https://steampipe.io/asdlkfjasdlfkjnotfound/", "method": "get"})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "Steampipe")
}

func TestHTTPPOSTRequestOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		"url":    "https://jsonplaceholder.typicode.com/posts",
		"method": "post",
		"body": `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "id")
}

func TestHTTPPOSTRequestOKWithTextBody(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		"url":    "https://jsonplaceholder.typicode.com/posts",
		"method": "post",
		"body":   "Test",
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "id")
}

func TestHTTPPOSTRequestOKWithRequestHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		"url":    "https://jsonplaceholder.typicode.com/posts",
		"method": "post",
		"body": `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		"request_headers": map[string]interface{}{
			"Accept":     "application/json",
			"User-Agent": "flowpipe",
		},
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "id")
}

func TestHTTPPOSTRequestOKWithTimeout(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		"url":    "https://jsonplaceholder.typicode.com/posts",
		"method": "post",
		"body": `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		"request_timeout_ms": 3000,
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "id")
}

func TestHTTPPOSTRequestOKWithNoVerifyCertificate(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		"url":    "https://jsonplaceholder.typicode.com/posts",
		"method": "post",
		"body": `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		"insecure":    true,
		"ca_cert_pem": "test",
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get("body"), "id")
}

func TestHTTPPOSTRequestWithVerifyCertificate(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		"url":    "https://jsonplaceholder.typicode.com/posts",
		"method": "post",
		"body": `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		"request_timeout_ms": 3000,
		"insecure":           false,
		"ca_cert_pem":        "test",
	})
	_, err := hr.Run(ctx, input)
	assert.NotNil(err, "no error found")
	assert.Contains(err.Error(), "unknown authority")
}
