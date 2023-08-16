package pipeparser

import (
	"fmt"
	"strings"

	"github.com/turbot/flowpipe/pipeparser/modconfig"

	"github.com/hashicorp/hcl/v2"
)

//nolint:unused // TODO: check usage
type unresolvedBlock struct {
	Name         string
	Block        *hcl.Block
	Dependencies map[string]*modconfig.ResourceDependency
}

//nolint:unused // TODO: check usage
func (b unresolvedBlock) String() string {
	depStrings := make([]string, len(b.Dependencies))
	idx := 0
	for _, dep := range b.Dependencies {
		depStrings[idx] = fmt.Sprintf(`%s -> %s`, b.Name, dep.String())
		idx++
	}
	return strings.Join(depStrings, "\n")
}
