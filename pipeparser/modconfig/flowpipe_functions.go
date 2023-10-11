package modconfig

type FlowpipeFunction struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Env     map[string]string `json:"env" cty:"env"`
	Runtime string            `json:"runtime" cty:"runtime"`
	Src     string            `json:"src" cty:"src"`
}

func (f *FlowpipeFunction) Equals(other *FlowpipeFunction) bool {
	if f == nil || other == nil {
		return false
	}
	return f.FullName == other.FullName &&
		f.Runtime == other.Runtime &&
		f.Src == other.Src

	// &&
	//f.EnvEquals(other)
}
