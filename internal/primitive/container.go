package primitive

import (
	"context"

	"github.com/turbot/flowpipe/internal/container"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

var containerCache = map[string]*container.Container{}

type Container struct{}

func (e *Container) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func (e *Container) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	c := containerCache[input[schema.AttributeTypeImage].(string)]

	if c == nil {
		var err error

		c, err = container.NewContainer(
			container.WithContext(context.Background()),
			container.WithDockerClient(docker.GlobalDockerClient),
		)
		if err != nil {
			panic(err)
		}

		c.Image = input[schema.AttributeTypeImage].(string)
		c.Cmd = input[schema.AttributeTypeCmd].([]string)
		c.Env = input[schema.AttributeTypeEnv].(map[string]string)

		err = c.Load()
		if err != nil {
			panic(err)
		}

		containerCache[c.Image] = c
	}

	result, err := c.Run()

	if err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{
			"result": result,
		},
	}

	return &o, nil
}
