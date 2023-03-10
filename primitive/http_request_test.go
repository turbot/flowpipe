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
	input := pipeline.StepInput(map[string]interface{}{"url": "https://steampipe.io/"})
	output, err := hr.Run(context.Background(), input)
	assert.Nil(err)
	assert.Equal("200 OK", output["status"])
	assert.Equal(200, output["status_code"])
	assert.Equal("text/html; charset=utf-8", output["headers"].(map[string]interface{})["Content-Type"])
	//fmt.Println(output["headers"])
	assert.Contains(output["body"], "Steampipe")
}

func TestHTTPRequestNotFound(t *testing.T) {
	assert := assert.New(t)
	hr := HTTPRequest{}
	input := pipeline.StepInput(map[string]interface{}{"url": "https://steampipe.io/asdlkfjasdlfkjnotfound/"})
	output, err := hr.Run(context.Background(), input)
	assert.Nil(err)
	assert.Equal("404 Not Found", output["status"])
	assert.Equal(404, output["status_code"])
	assert.Equal("text/html; charset=utf-8", output["headers"].(map[string]interface{})["Content-Type"])
	//fmt.Println(output["headers"])
	assert.Contains(output["body"], "Steampipe")
}
