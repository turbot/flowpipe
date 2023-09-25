package modconfig

import (
	"strings"

	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type ParsedPropertyPath struct {
	Mod          string
	ItemType     string
	Name         string
	PropertyPath []string
	// optional scope of this property path ("self")
	Scope    string
	Original string
}

func (p *ParsedPropertyPath) PropertyPathString() string {
	return strings.Join(p.PropertyPath, ".")
}

func (p *ParsedPropertyPath) ToParsedResourceName() *ParsedResourceName {
	return &ParsedResourceName{
		Mod:      p.Mod,
		ItemType: p.ItemType,
		Name:     p.Name,
	}
}

func (p *ParsedPropertyPath) ToResourceName() string {
	return BuildModResourceName(p.ItemType, p.Name)
}

func (p *ParsedPropertyPath) String() string {
	return p.Original
}

func ParseResourcePropertyPath(propertyPath string) (*ParsedPropertyPath, error) {
	res := &ParsedPropertyPath{Original: propertyPath}

	// valid property paths:
	// <mod>.<resource>.<name>.<property path...>
	// <resource>.<name>.<property path...>
	// so either the first or second slice must be a valid resource type

	parts := strings.Split(propertyPath, ".")
	if len(parts) < 2 {
		return nil, perr.BadRequestWithMessage("invalid property path passed to ParseResourcePropertyPath: " + propertyPath)
	}

	// special case handling for runtime dependencies which may have use the "self" qualifier
	// const RuntimeDependencyDashboardScope = "self"
	if parts[0] == "self" {
		res.Scope = parts[0]
		parts = parts[1:]
	}

	if schema.IsValidResourceItemType(parts[0]) {
		// put empty mod as first part
		parts = append([]string{""}, parts...)
	}
	switch len(parts) {
	case 3:
		// no property path specified
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
	default:
		res.Mod = parts[0]
		res.ItemType = parts[1]
		res.Name = parts[2]
		res.PropertyPath = parts[3:]
	}

	if !schema.IsValidResourceItemType(res.ItemType) {
		return nil, perr.BadRequestWithMessage("invalid resource item type passed to ParseResourcePropertyPath: " + propertyPath)
	}

	return res, nil
}
