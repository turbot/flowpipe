package primitive

import (
	"context"

	"github.com/turbot/flowpipe/internal/docker"
	function "github.com/turbot/flowpipe/internal/functions"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

var functionCache = map[string]*function.Function{}

type Function struct{}

func (e *Function) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func (e *Function) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	fn := functionCache[input[schema.AttributeTypeFunction].(string)]

	if fn == nil {

		fn, err := function.New(
			function.WithContext(ctx),
			function.WithDockerClient(docker.GlobalDockerClient),
		)
		if err != nil {
			return nil, err
		}
		fn.Name = input[schema.AttributeTypeFunction].(string)
		fn.Runtime = input[schema.AttributeTypeRuntime].(string)
		if input[schema.AttributeTypeHandler] != nil {
			fn.Handler = input[schema.AttributeTypeHandler].(string)
		}
		fn.Src = input[schema.AttributeTypeSrc].(string)
		err = fn.Load()
		if err != nil {
			return nil, err
		}

		functionCache[fn.Name] = fn
	}

	result, err := fn.Invoke([]byte{})
	if err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{
			"result": string(result),
		},
	}

	return &o, nil
}
