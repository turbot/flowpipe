package primitive

import "github.com/turbot/steampipe-pipelines/pipeline"

type Primitive interface {
	Run(pipeline.StepInput) (*pipeline.Output, error)
}
