package output

import (
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/printers"
)

func GetPrinter[T any](cmd *cobra.Command) (printers.ResourcePrinter[T], error) {
	return printers.GetPrinter[T](cmd, printers.WithTableCommands([]string{
		"flowpipe.trigger.list",
		"flowpipe.pipeline.list",
		"flowpipe.process.list",
	}))
}
