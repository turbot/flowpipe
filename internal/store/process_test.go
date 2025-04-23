package store

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func copyNewFlowpipeDbCleanFile(cleanSource string) error {
	// Open the source file
	sourceFile, err := os.Open(cleanSource)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file, this will overwrite if file already exists
	destFile, err := os.Create("flowpipe.db")
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Force to flush the file system's in-memory copy of recently written data to disk.
	err = destFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func TestLoadExecutionFromDB(t *testing.T) {
	assert := assert.New(t)

	err := copyNewFlowpipeDbCleanFile("./clean_test_files/flowpipe_clean.db")
	if err != nil {
		assert.FailNow(err.Error())
	}

	excutionIDs, err := ListExecutionIDs()
	assert.Nil(err)
	assert.Equal(3, len(excutionIDs))
	assert.Equal("exec_cqlecr4204vm48hs8lp0", excutionIDs[2])
	assert.Equal("exec_cqlecrk204vm4kl8io10", excutionIDs[1])
	assert.Equal("exec_cqled8k204vm0pm8h5dg", excutionIDs[0])
}
