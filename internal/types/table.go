package types

import (
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

type TableRow struct {
	Cells []interface{}
}

type Table struct {
	Rows    []TableRow
	Columns []TableColumnDefinition
}

func (Table) Transform(r flowpipeapiclient.FlowpipeAPIResource) (interface{}, error) {
	return nil, pcerr.BadRequestWithMessage("not supported")
}

func (p Table) GetItems() interface{} {
	return p.Rows
}

func (p Table) GetTable() (Table, error) {
	return p, nil
}

// Taken from kubectl
type TableColumnDefinition struct {
	// name is a human readable name for the column.
	Name string `json:"name"`
	// type is an OpenAPI type definition for this column, such as number, integer, string, or
	// array.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	Type string `json:"type"`
	// format is an optional OpenAPI type modifier for this column. A format modifies the type and
	// imposes additional rules, like date or time formatting for a string. The 'name' format is applied
	// to the primary identifier column which has type 'string' to assist in clients identifying column
	// is the resource name.
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#data-types for more.
	Format string `json:"format"`
	// description is a human readable description of this column.
	Description string `json:"description"`
	// priority is an integer defining the relative importance of this column compared to others. Lower
	// numbers are considered higher priority. Columns that may be omitted in limited space scenarios
	// should be given a higher priority.
	Priority int32 `json:"priority"`
}

func (t *TableColumnDefinition) Formatter() string {
	switch t.Type {
	case "integer":
		return "%d"
	case "number":
		return "%f"
	case "boolean":
		return "%t"
	case "string":
		return "%s"
	}
	return "%s"
}
