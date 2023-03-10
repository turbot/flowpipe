package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigInContext(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	cfgIn, err := NewConfig(WithLogDir("/tmp"))
	assert.Nil(err)
	configCtx := Set(ctx, cfgIn)
	assert.Implements((*context.Context)(nil), configCtx)
	cfgOut := Get(configCtx)
	assert.NotEmpty(cfgOut)
	assert.Equal(cfgIn, cfgOut)
	assert.Equal(cfgIn.LogDir, cfgOut.LogDir)
}
