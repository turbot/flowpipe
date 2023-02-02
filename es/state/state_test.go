package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	assert := assert.New(t)
	runID := "cc14106v9mc75ace8ocg"
	s, err := NewState(runID)
	assert.Nil(err)
	assert.Equal(runID, s.RunID)
	assert.Equal("e-gineer", s.IdentityID)
	assert.Equal("scratch", s.WorkspaceID)
}
