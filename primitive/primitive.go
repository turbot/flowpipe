package primitive

import "github.com/turbot/steampipe-pipelines/pipeline"

type Primitive interface {
	Run(pipeline.Input) (*pipeline.Output, error)
}
