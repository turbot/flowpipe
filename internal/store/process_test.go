package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadExecutionFromDB(t *testing.T) {
	assert := assert.New(t)

	excutionIDs, err := ListExecutionIDs()
	assert.Nil(err)
	assert.Equal(4, len(excutionIDs))
	assert.Equal("exec_cmsp5a272ijn3jbg6850", excutionIDs[3])
	assert.Equal("exec_cmspbei72ijj85acej90", excutionIDs[2])
	assert.Equal("exec_cmspbeq72ijj85acejfg", excutionIDs[1])
	assert.Equal("exec_cmspbf272ijj85acejm0", excutionIDs[0])
}
