package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/internal/util"
)

func TestCleanupDB(t *testing.T) {

	assert := assert.New(t)

	err := copyNewFlowpipeDbCleanFile("./clean_test_files/flowpipe_clean_2.db")
	if err != nil {
		assert.FailNow(err.Error())
	}

	// offset := time.Duration(-24*30) * time.Hour
	offset := -1 * time.Hour

	layout := util.RFC3389WithMS
	str := "2024-02-02T01:59:00.000Z"

	anchorTime, err := time.Parse(layout, str)
	if err != nil {
		assert.FailNow(err.Error())
	}

	currentTime := anchorTime.Add(1 * time.Hour)

	rowsAffected, err := cleanupFlowpipeDB(currentTime, offset)
	if err != nil {
		assert.FailNow(err.Error())
	}

	assert.Equal(2, rowsAffected, "rowsAffected should be 2")
}
