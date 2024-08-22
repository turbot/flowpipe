package execution

import (
	"context"
	"log/slog"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"golang.org/x/sync/semaphore"
)

var globalHttpStepSemaphore *semaphore.Weighted
var globalQueryStepSemaphore *semaphore.Weighted
var globalContainerStepSemaphore *semaphore.Weighted
var globalFunctionStepSemaphore *semaphore.Weighted

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

	globalHttpStepSemaphore = semaphore.NewWeighted(int64(httpMaxConcurrency))
	globalQueryStepSemaphore = semaphore.NewWeighted(int64(queryMaxConcurrency))
	globalContainerStepSemaphore = semaphore.NewWeighted(int64(containerMaxConcurrency))
	globalFunctionStepSemaphore = semaphore.NewWeighted(int64(functionMaxConcurrency))
}

func GetStepTypeSemaphore(stepType string) error {
	var err error
	switch stepType {
	case "http":
		slog.Debug("Getting semaphore for http")
		err = globalHttpStepSemaphore.Acquire(context.Background(), 1)
		slog.Debug("Semaphore acquired for http")
	case "query":
		slog.Debug("Getting semaphore for query")
		err = globalQueryStepSemaphore.Acquire(context.Background(), 1)
		slog.Debug("Semaphore acquired for query")
	case "container":
		slog.Debug("Getting semaphore for container")
		err = globalContainerStepSemaphore.Acquire(context.Background(), 1)
		slog.Debug("Semaphore acquired for container")
	case "function":
		slog.Debug("Getting semaphore for function")
		err = globalFunctionStepSemaphore.Acquire(context.Background(), 1)
		slog.Debug("Semaphore acquired for function")
	case "":
		slog.Warn("Step type is empty")
	}
	return err
}

func ReleaseStepTypeSemaphore(stepTeyp string) {
	switch stepTeyp {
	case "http":
		slog.Debug("Releasing semaphore for http")
		globalHttpStepSemaphore.Release(1)
		slog.Debug("Semaphore released for http")
	case "query":
		slog.Debug("Releasing semaphore for query")
		globalQueryStepSemaphore.Release(1)
		slog.Debug("Semaphore released for query")
	case "container":
		slog.Debug("Releasing semaphore for container")
		globalContainerStepSemaphore.Release(1)
		slog.Debug("Semaphore released for container")
	case "function":
		slog.Debug("Releasing semaphore for function")
		globalFunctionStepSemaphore.Release(1)
		slog.Debug("Semaphore released for function")
	case "":
		slog.Warn("Step type is empty")
	}
}
