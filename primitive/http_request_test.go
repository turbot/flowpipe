package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

func TestHTTPRequestOK(t *testing.T) {
	assert := assert.New(t)
	hr := HTTPRequest{}
	input := pipeline.Input(map[string]interface{}{"url": "https://steampipe.io/"})
	output, err := hr.Run(context.Background(), input)
	assert.Nil(err)
	assert.Equal("200 OK", output.Get("status"))
	assert.Equal(200, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	//fmt.Println(output.Get("headers"))
	assert.Contains(output.Get("body"), "Steampipe")
}

func TestHTTPRequestNotFound(t *testing.T) {
	assert := assert.New(t)
	hr := HTTPRequest{}
	input := pipeline.Input(map[string]interface{}{"url": "https://steampipe.io/asdlkfjasdlfkjnotfound/"})
	output, err := hr.Run(context.Background(), input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output.Get("status"))
	assert.Equal(404, output.Get("status_code"))
	assert.Equal("text/html; charset=utf-8", output.Get("headers").(map[string]interface{})["Content-Type"])
	//fmt.Println(output.Get("headers"))
	assert.Contains(output.Get("body"), "Steampipe")
}
