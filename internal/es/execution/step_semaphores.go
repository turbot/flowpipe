package execution

import (
	"log/slog"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
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
		slog.Info("Getting semaphore for http")
		globalHttpStepSemaphore <- struct{}{}
		slog.Info("Semaphore acquired for http")
	case "query":
		slog.Info("Getting semaphore for query")
		globalQueryStepSemaphore <- struct{}{}
		slog.Info("Semaphore acquired for query")
	case "container":
		slog.Info("Getting semaphore for container")
		globalContainerStepSemaphore <- struct{}{}
		slog.Info("Semaphore acquired for container")
	case "function":
		slog.Info("Getting semaphore for function")
		globalFunctionStepSemaphore <- struct{}{}
		slog.Info("Semaphore acquired for function")
	case "":
		slog.Warn("Step type is empty")
	}
}

func ReleaseStepTypeSemaphore(stepTeyp string) {
	switch stepTeyp {
	case "http":
		slog.Info("Releasing semaphore for http")
		<-globalHttpStepSemaphore
		slog.Info("Semaphore released for http")
	case "query":
		slog.Info("Releasing semaphore for query")
		<-globalQueryStepSemaphore
		slog.Info("Semaphore released for query")
	case "container":
		slog.Info("Releasing semaphore for container")
		<-globalContainerStepSemaphore
		slog.Info("Semaphore released for container")
	case "function":
		slog.Info("Releasing semaphore for function")
		<-globalFunctionStepSemaphore
		slog.Info("Semaphore released for function")
	case "":
		slog.Warn("Step type is empty")
	}
}
