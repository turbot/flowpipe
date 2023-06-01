package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/fperr"
)

type TableRow struct {
	Cells []string
}

type Table struct {
	Rows []TableRow
}

func (Table) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {
	return nil, fperr.BadRequestWithMessage("not supported")
}

func (p Table) GetItems() interface{} {
	return p.Rows
}

func (p Table) GetTable() (Table, error) {
	return p, nil
}
