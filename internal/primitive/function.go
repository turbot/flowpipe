package primitive

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/turbot/flowpipe/internal/docker"
	function "github.com/turbot/flowpipe/internal/functions"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

var functionCache = map[string]*function.Function{}

var functionCacheMutex sync.Mutex

type Function struct {
	ModPath string
}

func (e *Function) ValidateInput(ctx context.Context, i modconfig.Input) error {
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

	return nil
}

func (e *Function) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	newEnvs := map[string]string{}

	start := time.Now().UTC()

	// This must be set outside the function schema
	if input[schema.AttributeTypeEnv] != nil {
		newEnvs = convertMapToStrings(input[schema.AttributeTypeEnv].(map[string]interface{}))
	}
	functionCacheMutex.Lock()
	fn := functionCache[input[schema.LabelName].(string)]

	if fn != nil {
		slog.Info("Found cached function, checking cached function env variables", "name", fn.Name)

		less := func(a, b string) bool { return a < b }
		equalIgnoreOrder := cmp.Diff(newEnvs, fn.Env, cmpopts.SortSlices(less)) == ""

		if !equalIgnoreOrder {
			slog.Info("Cached function env variables are different, rebuilding function", "name", fn.Name)
			fn = nil
			delete(functionCache, input[schema.LabelName].(string))
		} else {
			slog.Info("Cached function env variables are the same, using cached function", "name", fn.Name)
		}
	}
	functionCacheMutex.Unlock()

	if fn == nil {
		var err error
		fn, err = function.New(
			// ! Docker breaks if we use Gin's context. So we pass in a context.Background() that will be used
			// ! by the Docker client and Flowpipe context for logging purpose.
			function.WithContext(context.Background()),
			function.WithRunContext(ctx),
			function.WithDockerClient(docker.GlobalDockerClient),
			function.WithName(input[schema.LabelName].(string)),
			function.WithRuntime(input[schema.AttributeTypeRuntime].(string)),
			function.WithBasePath(e.ModPath),

			// TODO: support in passing the Lambda function support timeout. Needs to either be added in the Pipeline Step definition or in Flowpipe config
			function.WithStartTimeoutInSeconds(30),
		)
		if err != nil {
			return nil, err
		}

		if input[schema.AttributeTypeHandler] != nil {
			fn.Handler = input[schema.AttributeTypeHandler].(string)
		}
		fn.Source = input[schema.AttributeTypeSource].(string)

		fn.Env = newEnvs

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
			fn.Timeout = &timeoutInSeconds
		}

		err = fn.Load()
		if err != nil {
			return nil, err
		}

		functionCacheMutex.Lock()
		functionCache[fn.Name] = fn
		functionCacheMutex.Unlock()
	}

	if input[schema.AttributeTypeEvent] != nil {
		fn.Event = input[schema.AttributeTypeEvent].(map[string]interface{})
	}

	finish := time.Now().UTC()

	body := "{}"
	if len(fn.Event) > 0 {
		// Convert event body to JSON String
		jsonString, err := json.Marshal(fn.Event)
		if err != nil {
			slog.Error("Unable to convert Event body to JSON", "error", err.Error())
			return nil, perr.BadRequestWithMessage("Unable to convert Event body to JSON: " + err.Error())
		}
		body = string(jsonString)
	}

	statusCode, result, err := fn.Invoke([]byte(body))
	if err != nil {
		return nil, err
	}

	// Create an instance of the struct
	var responseJson map[string]interface{}

	// Unmarshal the JSON string into the struct
	err = json.Unmarshal(result, &responseJson)
	if err != nil {
		return nil, err
	}

	// Guess if the result is actually an error
	if responseJson["errorType"] != nil && responseJson["errorMessage"] != nil && responseJson["trace"] != nil {
		slog.Error("Function returned an error", "errorType", responseJson["errorType"], "errorMessage", responseJson["errorMessage"], "trace", responseJson["trace"])
		return nil, perr.InternalWithMessage("Function returned an error: " + responseJson["errorMessage"].(string))
	}

	output := modconfig.Output{
		Data: map[string]interface{}{
			schema.AttributeTypeResponse:   responseJson,
			schema.AttributeTypeStatusCode: statusCode,
		},
	}

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	return &output, nil
}
