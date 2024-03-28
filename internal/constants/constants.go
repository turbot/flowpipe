package constants

import (
	"bufio"
)

const (
	DefaultFlowpipeHost       = "http://localhost:7103"
	DefaultServerPort         = 7103
	DefaultListen             = "network"
	DefaultExecutionMode      = ExecutionModeAsynchronous
	DefaultWaitRetry          = 60
	ExecutionModeSynchronous  = "synchronous"
	ExecutionModeAsynchronous = "asynchronous"

	MaxScanSize = bufio.MaxScanTokenSize * 40

	FormUrl = "form_url"
)

const FlowpipeSampleContent = `
#
# For detailed descriptions, see the reference documentation
# at https://flowpipe.io/docs
#

# integration "http" "default" {}

# notifier "default" {
#   notify {
#     integration = integration.http.default
#   }
# }
`
