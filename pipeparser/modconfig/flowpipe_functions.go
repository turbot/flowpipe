package modconfig

import (
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/zclconf/go-cty/cty"
)

type Function struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Env     map[string]string `json:"env" cty:"env"`
	Runtime string            `json:"runtime" cty:"runtime"`
	Src     string            `json:"src" cty:"src"`
	Handler string            `json:"handler" cty:"handler"`
}

func (f *Function) Equals(other *Function) bool {
	if f == nil || other == nil {
		return false
	}
	return f.FullName == other.FullName &&
		f.Runtime == other.Runtime &&
		f.Src == other.Src &&
		f.Handler == other.Handler

	// &&
	//f.EnvEquals(other)
}

func (f *Function) AsCtyValue() cty.Value {
	functionVars := map[string]cty.Value{}
	functionVars[schema.LabelName] = cty.StringVal(f.Name())

	if f.Description != nil {
		functionVars[schema.AttributeTypeDescription] = cty.StringVal(*f.Description)
	}

	return cty.ObjectVal(functionVars)
}
