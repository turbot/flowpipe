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

func TestHTTPMethodGET(t *testing.T) {
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

func TestHTTPMethodGETWithJSONResponse(t *testing.T) {
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

func TestHTTPMethodGETNotFound(t *testing.T) {
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

func TestHTTPMethodPOST(t *testing.T) {
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

func TestHTTPMethodPOSTWithTextBody(t *testing.T) {
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

func TestHTTPMethodPOSTNotFound(t *testing.T) {
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

func TestHTTPMethodPOSTWithRequestHeaders(t *testing.T) {
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

func TestHTTPMethodPOSTWithTimeout(t *testing.T) {
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

func TestHTTPMethodPOSTWithNoVerifyCertificate(t *testing.T) {
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

func TestHTTPMethodPOSTWithVerifyCertificate(t *testing.T) {
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

func TestHTTPMethodDELETE(t *testing.T) {
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

func TestHTTPMethodDELETENotFound(t *testing.T) {
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

func TestHTTPMethodPUT(t *testing.T) {
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

func TestHTTPMethodPUTWithTextBody(t *testing.T) {
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

func TestHTTPMethodPUTNotFound(t *testing.T) {
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

// PATCH

func TestHTTPMethodPATCH(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod: types.HttpMethodPatch,
		schema.AttributeTypeRequestBody: `{
			"title": "foo",
			"body": "Updating the body of the target resource"
		}`,
	})
	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Updating the body of the target resource")
	reflect.DeepEqual(output.Get(schema.AttributeTypeResponseBodyJson), map[string]interface{}{"body": "Updating the body of the target resource", "id": 1, "title": "foo", "userId": 1})
}

func TestHTTPMethodPATCHWithTextBody(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:         "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod:      types.HttpMethodPatch,
		schema.AttributeTypeRequestBody: "test",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPMethodPATCHNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: types.HttpMethodPatch})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}
