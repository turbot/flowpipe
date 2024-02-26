package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/cache"
)

func MapStepExecutionID(executionID, pipelineExecutionID, stepExecutionID string) string {

	// get stepExecutionID without sexec_ prefix
	key := strings.TrimPrefix(stepExecutionID, "sexec_")

	fullID := fmt.Sprintf("%s.%s.%s", executionID, pipelineExecutionID, stepExecutionID)

	// effectively forever
	cache.GetCache().SetWithTTL(key, fullID, 10*365*24*time.Hour)

	return key
}

func ResolveShortStepExecutionID(key string) (executionID string, pipelineExecutionID string, stepExecutionID string, found bool) {
	fullID, found := cache.GetCache().Get(key)
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
	key := strings.TrimPrefix(stepExecutionID, "sexec_")
	cache.GetCache().Delete(key)
}
