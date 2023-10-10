package modconfig

type Function struct {
	HclResourceImpl
	ResourceWithMetadataImpl

	Name    string            `json:"name"`
	Env     map[string]string `json:"env" cty:"env"`
	Runtime string            `json:"runtime" cty:"runtime"`
}
