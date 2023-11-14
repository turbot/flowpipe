package load

import (
	"context"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/workspace"
)

func LoadWorkspace(ctx context.Context, pipelineDir string, modInfo *modconfig.Mod, pipelines map[string]*modconfig.Pipeline, triggers map[string]*modconfig.Trigger) (*modconfig.Mod, map[string]*modconfig.Pipeline, map[string]*modconfig.Trigger, error, bool) {
	w, errorAndWarning := workspace.LoadWorkspacePromptingForVariables(ctx, pipelineDir, ".hcl", ".sp")
	if errorAndWarning.Error != nil {
		return nil, nil, nil, errorAndWarning.Error, true
	}

	//err := w.SetupWatcher(ctx, func(c context.Context, e error) {
	//	logger := fplog.Logger(ctx)
	//	logger.Error("error watching workspace", "error", e)
	//	m.apiService.ModMetadata.IsStale = true
	//})
	//if err != nil {
	//	return nil, nil, nil, err, true
	//}
	//
	//w.SetOnFileWatcherEventMessages(func() {
	//	logger := fplog.Logger(ctx)
	//	logger.Info("caching pipelines and triggers")
	//	err := m.CachePipelinesAndTriggers(w.Mod.ResourceMaps.Pipelines, w.Mod.ResourceMaps.Triggers)
	//	if err != nil {
	//		logger.Error("error caching pipelines and triggers", "error", err)
	//	} else {
	//		logger.Info("cached pipelines and triggers")
	//		m.apiService.ModMetadata.IsStale = false
	//		m.apiService.ModMetadata.LastLoaded = time.Now()
	//	}
	//
	//	// Reload scheduled triggers
	//	logger.Info("rescheduling triggers")
	//	if m.schedulerService != nil {
	//		m.schedulerService.Triggers = w.Mod.ResourceMaps.Triggers
	//		err := m.schedulerService.RescheduleTriggers()
	//		if err != nil {
	//			logger.Error("error rescheduling triggers", "error", err)
	//		} else {
	//			logger.Info("rescheduled triggers")
	//		}
	//	}
	//})

	mod := w.Mod
	modInfo = mod

	pipelines = mod.ResourceMaps.Pipelines
	triggers = mod.ResourceMaps.Triggers

	for _, depMod := range mod.ResourceMaps.Mods {
		// tactical - resource maps contains parent mod
		if depMod.Name() != mod.Name() {
			for _, pipeline := range depMod.ResourceMaps.Pipelines {
				pipelines[pipeline.Name()] = pipeline
			}
			for _, trigger := range depMod.ResourceMaps.Triggers {
				triggers[trigger.Name()] = trigger
			}
		}
	}
	return modInfo, pipelines, triggers, nil, false
}
