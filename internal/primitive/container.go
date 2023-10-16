package primitive

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/internal/container"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

type Input struct{}

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{
			"container_id": "1234",
			"stdout":       "hello world",
			"stderr":       "",
		},
	}

	return &o, nil
}

type Container struct{}

func (e *Container) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func convertToSliceOfString(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		if str, ok := v.(string); ok {
			result[i] = str
		} else {
			// Handle the case where the element is not a string
			// You can choose to skip, convert, or handle it as needed.
			result[i] = fmt.Sprint(v)
		}
	}
	return result
}

func convertMapToStrings(input map[string]interface{}) map[string]string {
	result := make(map[string]string)

	for key, value := range input {
		if str, ok := value.(string); ok {
			result[key] = str
		} else {
			// Handle the case where the value is not a string
			// You can choose to skip, convert, or handle it as needed.
			result[key] = fmt.Sprint(value)
		}
	}

	return result
}

func (e *Container) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	c, err := container.NewContainer(
		container.WithContext(context.Background()),
		container.WithRunContext(ctx),
		container.WithDockerClient(docker.GlobalDockerClient),
	)
	if err != nil {
		panic(err)
	}

	c.Name = input[schema.LabelName].(string)
	c.Image = input[schema.AttributeTypeImage].(string)
	if input[schema.AttributeTypeCmd] != nil {
		c.Cmd = convertToSliceOfString(input[schema.AttributeTypeCmd].([]interface{}))
	}

	if input[schema.AttributeTypeEnv] != nil {
		c.Env = convertMapToStrings(input[schema.AttributeTypeEnv].(map[string]interface{}))
	}

	err = c.Load()
	if err != nil {
		panic(err)
	}

	containerID, err := c.Run()

	stdout := c.Runs[containerID].Stdout
	stderr := c.Runs[containerID].Stderr

	if err != nil {
		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{
			"container_id": containerID,
			"stdout":       stdout,
			"stderr":       stderr,
		},
	}

	return &o, nil
}
