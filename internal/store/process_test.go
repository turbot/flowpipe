package store

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadExecutionFromDB(t *testing.T) {
	assert := assert.New(t)

	// Open the source file
	sourceFile, err := os.Open("flowpipe_clean.db")
	if err != nil {
		assert.Fail("error opening source file", "error", err)
		return
	}
	defer sourceFile.Close()

	// Create the destination file, this will overwrite if file already exists
	destFile, err := os.Create("flowpipe.db")
	if err != nil {
		assert.Fail("error creating destination file", "error", err)
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		assert.Fail("error copying file", "error", err)
	}

	// Force to flush the file system's in-memory copy of recently written data to disk.
	err = destFile.Sync()
	if err != nil {
		assert.Fail("error syncing file", "error", err)
	}

	excutionIDs, err := ListExecutionIDs()
	assert.Nil(err)
	assert.Equal(4, len(excutionIDs))
	assert.Equal("exec_cmsp5a272ijn3jbg6850", excutionIDs[3])
	assert.Equal("exec_cmspbei72ijj85acej90", excutionIDs[2])
	assert.Equal("exec_cmspbeq72ijj85acejfg", excutionIDs[1])
	assert.Equal("exec_cmspbf272ijj85acejm0", excutionIDs[0])
}
