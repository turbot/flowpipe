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

	// Validate the name attribute
	if i[schema.LabelName] == nil {
		return perr.BadRequestWithMessage("Container input must define '" + schema.LabelName + "'")
	}
	if _, ok := i[schema.LabelName].(string); !ok {
		return perr.BadRequestWithMessage("Container attribute '" + schema.LabelName + "' must be a string")
	}

	// Validate the image attribute
	if i[schema.AttributeTypeImage] == nil {
		return perr.BadRequestWithMessage("Container input must define '" + schema.AttributeTypeImage + "'")
	}
	if _, ok := i[schema.AttributeTypeImage].(string); !ok {
		return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeImage + "' must be a string")
	}

	// Validate the cmd attribute
	if i[schema.AttributeTypeCmd] != nil {
		if _, ok := i[schema.AttributeTypeCmd].([]interface{}); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeCmd + "' must be an array of strings")
		}
	}

	// Validate the env
	if i[schema.AttributeTypeEnv] != nil {
		if _, ok := i[schema.AttributeTypeEnv].(map[string]interface{}); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeEnv + "' must be a map of strings")
		}
	}

	// Validate the entrypoint attribute
	if i[schema.AttributeTypeEntryPoint] != nil {
		if _, ok := i[schema.AttributeTypeEntryPoint].([]interface{}); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeEntryPoint + "' must be an array of strings")
		}
	}

	// Validate the timeout attribute
	if i[schema.AttributeTypeTimeout] != nil {
		if _, ok := i[schema.AttributeTypeTimeout].(int64); !ok {
			switch i[schema.AttributeTypeTimeout].(type) {
			case float64, int64:
			default:
				return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeTimeout + "' must be an integer")
			}
		}
	}

	// Validate the memory attribute
	if i[schema.AttributeTypeMemory] != nil {
		if _, ok := i[schema.AttributeTypeMemory].(int64); !ok {
			switch i[schema.AttributeTypeMemory].(type) {
			case float64, int64:
			default:
				return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemory + "' must be an integer")
			}
		}
	}

	// Validate the memory reservation attribute
	if i[schema.AttributeTypeMemoryReservation] != nil {
		if _, ok := i[schema.AttributeTypeMemoryReservation].(int64); !ok {
			switch i[schema.AttributeTypeMemoryReservation].(type) {
			case float64, int64:
			default:
				return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemoryReservation + "' must be an integer")
			}
		}
	}

	// Validate the memory swap attribute
	if i[schema.AttributeTypeMemorySwap] != nil {
		if _, ok := i[schema.AttributeTypeMemorySwap].(int64); !ok {
			switch i[schema.AttributeTypeMemorySwap].(type) {
			case float64, int64:
			default:
				return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemorySwap + "' must be an integer")
			}
		}
	}

	// Validate the memory swappiness attribute
	if i[schema.AttributeTypeMemorySwappiness] != nil {
		if _, ok := i[schema.AttributeTypeMemorySwappiness].(int64); !ok {
			switch i[schema.AttributeTypeMemorySwappiness].(type) {
			case float64, int64:
			default:
				return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemorySwappiness + "' must be an integer")
			}
		}
	}

	// Validate the read-only attribute
	if i[schema.AttributeTypeReadOnly] != nil {
		if _, ok := i[schema.AttributeTypeReadOnly].(bool); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeReadOnly + "' must be a boolean")
		}
	}

	// Validate the user attribute
	if i[schema.AttributeTypeUser] != nil {
		if _, ok := i[schema.AttributeTypeUser].(string); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeUser + "' must be a string")
		}
	}

	// Validate the workdir attribute
	if i[schema.AttributeTypeWorkdir] != nil {
		if _, ok := i[schema.AttributeTypeWorkdir].(string); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeWorkdir + "' must be a string")
		}
	}

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

	if input[schema.AttributeTypeUser] != nil {
		c.User = input[schema.AttributeTypeUser].(string)
	}

	if input[schema.AttributeTypeWorkdir] != nil {
		c.Workdir = input[schema.AttributeTypeWorkdir].(string)
	}

	if input[schema.AttributeTypeReadOnly] != nil {
		readOnly := input[schema.AttributeTypeReadOnly].(bool)
		c.ReadOnly = &readOnly
	}

	if input[schema.AttributeTypeTimeout] != nil {
		var timeout int64
		switch t := input[schema.AttributeTypeTimeout].(type) {
		case float64:
			timeout = int64(t)
		case int64:
			timeout = t
		default:
			break
		}
		c.Timeout = &timeout
	}

	if input[schema.AttributeTypeMemory] != nil {
		var memory int64
		switch m := input[schema.AttributeTypeMemory].(type) {
		case float64:
			memory = int64(m)
		case int64:
			memory = m
		default:
			break
		}
		c.Memory = &memory
	}

	if input[schema.AttributeTypeMemoryReservation] != nil {
		var memoryReservation int64
		switch mr := input[schema.AttributeTypeMemoryReservation].(type) {
		case float64:
			memoryReservation = int64(mr)
		case int64:
			memoryReservation = mr
		default:
			break
		}
		c.MemoryReservation = &memoryReservation
	}

	if input[schema.AttributeTypeMemorySwap] != nil {
		var memorySwap int64
		switch ms := input[schema.AttributeTypeMemorySwap].(type) {
		case float64:
			memorySwap = int64(ms)
		case int64:
			memorySwap = ms
		default:
			break
		}
		c.MemorySwap = &memorySwap
	}

	if input[schema.AttributeTypeMemorySwappiness] != nil {
		var memorySwappiness int64
		switch ms := input[schema.AttributeTypeMemorySwappiness].(type) {
		case float64:
			memorySwappiness = int64(ms)
		case int64:
			memorySwappiness = ms
		default:
			break
		}
		c.MemorySwappiness = &memorySwappiness
	}

	err = c.Load()
	if err != nil {
		return nil, perr.InternalWithMessage("Error loading function config: " + err.Error())
	}

	// Construct the output
	output := modconfig.Output{
		Data: map[string]interface{}{},
	}

	containerID, exitCode, err := c.Run()
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
	output.Data["stdout"] = c.Runs[containerID].Stdout
	output.Data["stderr"] = c.Runs[containerID].Stderr
	output.Data["lines"] = c.Runs[containerID].Lines
	output.Data["exit_code"] = exitCode

	return &output, nil
}
