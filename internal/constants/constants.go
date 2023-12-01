package constants

import "bufio"

const (
	DefaultServerPort         = 7103
	DefaultListen             = "localhost"
	DefaultExecutionMode      = ExecutionModeAsynchronous
	DefaultWaitRetry          = 60
	ExecutionModeSynchronous  = "synchronous"
	ExecutionModeAsynchronous = "asynchronous"

	MaxScanSize = bufio.MaxScanTokenSize * 40
)
