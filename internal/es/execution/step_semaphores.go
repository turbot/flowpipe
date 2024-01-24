package execution

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
)

var globalHttpStepSemaphore chan struct{}
var globalQueryStepSemaphore chan struct{}
var globalContainerStepSemaphore chan struct{}
var globalFunctionStepSemaphore chan struct{}

func InitGlobalStepSemaphores() {
	// this helps with automated testing where we don't set Viper config
	httpMaxConcurrency := viper.GetInt(constants.ArgMaxConcurrencyHttp)
	if httpMaxConcurrency == 0 {
		httpMaxConcurrency = 500
	}

	queryMaxConcurrency := viper.GetInt(constants.ArgMaxConcurrencyQuery)
	if queryMaxConcurrency == 0 {
		queryMaxConcurrency = 50
	}

	containerMaxConcurrency := viper.GetInt(constants.ArgMaxConcurrencyContainer)
	if containerMaxConcurrency == 0 {
		containerMaxConcurrency = 25
	}

	functionMaxConcurrency := viper.GetInt(constants.ArgMaxConcurrencyFunction)
	if functionMaxConcurrency == 0 {
		functionMaxConcurrency = 50
	}

	globalHttpStepSemaphore = make(chan struct{}, httpMaxConcurrency)
	globalQueryStepSemaphore = make(chan struct{}, queryMaxConcurrency)
	globalContainerStepSemaphore = make(chan struct{}, containerMaxConcurrency)
	globalFunctionStepSemaphore = make(chan struct{}, functionMaxConcurrency)
}

func GetStepTypeSemaphore(stepType string) {
	switch stepType {
	case "http":
		slog.Debug("Getting semaphore for http")
		globalHttpStepSemaphore <- struct{}{}
		slog.Debug("Semaphore acquired for http")
	case "query":
		slog.Debug("Getting semaphore for query")
		globalQueryStepSemaphore <- struct{}{}
		slog.Debug("Semaphore acquired for query")
	case "container":
		slog.Debug("Getting semaphore for container")
		globalContainerStepSemaphore <- struct{}{}
		slog.Debug("Semaphore acquired for container")
	case "function":
		slog.Debug("Getting semaphore for function")
		globalFunctionStepSemaphore <- struct{}{}
		slog.Debug("Semaphore acquired for function")
	case "":
		slog.Warn("Step type is empty")
	}
}

func ReleaseStepTypeSemaphore(stepTeyp string) {
	switch stepTeyp {
	case "http":
		slog.Debug("Releasing semaphore for http")
		<-globalHttpStepSemaphore
		slog.Debug("Semaphore released for http")
	case "query":
		slog.Debug("Releasing semaphore for query")
		<-globalQueryStepSemaphore
		slog.Debug("Semaphore released for query")
	case "container":
		slog.Debug("Releasing semaphore for container")
		<-globalContainerStepSemaphore
		slog.Debug("Semaphore released for container")
	case "function":
		slog.Debug("Releasing semaphore for function")
		<-globalFunctionStepSemaphore
		slog.Debug("Semaphore released for function")
	case "":
		slog.Warn("Step type is empty")
	}
}

func pipelineStepSemaphoreCacheKey(pipelineExecutionID string, stepDefn modconfig.PipelineStep) string {
	return pipelineExecutionID + "-" + stepDefn.GetFullyQualifiedName()
}

func GetPipelineExecutionStepSemaphore(pipelineExecutionID string, stepDefn modconfig.PipelineStep) {
	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to get pipeline execution step semaphore")
		return
	}

	if stepDefn.GetMaxConcurrency() == nil {
		return
	}

	addToPipelineExecutionStepIndex(pipelineExecutionID, stepDefn)
	cacheKey := pipelineStepSemaphoreCacheKey(pipelineExecutionID, stepDefn)
	cachedChannel, found := cache.GetCache().Get(cacheKey)

	var semaphore chan struct{}
	if !found {

		semaphore = make(chan struct{}, *stepDefn.GetMaxConcurrency())
		// Effectively forever
		cache.GetCache().SetWithTTL(cacheKey, semaphore, 10*365*24*time.Hour)
	}

	if cachedChannel != nil {
		semaphore = cachedChannel.(chan struct{})
	}

	slog.Debug("Getting semaphore for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	semaphore <- struct{}{}
	slog.Debug("Semaphore acquired for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
}

func ReleasePipelineExecutionStepSemaphore(pipelineExecutionID string, stepDefn modconfig.PipelineStep) {
	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to release pipeline execution step semaphore")
		return
	}

	cacheKey := pipelineStepSemaphoreCacheKey(pipelineExecutionID, stepDefn)
	cachedChannel, found := cache.GetCache().Get(cacheKey)

	var semaphore chan struct{}
	if !found {
		slog.Warn("Semaphore not found for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
		return
	}

	if cachedChannel != nil {
		semaphore = cachedChannel.(chan struct{})
	}

	slog.Debug("Releasing semaphore for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	<-semaphore
	slog.Debug("Semaphore released for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
}

func CompletePipelineExecutionStepSemaphore(pipelineExecutionID string) {
	pipelineStepExecutionCacheMapCached, found := cache.GetCache().Get(pipelineExecutionStepSemaphoreCacheKey(pipelineExecutionID))

	if !found {
		return
	}

	pipelineStepExecutionCacheMap := pipelineStepExecutionCacheMapCached.(map[string]bool)

	for cacheKey := range pipelineStepExecutionCacheMap {
		cache.GetCache().Delete(cacheKey)
	}

	slog.Debug("Complete pipeline execution step semaphore", "pipeline_execution_id", pipelineExecutionID)
	cache.GetCache().Delete(pipelineExecutionStepSemaphoreCacheKey(pipelineExecutionID))
}

func pipelineExecutionStepSemaphoreCacheKey(pipelineExecutionID string) string {
	return pipelineExecutionID + "-pipeline_step_execution_cache_map"
}

func addToPipelineExecutionStepIndex(pipelineExecutionID string, stepDefn modconfig.PipelineStep) {
	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to get pipeline execution step index")
		return
	}

	cacheKey := pipelineExecutionStepSemaphoreCacheKey(pipelineExecutionID)
	pipelineStepExecutionCacheMapCached, found := cache.GetCache().Get(cacheKey)

	var pipelineStepExecutionCacheMap map[string]bool
	if !found {
		pipelineStepExecutionCacheMap = make(map[string]bool)
		cache.GetCache().SetWithTTL(cacheKey, pipelineStepExecutionCacheMap, 10*365*24*time.Hour)
	} else {
		pipelineStepExecutionCacheMap = pipelineStepExecutionCacheMapCached.(map[string]bool)
	}

	pipelineStepExecutionCacheMap[pipelineStepSemaphoreCacheKey(pipelineExecutionID, stepDefn)] = true
}
