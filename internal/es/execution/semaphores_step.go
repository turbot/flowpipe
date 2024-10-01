package execution

import (
	"context"
	"log/slog"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/pipe-fittings/modconfig"
	"golang.org/x/sync/semaphore"
)

func pipelineStepSemaphoreCacheKey(pipelineExecutionID string, stepDefn modconfig.PipelineStep) string {
	return pipelineExecutionID + "-" + stepDefn.GetFullyQualifiedName()
}

func GetPipelineExecutionStepSemaphore(pipelineExecutionID string, stepDefn modconfig.PipelineStep, evalContext *hcl.EvalContext) error {
	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to get pipeline execution step semaphore")
		return nil
	}

	stepDefnMaxConcurrency := stepDefn.GetMaxConcurrency(evalContext)
	if stepDefnMaxConcurrency == nil {
		return nil
	}

	addToPipelineExecutionStepIndex(pipelineExecutionID, stepDefn)
	cacheKey := pipelineStepSemaphoreCacheKey(pipelineExecutionID, stepDefn)
	cachedChannel, found := cache.GetCache().Get(cacheKey)

	var sem *semaphore.Weighted
	if !found {

		sem = semaphore.NewWeighted(int64(*stepDefnMaxConcurrency))
		// Effectively forever
		cache.GetCache().SetWithTTL(cacheKey, sem, 10*365*24*time.Hour)
	}

	if cachedChannel != nil {
		sem = cachedChannel.(*semaphore.Weighted)
	}

	slog.Debug("Getting semaphore for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	err := sem.Acquire(context.Background(), 1)
	if err != nil {
		slog.Error("Error acquiring semaphore", "error", err)
		return err
	}
	slog.Debug("Semaphore acquired for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	return nil
}

func ReleasePipelineExecutionStepSemaphore(pipelineExecutionID string, stepDefn modconfig.PipelineStep) error {
	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to release pipeline execution step semaphore")
		return nil
	}

	cacheKey := pipelineStepSemaphoreCacheKey(pipelineExecutionID, stepDefn)
	cachedChannel, found := cache.GetCache().Get(cacheKey)

	var sem *semaphore.Weighted
	if !found {
		return nil
	}

	if cachedChannel != nil {
		sem = cachedChannel.(*semaphore.Weighted)
	}

	slog.Debug("Releasing semaphore for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	sem.Release(1)
	slog.Debug("Semaphore released for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	return nil
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
