package inprocess

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
	"time"
)

func Initialize(ctx context.Context) (*es.ESService, error) {
	// We use the cache to store the pipelines
	cache.InMemoryInitialize(nil)

	var pipelines = map[string]*modconfig.Pipeline{}
	var triggers = map[string]*modconfig.Trigger{}
	var rootModName string
	modLocation := viper.GetString(constants.ArgModLocation)
	if load_mod.ModFileExists(modLocation, app_specific.ModFileName) {
		w, errorAndWarning := workspace.LoadWorkspacePromptingForVariables(ctx, modLocation, ".hcl", ".sp")
		// TODO kai what about warnings
		if errorAndWarning.Error != nil {
			return nil, errorAndWarning.Error
		}
		rootModName = w.Mod.Name()

		pipelines = workspace.GetWorkspaceResourcesOfType[*modconfig.Pipeline](w)
		triggers = workspace.GetWorkspaceResourcesOfType[*modconfig.Trigger](w)
	} else {
		// TODO remove this when having a mod is mandatory <mandatory mod>
		var err error
		pipelines, triggers, err = load_mod.LoadPipelines(ctx, modLocation)
		if err != nil {
			return nil, err
		}
		rootModName = "local"
	}

	cache.GetCache().SetWithTTL("#rootmod.name", rootModName, 24*7*52*99*time.Hour)
	err := trigger.CachePipelinesAndTriggers(pipelines, triggers)
	error_helpers.FailOnErrorWithMessage(err, "failed to cache pipelines and triggers")

	// create the event sourcing service
	esService, err := es.NewESService(ctx)
	if err != nil {
		return nil, err
	}
	err = esService.Start()
	if err != nil {
		return nil, err
	}
	esService.Status = "running"
	esService.StartedAt = utils.TimeNow()

	// TODO should this just be in the test code??
	// Give some time for Watermill to fully start
	time.Sleep(2 * time.Second)
	return esService, nil
}
