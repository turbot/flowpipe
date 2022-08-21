package primitive

type Input map[string]interface{}

type Output map[string]interface{}

type Primitive interface {
	Run(Input) (Output, error)
}
