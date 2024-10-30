package resources

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

func GetModResources(mod *modconfig.Mod) *FlowpipeModResources {
	resourceMaps, ok := mod.GetResourceMaps().(*FlowpipeModResources)
	if !ok {
		// should never happen
		panic(fmt.Sprintf("mod.GetResourceMaps() did not return a flowpipe ModResources: %T", mod.GetResourceMaps()))
	}
	return resourceMaps
}

// FlowpipeModResources is a struct containing maps of all mod resource types
// This is provided to avoid db needing to reference workspace package
type FlowpipeModResources struct {
	// the parent mod
	Mod *modconfig.Mod

	Variables map[string]*modconfig.Variable
	// all mods (including deps)
	// TODO KAI store as interfaces so GetMods satisfies the interface
	Mods       map[string]*modconfig.Mod
	References map[string]*modconfig.ResourceReference
	Pipelines  map[string]*Pipeline
	Triggers   map[string]*Trigger
	Locals     map[string]*modconfig.Local
}

func NewModResources(mod *modconfig.Mod, sourceMaps ...modconfig.ModResources) modconfig.ModResources {
	res := emptyFlowpipeModResources()
	res.Mod = mod
	res.Mods[mod.GetInstallCacheKey()] = mod
	res.AddMaps(sourceMaps...)
	return res
}

func emptyFlowpipeModResources() *FlowpipeModResources {
	return &FlowpipeModResources{

		Mods:      make(map[string]*modconfig.Mod),
		Variables: make(map[string]*modconfig.Variable),
		Locals:    make(map[string]*modconfig.Local),

		// Flowpipe
		Pipelines: make(map[string]*Pipeline),
		Triggers:  make(map[string]*Trigger),
	}
}

//// TopLevelResources returns a new ModResources containing only top level resources (i.e. no dependencies)
//func (m *ModResources) TopLevelResources() *ModResources {
//	res := NewModResources(m.Mod)
//
//	f := func(item HclResource) (bool, error) {
//		if modItem, ok := item.(ModItem); ok {
//			if mod := modItem.GetMod(); mod != nil && mod.GetFullName() == m.Mod.GetFullName() {
//				// the only error we expect is a duplicate item error - ignore
//				_ = res.AddResource(item)
//			}
//		}
//		return true, nil
//	}
//
//	// resource func does not return an error
//	_ = m.WalkResources(f)
//
//	return res
//}

func (m *FlowpipeModResources) Equals(o modconfig.ModResources) bool {
	other, ok := o.(*FlowpipeModResources)
	if !ok {
		return false
	}

	if other == nil {
		return false
	}

	for name, variable := range m.Variables {
		if otherVariable, ok := other.Variables[name]; !ok {
			return false
		} else if !variable.Equals(otherVariable) {
			return false
		}
	}
	for name := range other.Variables {
		if _, ok := m.Variables[name]; !ok {
			return false
		}
	}

	for name, pipeline := range m.Pipelines {
		if otherPipeline, ok := other.Pipelines[name]; !ok {
			return false
		} else if !pipeline.Equals(otherPipeline) {
			return false
		}
	}
	for name := range other.Pipelines {
		if _, ok := m.Pipelines[name]; !ok {
			return false
		}
	}

	// TODO K: do we need integration & notifier here?

	for name, trigger := range m.Triggers {
		if otherTrigger, ok := other.Triggers[name]; !ok {
			return false
		} else if !trigger.Equals(otherTrigger) {
			return false
		}
	}
	for name := range other.Triggers {
		if _, ok := m.Triggers[name]; !ok {
			return false
		}
	}

	for name := range other.Locals {
		if _, ok := m.Locals[name]; !ok {
			return false
		}
	}

	return true
}

// GetResource tries to find a resource with the given name in the ModResources
// NOTE: this does NOT support inputs, which are NOT uniquely named in a mod
func (m *FlowpipeModResources) GetResource(parsedName *modconfig.ParsedResourceName) (resource modconfig.HclResource, found bool) {
	modName := parsedName.Mod
	if modName == "" {
		modName = m.Mod.ShortName
	}
	longName := fmt.Sprintf("%s.%s.%s", modName, parsedName.ItemType, parsedName.Name)

	// NOTE: we could use WalkResources, but this is quicker

	switch parsedName.ItemType {
	// note the special case for variables - "var" rather than "variable"
	case schema.AttributeVar:
		resource, found = m.Variables[longName]
	case schema.BlockTypePipeline:
		resource, found = m.Pipelines[longName]
	case schema.BlockTypeTrigger:
		resource, found = m.Triggers[longName]
	case schema.BlockTypeMod:
		for _, mod := range m.Mods {
			if mod.ShortName == parsedName.Name {
				resource = mod
				found = true
				break
			}
		}

	}
	return resource, found
}

func (m *FlowpipeModResources) Empty() bool {
	return len(m.Mods)+
		len(m.Variables)+
		len(m.Pipelines)+
		len(m.Triggers) == 0
}

// WalkResources calls resourceFunc for every resource in the mod
// if any resourceFunc returns false or an error, return immediately
func (m *FlowpipeModResources) WalkResources(resourceFunc func(item modconfig.HclResource) (bool, error)) error {
	for _, r := range m.Mods {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	// we cannot walk source snapshots as they are not a HclResource
	for _, r := range m.Variables {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	for _, r := range m.Pipelines {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}

	for _, r := range m.Triggers {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	for _, r := range m.Locals {
		if continueWalking, err := resourceFunc(r); err != nil || !continueWalking {
			return err
		}
	}
	return nil
}

func (m *FlowpipeModResources) AddResource(item modconfig.HclResource) hcl.Diagnostics {
	var diags hcl.Diagnostics
	switch r := item.(type) {
	case *Pipeline:
		name := r.Name()
		if existing, ok := m.Pipelines[name]; ok {
			diags = append(diags, modconfig.CheckForDuplicate(existing, item)...)
			break
		}
		m.Pipelines[name] = r

	case *Trigger:
		name := r.Name()
		if existing, ok := m.Triggers[name]; ok {
			diags = append(diags, modconfig.CheckForDuplicate(existing, item)...)
			break
		}
		m.Triggers[name] = r
	case *modconfig.Variable:
		name := r.Name()
		if existing, ok := m.Variables[name]; ok {
			diags = append(diags, modconfig.CheckForDuplicate(existing, item)...)
			break
		}
		m.Variables[name] = r
	case *modconfig.Local:
		name := r.Name()
		if existing, ok := m.Locals[name]; ok {
			diags = append(diags, modconfig.CheckForDuplicate(existing, item)...)
			break
		}
		m.Locals[name] = r
	}

	return diags
}

func (m *FlowpipeModResources) AddMaps(sourceMaps ...modconfig.ModResources) {
	for _, s := range sourceMaps {
		source := s.(*FlowpipeModResources)
		for k, v := range source.
			Pipelines {
			m.Pipelines[k] = v
		}
		for k, v := range source.Triggers {
			m.Triggers[k] = v
		}
		for k, v := range source.Pipelines {
			m.Pipelines[k] = v
		}
		for k, v := range source.Variables {
			// TODO check why this was necessary and test variables thoroughly
			// NOTE: only include variables from root mod  - we add in the others separately
			//if v.Mod.GetFullName() == m.Mod.GetFullName() {
			m.Variables[k] = v
			//}
		}
		for k, v := range source.Locals {
			m.Locals[k] = v
		}
		for k, v := range source.Mods {
			m.Mods[k] = v
		}

	}
}
func (m *FlowpipeModResources) AddReference(ref *modconfig.ResourceReference) {
	m.References[ref.String()] = ref
}

func (m *FlowpipeModResources) GetReferences() map[string]*modconfig.ResourceReference {
	return m.References
}

func (m *FlowpipeModResources) GetVariables() map[string]*modconfig.Variable {
	return m.Variables
}

func (m *FlowpipeModResources) GetMods() map[string]*modconfig.Mod {

	return m.Mods
}

// TopLevelResources returns a new PowerpipeResourceMaps containing only top level resources (i.e. no dependencies)
func (m *FlowpipeModResources) TopLevelResources() modconfig.ModResources {
	res := NewModResources(m.Mod)

	f := func(item modconfig.HclResource) (bool, error) {
		if modItem, ok := item.(modconfig.ModItem); ok {
			if mod := modItem.GetMod(); mod != nil && mod.GetFullName() == m.Mod.GetFullName() {
				// the only error we expect is a duplicate item error - ignore
				_ = res.AddResource(item)
			}
		}
		return true, nil
	}

	// resource func does not return an error
	_ = m.WalkResources(f)

	return res
}
