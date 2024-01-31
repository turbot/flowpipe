package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCleanupDB(t *testing.T) {

	assert := assert.New(t)

	err := copyNewFlowpipeDbCleanFile("./clean_test_files/flowpipe_clean_2.db")
	if err != nil {
		assert.FailNow(err.Error())
	}

	// offset := time.Duration(-24*30) * time.Hour
	offset := -1 * time.Hour

	layout := "2006-01-02T15:04:05Z" // This is the reference layout for ISO 8601 in Go
	str := "2024-01-31T09:12:00Z"

	anchorTime, err := time.Parse(layout, str)
	if err != nil {
		assert.FailNow(err.Error())
	}

	currentTime := anchorTime.Add(1 * time.Hour)

	rowsAffected, err := CleanupFlowpipeDB(currentTime, offset)
	if err != nil {
		assert.FailNow(err.Error())
	}

	assert.Equal(10, rowsAffected, "rowsAffected should be 10")
}
