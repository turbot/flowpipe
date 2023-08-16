package primitive

import (
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

type Primitive interface {
	Run(pipeline.Input) (*pipeline.Output, error)
}
