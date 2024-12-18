package primitive

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/turbot/flowpipe/internal/container"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type Container struct {
	FullyQualifiedStepName string
}

var containerCache = map[string]*container.Container{}
var containerCacheMutex sync.Mutex

func (cp *Container) ValidateInput(ctx context.Context, i resources.Input) error {

	// Validate the name attribute
	if i[schema.LabelName] == nil {
		return perr.BadRequestWithMessage("Container input must define '" + schema.LabelName + "'")
	}
	if _, ok := i[schema.LabelName].(string); !ok {
		return perr.BadRequestWithMessage("Container attribute '" + schema.LabelName + "' must be a string")
	}

	// Validate the image attribute
	if i[schema.AttributeTypeImage] == nil && i[schema.AttributeTypeSource] == nil {
		return perr.BadRequestWithMessage("Container input must define '" + schema.AttributeTypeImage + "' or '" + schema.AttributeTypeSource + "', but not both")
	}
	if i[schema.AttributeTypeImage] != nil {
		if _, ok := i[schema.AttributeTypeImage].(string); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeImage + "' must be a string")
		}
	} else {
		if _, ok := i[schema.AttributeTypeSource].(string); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeSource + "' must be a string")
		}
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
	if i[schema.AttributeTypeEntrypoint] != nil {
		if _, ok := i[schema.AttributeTypeEntrypoint].([]interface{}); !ok {
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeEntrypoint + "' must be an array of strings")
		}
	}

	// Validate the timeout attribute
	if i[schema.AttributeTypeTimeout] != nil {
		switch duration := i[schema.AttributeTypeTimeout].(type) {
		case string:
			_, err := time.ParseDuration(duration)
			if err != nil {
				return perr.BadRequestWithMessage("invalid sleep duration " + duration)
			}
		case int64:
			if duration < 0 {
				return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a positive whole number")
			}
		case float64:
			if duration < 0 {
				return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a positive whole number")
			}
		default:
			return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a string or a whole number")
		}
	}

	// Validate the cpu shares attribute
	if i[schema.AttributeTypeCpuShares] != nil {
		switch i[schema.AttributeTypeCpuShares].(type) {
		case float64, int64:
		default:
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeCpuShares + "' must be an integer")
		}
	}

	// Validate the memory attribute
	if i[schema.AttributeTypeMemory] != nil {
		switch i[schema.AttributeTypeMemory].(type) {
		case float64, int64:
		default:
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemory + "' must be an integer")
		}
	}

	// Validate the memory reservation attribute
	if i[schema.AttributeTypeMemoryReservation] != nil {
		switch i[schema.AttributeTypeMemoryReservation].(type) {
		case float64, int64:
		default:
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemoryReservation + "' must be an integer")
		}
	}

	// Validate the memory swap attribute
	if i[schema.AttributeTypeMemorySwap] != nil {
		switch i[schema.AttributeTypeMemorySwap].(type) {
		case float64, int64:
		default:
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemorySwap + "' must be an integer")
		}
	}

	// Validate the memory swappiness attribute
	if i[schema.AttributeTypeMemorySwappiness] != nil {
		switch i[schema.AttributeTypeMemorySwappiness].(type) {
		case float64, int64:
		default:
			return perr.BadRequestWithMessage("Container attribute '" + schema.AttributeTypeMemorySwappiness + "' must be an integer")
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

func (cp *Container) Run(ctx context.Context, input resources.Input) (*resources.Output, error) {
	if err := cp.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	start := time.Now().UTC()

	c, err := cp.getFromCacheOrNew(ctx, input, cp.FullyQualifiedStepName)
	if err != nil {
		return nil, err
	}

	cConfig := container.ContainerRunConfig{}

	if input[schema.AttributeTypeCmd] != nil {
		cConfig.Cmd = convertToSliceOfString(input[schema.AttributeTypeCmd].([]interface{}))
	}

	if input[schema.AttributeTypeEnv] != nil {
		cConfig.Env = convertMapToStrings(input[schema.AttributeTypeEnv].(map[string]interface{}))
	}

	if input[schema.AttributeTypeEntrypoint] != nil {
		cConfig.EntryPoint = convertToSliceOfString(input[schema.AttributeTypeEntrypoint].([]interface{}))
	}

	if input[schema.AttributeTypeUser] != nil {
		cConfig.User = input[schema.AttributeTypeUser].(string)
	}

	if input[schema.AttributeTypeWorkdir] != nil {
		cConfig.Workdir = input[schema.AttributeTypeWorkdir].(string)
	}

	if input[schema.AttributeTypeReadOnly] != nil {
		readOnly := input[schema.AttributeTypeReadOnly].(bool)
		cConfig.ReadOnly = &readOnly
	}

	if input[schema.AttributeTypeTimeout] != nil {
		var timeout time.Duration
		switch timeoutDuration := input[schema.AttributeTypeTimeout].(type) {
		case string:
			timeout, _ = time.ParseDuration(timeoutDuration)
		case int64:
			timeout = time.Duration(timeoutDuration) * time.Millisecond // in milliseconds
		case float64:
			timeout = time.Duration(timeoutDuration) * time.Millisecond // in milliseconds
		}
		timeoutInMs := timeout.Milliseconds()

		// Convert milliseconds to seconds, and round up to the nearest second
		timeoutInSeconds := int64(math.Ceil(float64(timeoutInMs) / 1000))
		cConfig.Timeout = &timeoutInSeconds
	}

	if input[schema.AttributeTypeCpuShares] != nil {
		var cpuShares int64
		switch c := input[schema.AttributeTypeCpuShares].(type) {
		case float64:
			cpuShares = int64(c)
		case int64:
			cpuShares = c
		default:
			break
		}
		cConfig.CpuShares = &cpuShares
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
		cConfig.Memory = &memory
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
		cConfig.MemoryReservation = &memoryReservation
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
		cConfig.MemorySwap = &memorySwap
	}

	// Check if DOCKER_HOST is not set or does not end with "podman.sock"
	dockerHost, exists := os.LookupEnv("DOCKER_HOST")
	if !exists || !strings.HasSuffix(dockerHost, "podman.sock") {
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
			cConfig.MemorySwappiness = &memorySwappiness
		}
	}

	// Construct the output
	output := resources.Output{
		Data: map[string]interface{}{},
	}

	containerID, exitCode, err := c.Run(cConfig)
	if err != nil {
		if e, ok := err.(perr.ErrorModel); !ok {
			output.Errors = []resources.StepError{
				{
					Error: perr.InternalWithMessage("Error loading function config: " + err.Error()),
				},
			}
		} else {
			output.Errors = []resources.StepError{
				{
					Error: e,
				},
			}
		}
		output.Status = "failed"
	} else {
		output.Status = "finished"
	}
	finish := time.Now().UTC()

	output.Data[schema.AttributeTypeContainerId] = containerID
	output.Data[schema.AttributeTypeExitCode] = exitCode

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	// If there are any error while creating the container, then the containerID will be empty
	if c.Runs[containerID] != nil {
		output.Data[schema.AttributeTypeStdout] = c.Runs[containerID].Stdout
		output.Data[schema.AttributeTypeStderr] = c.Runs[containerID].Stderr
		output.Data[schema.AttributeTypeLines] = c.Runs[containerID].Lines
	}

	return &output, nil
}

func (cp *Container) getFromCacheOrNew(ctx context.Context, input resources.Input, stepFullName string) (*container.Container, error) {
	c := containerCache[stepFullName]

	// if Dockerfile source path changed ignore cache & rebuild
	if c != nil && input[schema.AttributeTypeSource] != nil && c.Source != input[schema.AttributeTypeSource].(string) {
		c = nil
	}

	// if image has been changed in input ignore cache and create new reference
	if c != nil && input[schema.AttributeTypeImage] != nil && c.Image != input[schema.AttributeTypeImage].(string) {
		c = nil
	}

	if c != nil {
		return c, nil
	}

	containerCacheMutex.Lock()
	defer containerCacheMutex.Unlock()

	c, err := container.NewContainer(
		container.WithContext(context.Background()),
		container.WithRunContext(ctx),
		container.WithDockerClient(docker.GlobalDockerClient),
		container.WithName(stepFullName),
	)
	if err != nil {
		return nil, perr.InternalWithMessage("Error creating container config with the provided options:" + err.Error())
	}

	if input[schema.AttributeTypeImage] != nil {
		c.Image = input[schema.AttributeTypeImage].(string)
	}

	if input[schema.AttributeTypeSource] != nil {
		c.Source = input[schema.AttributeTypeSource].(string)
	}

	err = c.Load()
	if err != nil {
		return nil, perr.InternalWithMessage("failed loading container config: " + err.Error())
	}

	containerCache[stepFullName] = c

	return c, nil
}
