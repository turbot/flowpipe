package cmdconfig

import (
	"github.com/spf13/cobra"
	"slices"
	"strings"
)

func CommandFullKey(cmd *cobra.Command) string {
	var parents []string
	parents = append(parents, cmd.Name())
	cmd.VisitParents(func(parent *cobra.Command) {
		parents = append(parents, parent.Name())
	})

	slices.Reverse(parents)

	return strings.Join(parents, ".")
}
