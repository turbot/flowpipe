package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextWithSession(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	sessionCtx := ContextWithSession(ctx)
	assert.Implements((*context.Context)(nil), sessionCtx)
	sess := Session(sessionCtx)
	assert.NotEmpty(sess)
}

func TestContextWithSessionID(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	desiredSessionID := "foo"
	sessionCtx := ContextWithSessionID(ctx, desiredSessionID)
	assert.Implements((*context.Context)(nil), sessionCtx)
	actualSessionID := Session(sessionCtx)
	assert.Equal(desiredSessionID, actualSessionID)
}

func TestParallelContextWithSession(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	sessionCtx1 := ContextWithSession(ctx)
	sessionCtx2 := ContextWithSession(ctx)
	sess1 := Session(sessionCtx1)
	sess2 := Session(sessionCtx2)
	assert.NotEmpty(sess1)
	assert.NotEmpty(sess2)
	assert.NotEqual(sess1, sess2)
}
