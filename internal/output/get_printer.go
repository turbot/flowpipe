package output

import (
	"github.com/spf13/cobra"
	"github.com/turbot/pipe-fittings/printers"
)

func GetPrinter[T any](cmd *cobra.Command) (printers.ResourcePrinter[T], error) {
	return printers.GetPrinter[T](cmd, printers.WithStringPrinterCommands([]string{"flowpipe.mod.list"}))
}
