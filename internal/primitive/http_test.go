package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

// GET

func TestHTTPMethodGET(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl: "https://steampipe.io/",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe")
}

func TestHTTPMethodGETWithQueryString(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl: "https://hub.steampipe.io/plugins?categories=saas",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe Cloud")
}

func TestHTTPMethodGETWithJSONResponse(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://api.open-notify.org/astros.json",
		schema.AttributeTypeMethod: modconfig.HttpMethodGet,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody).(map[string]interface{})["message"], "success")
}

func TestHTTPMethodGETNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://steampipe.io/asdlkfjasdlfkjnotfound/",
		schema.AttributeTypeMethod: modconfig.HttpMethodGet,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(404, e.Error.Status)
		assert.Equal("404 Not Found", e.Error.Detail)
	}
	assert.Equal("text/html; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "Steampipe")
}

func TestHTTPMethodGETUnauthorized(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://cloud.steampipe.io/api/v0/user/flowpipe/connection",
		schema.AttributeTypeMethod: modconfig.HttpMethodGet,
		schema.AttributeTypeRequestBody: `{
			"Authorization": "Bearer spt_flo3pipe00g0t1nvali_3test0axy78ic8h6http77o24"
		}`,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(401, e.Error.Status)
		assert.Equal("401 Unauthorized", e.Error.Detail)
	}
}

// POST

func TestHTTPMethodPOST(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
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
	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:         "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod:      modconfig.HttpMethodPost,
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

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(404, e.Error.Status)
		assert.Equal("404 Not Found", e.Error.Detail)
	}
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}

func TestHTTPMethodPOSTWithRequestHeaders(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
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

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
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

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
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

func TestHTTPMethodPOSTWithVerifyInvalidCertificate(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
		schema.AttributeTypeRequestBody: `{
			"userId": 1001,
			"it": 1001,
			"title": "Test 1001"
		}`,
		schema.AttributeTypeCaCertPem: "test",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)
	assert.Contains(err.Error(), "tls: failed to verify certificate")
}

func TestHTTPMethodPOSTWithVerifyCertificate(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts",
		schema.AttributeTypeMethod: modconfig.HttpMethodPost,
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

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod: modconfig.HttpMethodDelete,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}

func TestHTTPMethodDELETENotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: modconfig.HttpMethodDelete,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(404, e.Error.Status)
		assert.Equal("404 Not Found", e.Error.Detail)
	}
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}

// PUT

func TestHTTPMethodPUT(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod: modconfig.HttpMethodPut,
		schema.AttributeTypeRequestHeaders: map[string]interface{}{
			"Content-Type": "application/json",
		},
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
	assert.Equal(output.Get(schema.AttributeTypeResponseBody).(map[string]interface{})["body"], "bar")
	assert.Equal(output.Get(schema.AttributeTypeResponseBody).(map[string]interface{})["id"], float64(1))
	assert.Equal(output.Get(schema.AttributeTypeResponseBody).(map[string]interface{})["title"], "foo")
}

func TestHTTPMethodPUTWithTextBody(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:         "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod:      modconfig.HttpMethodPut,
		schema.AttributeTypeRequestBody: "test",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("application/json; charset=utf-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
	assert.Contains(output.Get(schema.AttributeTypeResponseBody), "id")
}

func TestHTTPMethodPUTNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: modconfig.HttpMethodPut,
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(404, e.Error.Status)
		assert.Equal("404 Not Found", e.Error.Detail)
	}
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}

// PATCH

func TestHTTPMethodPATCH(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}
	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod: modconfig.HttpMethodPatch,
		schema.AttributeTypeRequestHeaders: map[string]interface{}{
			"Content-Type": "application/json",
		},
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
	assert.Equal(output.Get(schema.AttributeTypeResponseBody).(map[string]interface{})["body"], "Updating the body of the target resource")
}

func TestHTTPMethodPATCHWithTextBody(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := HTTPRequest{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:         "https://jsonplaceholder.typicode.com/posts/1",
		schema.AttributeTypeMethod:      modconfig.HttpMethodPatch,
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

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeUrl:    "http://www.example.com/notfound",
		schema.AttributeTypeMethod: modconfig.HttpMethodPatch})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	output.HasErrors()
	for _, e := range output.Errors {
		assert.Equal(404, e.Error.Status)
		assert.Equal("404 Not Found", e.Error.Detail)
	}
	assert.Equal("text/html; charset=UTF-8", output.Get(schema.AttributeTypeResponseHeaders).(map[string]interface{})["Content-Type"])
}
