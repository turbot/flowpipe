package primitive

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/internal/container"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

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
		return nil, perr.InternalWithMessage("Error creating function config with the provided options:" + err.Error())
	}

	c.Name = input[schema.LabelName].(string)
	c.Image = input[schema.AttributeTypeImage].(string)
	if input[schema.AttributeTypeCmd] != nil {
		c.Cmd = convertToSliceOfString(input[schema.AttributeTypeCmd].([]interface{}))
	}

	if input[schema.AttributeTypeEnv] != nil {
		c.Env = convertMapToStrings(input[schema.AttributeTypeEnv].(map[string]interface{}))
	}

	if input[schema.AttributeTypeEntryPoint] != nil {
		c.EntryPoint = convertToSliceOfString(input[schema.AttributeTypeEntryPoint].([]interface{}))
	}

	err = c.Load()
	if err != nil {
		return nil, perr.InternalWithMessage("Error loading function config: " + err.Error())
	}

	// Construct the output
	output := modconfig.Output{
		Data: map[string]interface{}{},
	}

	containerID, exitCode, streamLines, err := c.Run()

	stdout := c.Runs[containerID].Stdout
	stderr := c.Runs[containerID].Stderr
	combined := c.Runs[containerID].Combined

	if err != nil {
		if e, ok := err.(perr.ErrorModel); !ok {
			output.Errors = []modconfig.StepError{
				{
					Error: perr.InternalWithMessage("Error loading function config: " + err.Error()),
				},
			}
		} else {
			output.Errors = []modconfig.StepError{
				{
					Error: e,
				},
			}
		}
		output.Status = "failed"
	} else {
		output.Status = "finished"
	}

	output.Data["container_id"] = containerID
	output.Data["stdout"] = stdout
	output.Data["stderr"] = stderr
	output.Data["combined"] = combined
	output.Data["exit_code"] = exitCode

	if streamLines != nil && len(streamLines) > 0 {
		output.Data["lines"] = streamLines
	}

	return &output, nil
}
