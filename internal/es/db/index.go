package db

import (
	"github.com/turbot/flowpipe/internal/flowpipeconfig"
	flowpipe2 "github.com/turbot/flowpipe/internal/resources"
	"reflect"
	"strings"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

//
// TODO: we seem to cache resources AND cache the Mod and FlowpipeConfig object in memory. This is a duplicate, I think we should drop the cache
// TODO: and use the cached FlowpipeConfig and Mod struct
//

func typeName[T any](t T) string {
	if !helpers.IsNil(t) {
		return reflect.TypeOf(t).String()
	}

	return "unknown"
}

func GetCachedItem[T any](name string) (T, error) {
	var defaultT T // default zero value for type T

	// Special handling for pipeline names
	if _, ok := any(defaultT).(*flowpipe2.Pipeline); ok {
		parts := strings.Split(name, ".")
		if len(parts) == 1 {
			name = "local.pipeline." + name
		}
	}

	fpCache := cache.GetCache()
	if fpCache == nil {
		return defaultT, perr.InternalWithMessage("cache not initialized")
	}

	cachedItem, found := fpCache.Get(name)
	if !found {
		return defaultT, perr.NotFoundWithMessage(typeName(defaultT) + " definition not found: " + name)
	}

	item, ok := cachedItem.(T)
	if !ok {
		return defaultT, perr.InternalWithMessage("invalid " + typeName(defaultT))
	}

	return item, nil
}

func GetNotifier(name string) (flowpipe2.Notifier, error) {
	return GetCachedItem[flowpipe2.Notifier](name)
}

func GetIntegration(name string) (flowpipe2.Integration, error) {
	return GetCachedItem[flowpipe2.Integration](name)
}

func GetVariable(name string) (*modconfig.Variable, error) {
	return GetCachedItem[*modconfig.Variable](name)
}

func GetPipelineWithModFullVersion(modFullVersion, name string) (*flowpipe2.Pipeline, error) {
	if modFullVersion == "" {
		return GetPipeline(name)
	}
	p, err := GetCachedItem[*flowpipe2.Pipeline](modFullVersion + "." + name)
	if perr.IsNotFound(err) {
		return GetPipeline(name)
	}
	return p, err
}

func GetPipeline(name string) (*flowpipe2.Pipeline, error) {
	return GetCachedItem[*flowpipe2.Pipeline](name)
}

func GetPipelineResolvedFromMod(mod *modconfig.Mod, name string) (*flowpipe2.Pipeline, error) {

	// check if the pipeline is coming from the given mod
	pipelineParts := strings.Split(name, ".")
	if len(pipelineParts) != 3 {
		return nil, perr.BadRequestWithMessage("invalid pipeline name: " + name)
	}

	pipelineModName := pipelineParts[0]

	if pipelineParts[0] == "local" {
		return GetCachedItem[*flowpipe2.Pipeline](name)
	}

	// check if it's coming from the current mod
	if pipelineModName == mod.ModName {
		return GetPipelineFromCurrentMod(mod, name)
	}

	// If not check if the mod in the pipeline is a dependent of the given mod
	//
	// Don't recurse because you can only call a pipeline from a mod that is a direct dependency
	// not a dependency of a dependency
	for _, m := range mod.ResourceMaps.GetMods() {
		if m.ModName == pipelineModName {
			return GetPipelineFromCurrentMod(m, name)
		}
	}

	return nil, perr.NotFoundWithMessage("pipeline not found: " + name + " from mod " + mod.Name())
}

func GetPipelineFromCurrentMod(mod *modconfig.Mod, name string) (*flowpipe2.Pipeline, error) {
	if mod == nil {
		return nil, perr.BadRequestWithMessage("mod is nil")
	}

	prefixCacheKey := mod.Name()
	if mod.Version != nil {
		prefixCacheKey += "." + mod.Version.String()
	}

	cacheKey := prefixCacheKey + "." + name

	return GetCachedItem[*flowpipe2.Pipeline](cacheKey)
}

func GetTrigger(name string) (*flowpipe2.Trigger, error) {
	return GetCachedItem[*flowpipe2.Trigger](name)
}

func ListAllPipelines() ([]*flowpipe2.Pipeline, error) {
	pipelineNamesCached, found := cache.GetCache().Get("#pipeline.names")
	if !found {
		return nil, perr.NotFoundWithMessage("pipeline names not found")
	}

	pipelineNames, ok := pipelineNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid pipeline names")
	}

	var pipelines []*flowpipe2.Pipeline
	for _, name := range pipelineNames {
		pipeline, err := GetPipeline(name)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

func ListAllIntegrations() ([]flowpipe2.Integration, error) {
	integrationNamesCached, found := cache.GetCache().Get("#integration.names")
	if !found {
		return nil, perr.NotFoundWithMessage("integration names not found")
	}

	integrationNames, ok := integrationNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("integration name cached is not a list of string")
	}

	var integrations []flowpipe2.Integration
	for _, name := range integrationNames {
		integration, err := GetIntegration(name)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}

	return integrations, nil
}

func ListAllVariables() ([]*modconfig.Variable, error) {
	variableNamesCached, found := cache.GetCache().Get("#variable.names")
	if !found {
		return nil, perr.NotFoundWithMessage("variable names not found")
	}

	variableNames, ok := variableNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid variable names")
	}

	var variables []*modconfig.Variable
	for _, name := range variableNames {
		variable, err := GetVariable(name)
		if err != nil {
			return nil, err
		}
		variables = append(variables, variable)
	}

	return variables, nil

}

func ListAllNotifiers() ([]flowpipe2.Notifier, error) {
	notifierNamesCached, found := cache.GetCache().Get("#notifier.names")
	if !found {
		return nil, perr.NotFoundWithMessage("notifier names not found")
	}

	notifierNames, ok := notifierNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid notifier names")
	}

	var notifiers []flowpipe2.Notifier
	for _, name := range notifierNames {
		notifier, err := GetNotifier(name)
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, notifier)
	}

	return notifiers, nil
}

func ListAllTriggers() ([]flowpipe2.Trigger, error) {

	triggerNamesCached, found := cache.GetCache().Get("#trigger.names")
	if !found {
		return nil, perr.NotFoundWithMessage("trigger names not found")
	}

	triggerNames, ok := triggerNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid trigger names")
	}

	var triggers []flowpipe2.Trigger
	for _, name := range triggerNames {
		trigger, err := GetTrigger(name)
		if err != nil {
			return nil, err
		}
		triggers = append(triggers, *trigger)
	}

	return triggers, nil
}

func GetFlowpipeConfig() (*flowpipeconfig.FlowpipeConfig, error) {
	flowpipeConfigCached, found := cache.GetCache().Get(constants.FlowpipeConfigCacheKey)

	if !found {
		// TODO: if we return an error all our "non mod based test" fail
		// return nil, perr.BadRequestWithMessage("flowpipe config not found")
		return &flowpipeconfig.FlowpipeConfig{}, nil
	}

	flowpipeConfig, ok := flowpipeConfigCached.(*flowpipeconfig.FlowpipeConfig)
	if !ok {
		return nil, perr.InternalWithMessage("invalid flowpipe config")
	}

	return flowpipeConfig, nil
}
