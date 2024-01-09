package output

import (
	"github.com/spf13/cobra"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/pipe-fittings/printers"
)

func GetPrinter[T any](cmd *cobra.Command) (printers.ResourcePrinter[T], error) {
	return printers.GetPrinter[T](cmd, printers.WithTableCommands(localcmdconfig.CommandsWithTableOutput))
}
