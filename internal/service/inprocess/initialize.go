package inprocess

import (
	"context"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
	"time"
)

func Initialize(ctx context.Context) (*es.ESService, error) {
	// We use the cache to store the pipelines
	cache.InMemoryInitialize(nil)

	w, errorAndWarning := workspace.LoadWorkspacePromptingForVariables(ctx, viper.GetString(constants.ArgModLocation), ".hcl", ".sp")
	// TODO kai what about warnings
	if errorAndWarning.Error != nil {
		return nil, errorAndWarning.Error
	}

	pipelines := workspace.GetWorkspaceResourcesOfType[*modconfig.Pipeline](w)
	triggers := workspace.GetWorkspaceResourcesOfType[*modconfig.Trigger](w)

	cache.GetCache().SetWithTTL("#rootmod.name", w.Mod.Name(), 24*7*52*99*time.Hour)
	err := trigger.CachePipelinesAndTriggers(pipelines, triggers)
	error_helpers.FailOnErrorWithMessage(errorAndWarning.Error, "failed to cache pipelines and triggers")

	// create a watermill(?) service
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
