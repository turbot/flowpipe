package execution

import (
	"context"
	"log/slog"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/cache"
	"golang.org/x/sync/semaphore"
)

func pipelineStepSemaphoreCacheKey(pipelineExecutionID string, stepDefn resources.PipelineStep) string {
	return pipelineExecutionID + "-" + stepDefn.GetFullyQualifiedName()
}

func GetPipelineExecutionStepSemaphore(pipelineExecutionID string, stepDefn resources.PipelineStep, evalContext *hcl.EvalContext) error {
	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to get pipeline execution step semaphore")
		return nil
	}

	stepDefnMaxConcurrency := stepDefn.GetMaxConcurrency(evalContext)
	if stepDefnMaxConcurrency == nil {
		return nil
	}

	return GetPipelineExecutionStepSemaphoreMaxConcurrency(pipelineExecutionID, stepDefn, stepDefnMaxConcurrency, false)
}

func GetPipelineExecutionStepSemaphoreMaxConcurrency(pipelineExecutionID string, stepDefn resources.PipelineStep, stepDefnMaxConcurrency *int, tryAcquire bool) error {
	if stepDefnMaxConcurrency == nil {
		return nil
	}

	if stepDefn == nil || pipelineExecutionID == "" {
		slog.Warn("Step definition or pipeline execution ID is nil, unable to get pipeline execution step semaphore")
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

	slog.Info("Getting semaphore for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	if tryAcquire {
		res := sem.TryAcquire(1)
		slog.Info("Try acquire semaphore for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName(), "result", res)
	} else {
		err := sem.Acquire(context.Background(), 1)
		if err != nil {
			slog.Error("Error acquiring semaphore", "error", err)
			return err
		}
	}
	slog.Info("Semaphore acquired for pipeline execution step", "pipeline_execution_id", pipelineExecutionID, "step_name", stepDefn.GetName())
	return nil
}

func ReleasePipelineExecutionStepSemaphore(pipelineExecutionID string, stepDefn resources.PipelineStep) error {
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

func addToPipelineExecutionStepIndex(pipelineExecutionID string, stepDefn resources.PipelineStep) {
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
