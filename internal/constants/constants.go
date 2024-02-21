package constants

import (
	"bufio"
)

const (
	DefaultServerPort         = 7103
	DefaultListen             = "network"
	DefaultExecutionMode      = ExecutionModeAsynchronous
	DefaultWaitRetry          = 60
	ExecutionModeSynchronous  = "synchronous"
	ExecutionModeAsynchronous = "asynchronous"

	MaxScanSize = bufio.MaxScanTokenSize * 40
)

const FlowpipeSampleContent = `
#
# For detailed descriptions, see the reference documentation
# at https://flowpipe.io/docs
#

# integration "webform" "default" {}

# notifier "default" {
#   notify {
#     integration = integration.webform.default
#   }
# }
`
