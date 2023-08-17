package db

import (
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

// Ristretto backed pipeline datatabase

func GetPipeline(name string) (*modconfig.Pipeline, error) {
	pipelineCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, pcerr.NotFoundWithMessage("pipeline not found: " + name)
	}

	pipeline, ok := pipelineCached.(*modconfig.Pipeline)
	if !ok {
		return nil, pcerr.InternalWithMessage("invalid pipeline")
	}
	return pipeline, nil
}

func ListAllPipelines() ([]modconfig.Pipeline, error) {

	pipelineNamesCached, found := cache.GetCache().Get("#pipeline.names")
	if !found {
		return nil, pcerr.NotFoundWithMessage("pipeline names not found")
	}

	pipelineNames, ok := pipelineNamesCached.([]string)
	if !ok {
		return nil, pcerr.InternalWithMessage("invalid pipeline names")
	}

	var pipelines []modconfig.Pipeline
	for _, name := range pipelineNames {
		pipeline, err := GetPipeline(name)
		if err != nil {
			return nil, err
		}
		pipelines = append(pipelines, *pipeline)
	}

	return pipelines, nil
}

func GetTrigger(name string) (*modconfig.Trigger, error) {
	triggerCached, found := cache.GetCache().Get(name)
	if !found {
		return nil, pcerr.NotFoundWithMessage("trigger not found: " + name)
	}

	trigger, ok := triggerCached.(modconfig.ITrigger)
	if !ok {
		return nil, pcerr.InternalWithMessage("invalid trigger")
	}
	result := modconfig.Trigger{
		Name:        trigger.GetName(),
		Description: trigger.GetDescription(),
		Args:        trigger.GetArgs(),
		Pipeline:    trigger.GetPipeline(),
	}
	return &result, nil
}

func ListAllTriggers() ([]modconfig.Trigger, error) {

	triggerNamesCached, found := cache.GetCache().Get("#trigger.names")
	if !found {
		return nil, pcerr.NotFoundWithMessage("trigger names not found")
	}

	triggerNames, ok := triggerNamesCached.([]string)
	if !ok {
		return nil, pcerr.InternalWithMessage("invalid trigger names")
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
