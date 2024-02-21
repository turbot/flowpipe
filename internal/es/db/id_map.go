package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/cache"
)

func MapStepExecutionID(executionID, pipelineExecutionID, stepExecutionID string) string {

	// get the last 8 chars of stepExecutionID
	last8 := stepExecutionID[len(stepExecutionID)-8:]

	fullID := fmt.Sprintf("%s.%s.%s", executionID, pipelineExecutionID, stepExecutionID)

	// effectively forever
	cache.GetCache().SetWithTTL(last8, fullID, 10*365*24*time.Hour)

	return last8
}

func ResolveShortStepExecutionID(last8 string) (executionID string, pipelineExecutionID string, stepExecutionID string, found bool) {
	fullID, found := cache.GetCache().Get(last8)
	if !found {
		return "", "", "", false
	}

	fullIDStr, ok := fullID.(string)
	if !ok {
		return "", "", "", false
	}

	parts := strings.Split(fullIDStr, ".")
	return parts[0], parts[1], parts[2], true
}

func RemoveStepExecutionIDMap(stepExecutionID string) {
	last8 := stepExecutionID[len(stepExecutionID)-8:]
	cache.GetCache().Delete(last8)
}
