package constants

import (
	"bufio"
)

const (
	DefaultServerPort         = 7103
	DefaultListen             = "localhost"
	DefaultExecutionMode      = ExecutionModeAsynchronous
	DefaultWaitRetry          = 60
	ExecutionModeSynchronous  = "synchronous"
	ExecutionModeAsynchronous = "asynchronous"

	MaxScanSize = bufio.MaxScanTokenSize * 40
)

const DefaultFlowpipeIntegrationContent = `
integration "webform" "default" {}
`

const DefaultFlowpipeNotifierContent = `
notifier "default" {
  integration "default" {
    base = integration.webform.default 
  }
}
`
