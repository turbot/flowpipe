package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/sanitize"
)

type PrintableResource interface {
	Transform(flowpipeapiclient.FlowpipeAPIResource) (any, error)
	GetItems(sanitizer *sanitize.Sanitizer) any
	GetTable() (Table, error)
}
