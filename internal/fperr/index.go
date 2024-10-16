package fperr

import (
	"reflect"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
)

const (
	ErrorCodeModLoadFailed    = "error_mod_load_failed"
	ErrorCodeModInstallFailed = "error_mod_install_failed"
	ErrorCodeAPIInitFailed    = "error_api_init_failed"
	ErrorCodeUnknownError     = "error_unknown_error"
	ErrorCodeNotFound         = "error_not_found"

	ExitCodeExecutionPaused      = 1
	ExitCodeExecutionFailed      = 2
	ExitCodeExecutionCancelled   = 3
	ExitCodeNotFound             = 4
	ExitCodeExecutionDidNotStart = 9
	ExitCodeUnknownFlowpipeError = 10
)

func GetExitCode(err error, fromPanic bool) int {
	if e, ok := err.(perr.ErrorModel); ok {
		switch e.Type {
		case ErrorCodeModLoadFailed:
			return constants.ExitCodeModInitFailed
		case ErrorCodeModInstallFailed:
			return constants.ExitCodeModInstallFailed
		case ErrorCodeAPIInitFailed:
			return constants.ExitCodeInitializationFailed
		case ErrorCodeNotFound:
			return ExitCodeNotFound
		case ErrorCodeUnknownError:
			return ExitCodeUnknownFlowpipeError
		}
	}

	if fromPanic {
		return constants.ExitCodeUnknownErrorPanic
	}

	// maybe one day we'll have a different exit code unknown panic vs unknown error
	return constants.ExitCodeUnknownErrorPanic
}

func FailOnError(sourceError error, wrapWith reflect.Type, errorCode string) {
	if sourceError == nil {
		return
	}

	flowpipeError := WrapsWith(sourceError, wrapWith, errorCode)
	error_helpers.FailOnError(flowpipeError)
}

func FailOnErrorWithMessage(sourceError error, message string, wrapWith reflect.Type, errorCode string) {
	if sourceError == nil {
		return
	}

	flowpipeError := WrapsWith(sourceError, wrapWith, errorCode)
	flowpipeError.Detail += " " + message
	error_helpers.FailOnError(flowpipeError)
}

func WrapsWith(sourceError error, wrapWith reflect.Type, errorCode string) perr.ErrorModel {
	if flowpipeError, ok := sourceError.(perr.ErrorModel); ok {
		if flowpipeError.Type == "" {
			flowpipeError.Type = errorCode
		}
		return flowpipeError
	}

	if wrapWith != nil {
		// create an instance of wrapWith
		wrapInstance := reflect.New(wrapWith).Interface()
		if flowpipeError, ok := wrapInstance.(perr.ErrorModel); ok {
			flowpipeError.Type = errorCode
			flowpipeError.Detail = sourceError.Error()
			return flowpipeError
		}
	}

	// Wrap the error in an internal error
	flowpipeError := perr.Internal(sourceError)
	flowpipeError.Type = errorCode
	return flowpipeError
}
