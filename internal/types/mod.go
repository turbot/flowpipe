package types

import "github.com/turbot/pipe-fittings/modconfig"

type Mod struct {
	Name          string     `json:"name"`
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Documentation *string    `json:"documentation,omitempty"`
	Color         *string    `json:"color,omitempty"`
	Categories    []string   `json:"categories,omitempty"`
	OpenGraph     *OpenGraph `json:"opengraph,omitempty"`
}

func NewModFromModConfigMod(mod *modconfig.Mod) *Mod {
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

	return &fpMod
}

type OpenGraph struct {
	// The opengraph description (og:description) of the mod, for use in social media applications
	Description *string `json:"description,omitempty"`
	Title       *string `json:"title,omitempty"`
}
