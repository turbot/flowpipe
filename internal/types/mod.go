package types

import (
	"github.com/turbot/pipe-fittings/modconfig"
)

type Mod struct {
	Name          string     `json:"name"`
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Documentation *string    `json:"documentation,omitempty"`
	Color         *string    `json:"color,omitempty"`
	Categories    []string   `json:"categories,omitempty"`
	OpenGraph     *OpenGraph `json:"opengraph,omitempty"`
	Require       *Require   `json:"require,omitempty"`
}

type Require struct {
	Flowpipe *FlowpipeRequire       `json:"flowpipe,omitempty"`
	Mods     []ModVersionConstraint `json:"mods,omitempty"`
}

type FlowpipeRequire struct {
	MinVersionString string `json:"min_version,omitempty"`
}

type ModVersionConstraint struct {
	// the fully qualified mod name, e.g. github.com/turbot/mod1
	Name          string `json:"name"`
	VersionString string `json:"version,omitempty"`
}

func NewModFromModConfigMod(mod modconfig.ModI) *Mod {
	fpMod := Mod{
		Name:          mod.Name(),
		Title:         mod.Title,
		Description:   mod.Description,
		Documentation: mod.Documentation,
		Color:         mod.Color,
		Categories:    mod.Categories,
	}

	if mod.OpenGraph != nil {
		fpMod.OpenGraph = &OpenGraph{
			Description: mod.OpenGraph.Description,
			Title:       mod.OpenGraph.Title,
		}
	}

	if mod.Require != nil {
		fpMod.Require = &Require{}
		if mod.Require.Flowpipe != nil {
			fpMod.Require.Flowpipe = &FlowpipeRequire{
				MinVersionString: mod.Require.Flowpipe.MinVersionString,
			}
		}
		if mod.Require.Mods != nil {
			fpMod.Require.Mods = make([]ModVersionConstraint, len(mod.Require.Mods))

			for i, mvc := range mod.Require.Mods {
				fpMod.Require.Mods[i] = ModVersionConstraint{
					Name:          mvc.Name,
					VersionString: mvc.VersionString,
				}
			}
		}
	}

	return &fpMod
}

type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description *string `json:"description,omitempty"`
	Title       *string `json:"title,omitempty"`
}
