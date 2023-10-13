package primitive

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fplog"
	function "github.com/turbot/flowpipe/internal/functions"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

var functionCache = map[string]*function.Function{}

var functionCacheMutex sync.Mutex

type Function struct{}

func (e *Function) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return nil
}

func (e *Function) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	logger := fplog.Logger(ctx)

	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	newEnvs := map[string]string{}

	// This must be set outside the function schema
	if input[schema.AttributeTypeEnv] != nil {
		newEnvs = convertMapToStrings(input[schema.AttributeTypeEnv].(map[string]interface{}))
	}
	functionCacheMutex.Lock()
	fn := functionCache[input[schema.LabelName].(string)]

	if fn != nil {
		logger.Info("Found cached function, checking cached function env variables", "name", fn.Name)

		less := func(a, b string) bool { return a < b }
		equalIgnoreOrder := cmp.Diff(newEnvs, fn.Env, cmpopts.SortSlices(less)) == ""

		if !equalIgnoreOrder {
			logger.Info("Cached function env variables are different, rebuilding function", "name", fn.Name)
			fn = nil
			delete(functionCache, input[schema.LabelName].(string))
		} else {
			logger.Info("Cached function env variables are the same, using cached function", "name", fn.Name)
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
		)
		if err != nil {
			return nil, err
		}
		fn.Name = input[schema.LabelName].(string)
		fn.Runtime = input[schema.AttributeTypeRuntime].(string)
		if input[schema.AttributeTypeHandler] != nil {
			fn.Handler = input[schema.AttributeTypeHandler].(string)
		}
		fn.Src = input[schema.AttributeTypeSrc].(string)

		fn.Env = newEnvs

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

	body := "{}"
	if len(fn.Event) > 0 {
		// Convert event body to JSON String
		jsonString, err := json.Marshal(fn.Event)
		if err != nil {
			logger.Error("Unable to convert Event body to JSON", "error", err.Error())
			return nil, perr.BadRequestWithMessage("Unable to convert Event body to JSON: " + err.Error())
		}
		body = string(jsonString)
	}

	statusCode, result, err := fn.Invoke([]byte(body))
	if err != nil {
		return nil, err
	}

	// Create an instance of the struct
	var resultsJson map[string]interface{}

	// Unmarshal the JSON string into the struct
	err = json.Unmarshal(result, &resultsJson)
	if err != nil {

		return nil, err
	}

	o := modconfig.Output{
		Data: map[string]interface{}{
			"result":      resultsJson,
			"status_code": statusCode,
		},
	}

	return &o, nil
}
