package db

import (
	"strings"

	"github.com/turbot/flowpipe/internal/cache"
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

func GetFlowpipeConfig() (*modconfig.FlowpipeConfig, error) {
	flowpipeConfigCached, found := cache.GetCache().Get("#flowpipeconfig")
	if !found {
		return nil, perr.NotFoundWithMessage("flowpipe config not found")
	}

	flowpipeConfig, ok := flowpipeConfigCached.(*modconfig.FlowpipeConfig)
	if !ok {
		return nil, perr.InternalWithMessage("invalid flowpipe config")
	}

	return flowpipeConfig, nil
}
