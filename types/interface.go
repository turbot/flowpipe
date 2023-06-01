package types

type FlowpipeResources interface {
	GetResources() []FlowpipeResource
}

type FlowpipeResource interface {
	GetType() string
}
