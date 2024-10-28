package execution

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"log/slog"
	"time"

	"github.com/turbot/flowpipe/internal/cache"
	"golang.org/x/sync/semaphore"
)

func pipelineSemaphoreCacheKey(name string) string {
	return name + "-pipeline-sempahore"
}

func GetPipelineSemaphore(pipelineDefn *flowpipe.Pipeline) error {
	if pipelineDefn == nil {
		slog.Warn("Pipeline definition is nil, unable to get pipeline semaphore")
		return nil
	}

	if pipelineDefn.MaxConcurrency == nil {
		return nil
	}

	cacheKey := pipelineSemaphoreCacheKey(pipelineDefn.FullName)
	cachedSem, found := cache.GetCache().Get(cacheKey)

	var sem *semaphore.Weighted
	if !found {
		sem = semaphore.NewWeighted(int64(*pipelineDefn.MaxConcurrency))

		// Effectively forever
		cache.GetCache().SetWithTTL(cacheKey, sem, 10*365*24*time.Hour)
	}

	if cachedSem != nil {
		sem = cachedSem.(*semaphore.Weighted)
	}

	slog.Debug("Getting semaphore for pipeline", "pipeline", cacheKey)
	err := sem.Acquire(context.Background(), 1)
	if err != nil {
		slog.Error("Error acquiring semaphore", "error", err)
		return err
	}
	slog.Debug("Semaphore acquired for pipeline", "pipeline", cacheKey)
	return nil
}

func ReleasePipelineSemaphore(pipelineDefn *flowpipe.Pipeline) error {
	if pipelineDefn == nil {
		slog.Warn("Pipeline definition is nil, unable to release pipeline semaphore")
		return nil
	}

	cacheKey := pipelineSemaphoreCacheKey(pipelineDefn.FullName)
	cachedSem, found := cache.GetCache().Get(cacheKey)

	var sem *semaphore.Weighted
	if !found {
		return nil
	}

	if cachedSem != nil {
		sem = cachedSem.(*semaphore.Weighted)
	}

	slog.Debug("Releasing semaphore for pipeline", "pipeline", cacheKey)
	sem.Release(1)
	slog.Debug("Semaphore released for pipeline", "pipeline", cacheKey)
	return nil
}
