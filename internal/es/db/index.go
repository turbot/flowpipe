package db

import (
	"strings"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

// Ristretto backed pipeline datatabase

func GetPipeline(name string) (*modconfig.Pipeline, error) {

	// TODO: hack while we're transitioning to mod format
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		name = "local.pipeline." + name
	}

	pipelineCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, perr.NotFoundWithMessage("pipeline definition not found: " + name)
	}

	pipeline, ok := pipelineCached.(*modconfig.Pipeline)
	if !ok {
		return nil, perr.InternalWithMessage("invalid pipeline")
	}
	return pipeline, nil
}

func ListAllPipelines() ([]*modconfig.Pipeline, error) {
	pipelineNamesCached, found := cache.GetCache().Get("#pipeline.names")
	if !found {
		return nil, perr.NotFoundWithMessage("pipeline names not found")
	}

	pipelineNames, ok := pipelineNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid pipeline names")
	}

	var pipelines []*modconfig.Pipeline
	for _, name := range pipelineNames {
		pipeline, err := GetPipeline(name)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, pipeline)
	}

	return pipelines, nil
}

func GetIntegration(name string) (modconfig.Integration, error) {

	integrationCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, perr.NotFoundWithMessage("integration definition not found: " + name)
	}

	integration, ok := integrationCached.(modconfig.Integration)
	if !ok {
		return nil, perr.InternalWithMessage("invalid integration")
	}
	return integration, nil
}

func ListAllIntegrations() ([]modconfig.Integration, error) {
	integrationNamesCached, found := cache.GetCache().Get("#integration.names")
	if !found {
		return nil, perr.NotFoundWithMessage("integration names not found")
	}

	integrationNames, ok := integrationNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("integration name cached is not a list of string")
	}

	var integrations []modconfig.Integration
	for _, name := range integrationNames {
		integration, err := GetIntegration(name)
		if err != nil {
			return nil, err
		}
		integrations = append(integrations, integration)
	}

	return integrations, nil
}

func ListAllNotifiers() ([]modconfig.Notifier, error) {
	notifierNamesCached, found := cache.GetCache().Get("#notifier.names")
	if !found {
		return nil, perr.NotFoundWithMessage("notifier names not found")
	}

	notifierNames, ok := notifierNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid notifier names")
	}

	var notifiers []modconfig.Notifier
	for _, name := range notifierNames {
		notifier, err := GetNotifier(name)
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, notifier)
	}

	return notifiers, nil
}

func GetTrigger(name string) (*modconfig.Trigger, error) {
	triggerCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, perr.NotFoundWithMessage("trigger definition not found: " + name)
	}

	trigger, ok := triggerCached.(*modconfig.Trigger)
	if !ok {
		return nil, perr.InternalWithMessage("invalid trigger")
	}

	return trigger, nil
}

func GetNotifier(name string) (modconfig.Notifier, error) {
	notifierCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, perr.NotFoundWithMessage("notifier definition not found: " + name)
	}

	notifier, ok := notifierCached.(modconfig.Notifier)
	if !ok {
		return nil, perr.InternalWithMessage("invalid notifier")
	}

	return notifier, nil
}

func ListAllTriggers() ([]modconfig.Trigger, error) {

	triggerNamesCached, found := cache.GetCache().Get("#trigger.names")
	if !found {
		return nil, perr.NotFoundWithMessage("trigger names not found")
	}

	triggerNames, ok := triggerNamesCached.([]string)
	if !ok {
		return nil, perr.InternalWithMessage("invalid trigger names")
	}

	var triggers []modconfig.Trigger
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
	flowpipeConfigCached, found := cache.GetCache().Get("#flowpipeconfig")

	if !found {
		return flowpipeconfig.NewFlowpipeConfig(), nil
	}

	flowpipeConfig, ok := flowpipeConfigCached.(*flowpipeconfig.FlowpipeConfig)
	if !ok {
		return nil, perr.InternalWithMessage("invalid flowpipe config")
	}

	return flowpipeConfig, nil
}
