package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)
	cfg, err := NewConfig()
	assert.Nil(err)
	assert.Equal(cfg.LogDir, "logs")
}
