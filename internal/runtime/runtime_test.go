package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeAvailable(t *testing.T) {
	assert := assert.New(t)

	runtimes, err := RuntimesAvailable()

	assert.NoError(err)
	assert.Contains(runtimes, "nodejs:18")
	assert.Contains(runtimes, "nodejs:20")
	assert.Contains(runtimes, "python:3.10")
}
