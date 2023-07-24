package primitive

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

// GET

func TestHTTPRequestOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl: "https://steampipe.io/",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe")
}

func TestHTTPRequestJSONResponseOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://api.open-notify.org/astros.json",
		schema.AttributeTypeMethod: types.HttpMethodGet,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "success")
}

func TestHTTPRequestNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://steampipe.io/asdlkfjasdlfkjnotfound/",
		schema.AttributeTypeMethod: types.HttpMethodGet,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe")
}

// POST

func TestHTTPPOSTRequestOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: types.HttpMethodPost,
		schema.AttributeTypeRequestBody: `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get(schema.AttributeTypeStatus))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPPOSTRequestOKWithTextBody(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:         "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod:      types.HttpMethodPost,
		schema.AttributeTypeRequestBody: "Test",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPPOSTRequestNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: types.HttpMethodPost,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}

func TestHTTPPOSTRequestOKWithRequestHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: types.HttpMethodPost,
		schema.AttributeTypeRequestBody: `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		"request_headers": map[string]interface{}{
			"Accept":       "*/*",
			"User-Agent":   "flowpipe",
			"Content-Type": "application/json",
		},
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])

	// TODO: check for body_json
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPPOSTRequestOKWithTimeout(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: types.HttpMethodPost,
		schema.AttributeTypeRequestBody: `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		schema.AttributeTypeRequestTimeoutMs: 3000,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPPOSTRequestOKWithNoVerifyCertificate(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: types.HttpMethodPost,
		schema.AttributeTypeRequestBody: `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		schema.AttributeTypeInsecure:  true,
		schema.AttributeTypeCaCertPem: "test",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("201 Created", output.Get("status"))
	assert.Equal(201, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPPOSTRequestWithVerifyCertificate(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: types.HttpMethodPost,
		schema.AttributeTypeRequestBody: `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		schema.AttributeTypeRequestTimeoutMs: 3000,
		schema.AttributeTypeCaCertPem:        "test",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err, "no error found")
	assert.Contains(err.Error(), "unknown authority")
}

// DELETE

func TestHTTPDELETERequestOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod: types.HttpMethodDelete,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Equal(output.Get(schema.AttributeTypeResponseBody), "{}")
}

func TestHTTPDELETERequestNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: types.HttpMethodDelete,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}

// PUT

func TestHTTPPUTRequestOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod: types.HttpMethodPut,
		schema.AttributeTypeRequestBody: `{
				"id": 1,
				"title": "foo",
				"body": "bar",
				"userId": 1
			}`,
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "foo")
	reflect.DeepEqual(output.Get(schema.AttributeTypeResponseBodyJson), map[string]interface{}{"body": "bar", "id": 1, "title": "foo", "userId": 1})
}

func TestHTTPPUTRequestWithTextBodyOK(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:         "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod:      types.HttpMethodPut,
		schema.AttributeTypeRequestBody: "test",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
	reflect.DeepEqual(output.Get(schema.AttributeTypeResponseBodyJson), map[string]interface{}{"id": 1})
}

func TestHTTPPUTRequestNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: types.HttpMethodPut})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}
