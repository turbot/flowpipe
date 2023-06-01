package types

import flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"

type PrintableResource interface {
	Transform(flowpipeapiclient.FlowpipeAPIResource) (interface{}, error)
	GetItems() interface{}
	GetTable() (Table, error)
}
