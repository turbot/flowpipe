package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := NewConfig()
	assert.Equal(cfg.LogDir, "logs")
}
