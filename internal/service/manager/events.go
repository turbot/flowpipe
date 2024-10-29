package manager

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/flowpipeconfig"
	fpparse "github.com/turbot/flowpipe/internal/parse"
	flowpipe2 "github.com/turbot/flowpipe/internal/resources"
	"log/slog"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	fpconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/sanitize"
	"github.com/turbot/pipe-fittings/workspace"
)

func (m *Manager) flowpipeConfigUpdated(ctx context.Context, newFpConfig *flowpipeconfig.FlowpipeConfig) {
	m.fpConfigLoadLock.Lock()
	defer m.fpConfigLoadLock.Unlock()

	slog.Debug("flowpipe config updated")

	if newFpConfig == nil {
		slog.Debug("flowpipe config is nil")
		return
	}

	cache.ResetCredentialCache()

	cache.GetCache().SetWithTTL(fpconstants.FlowpipeConfigCacheKey, newFpConfig, 24*7*52*99*time.Hour)

	err := m.cacheConfigData()
	if err != nil {
		slog.Error("error caching config data", "error", err)
		return
	}

	err = m.loadMod()
	if err != nil {
		slog.Error("error loading mod", "error", err)
		return
	}
}

func (m *Manager) modUpdated() {
	m.rootModLoadLock.Lock()
	defer m.rootModLoadLock.Unlock()

	// At this point the w.Mod has already been updated, the code that does it is in pipe-fittings handleFileWatcherEvent function
	m.RootMod = m.workspace.Mod

	// get resources from mod
	resourceMaps := flowpipe2.GetModResources(m.RootMod)

	var serverOutput []sanitize.SanitizedStringer
	var err error
	slog.Info("caching pipelines and triggers")
	serverOutput = append(serverOutput, types.NewServerOutputLoaded(types.NewServerOutputPrefix(time.Now(), "flowpipe"), m.RootMod.Name(), true))
	m.triggers = resourceMaps.Triggers
	err = m.cacheModData(m.RootMod)
	if err != nil {
		slog.Error("error caching pipelines and triggers", "error", err)
		serverOutput = append(serverOutput, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "Failed caching pipelines and triggers", err))
	} else {
		slog.Info("cached pipelines and triggers")
		serverOutput = append(serverOutput, types.NewServerOutput(time.Now(), "flowpipe", "Cached pipelines and triggers"))
		m.apiService.ModMetadata.IsStale = false
		m.apiService.ModMetadata.LastLoaded = time.Now()
	}

	// Reload scheduled triggers
	slog.Info("rescheduling triggers")
	if m.schedulerService != nil {
		m.schedulerService.Triggers = resourceMaps.Triggers
		err := m.schedulerService.RescheduleTriggers()
		if err != nil {
			slog.Error("error rescheduling triggers", "error", err)
			serverOutput = append(serverOutput, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "Failed rescheduling triggers", err))
		} else {
			slog.Info("rescheduled triggers")
			serverOutput = append(serverOutput, types.NewServerOutput(time.Now(), "flowpipe", "Rescheduled triggers"))
			serverOutput = append(serverOutput, renderServerTriggers(m.triggers)...)
		}
	}

	if output.IsServerMode {
		output.RenderServerOutput(m.ctx, serverOutput...)
	}
}

func (m *Manager) setupWatcher(w *workspace.Workspace) error {
	if !viper.GetBool(constants.ArgWatch) {
		return nil
	}

	err := w.SetupWatcher(m.ctx, func(c context.Context, e error) {
		slog.Error("error watching workspace", "error", e)
		if output.IsServerMode {
			output.RenderServerOutput(c, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), fmt.Sprintf("Failed watching workspace for mod %s", w.Mod.Name()), e))
		}
		m.apiService.ModMetadata.IsStale = true
	})

	if err != nil {
		return err
	}

	w.SetOnFileWatcherEventMessages(m.modUpdated)
	return nil
}

func (m *Manager) loadMod() error {
	modLocation := viper.GetString(constants.ArgModLocation)

	flowpipeConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		slog.Error("error getting flowpipe config", "error", err)
		return err
	}

	w, errorAndWarning := workspace.LoadWorkspacePromptingForVariables(
		m.ctx,
		modLocation,
		workspace.WithPipelingConnections(flowpipeConfig.PipelingConnections),
		workspace.WithDecoderOptions(fpparse.WithCredentials(flowpipeConfig.Credentials)),
	)

	if errorAndWarning.Error != nil {
		return errorAndWarning.Error
	}

	m.workspace = w

	// if we are running in server mode, setup the file watcher
	if m.shouldStartAPI() {
		if err := m.setupWatcher(w); err != nil {
			return err
		}
	}

	mod := w.Mod
	m.RootMod = w.Mod

	if mod.Require != nil && mod.Require.Flowpipe != nil && mod.Require.Flowpipe.Constraint != nil {
		flowpipeCliVersion := viper.GetString("main.version")
		flowpipeSemverVersion := semver.MustParse(flowpipeCliVersion)
		if !mod.Require.Flowpipe.Constraint.Check(flowpipeSemverVersion) {
			return perr.BadRequestWithMessage(fmt.Sprintf("flowpipe version %s does not satisfy %s which requires version %s", flowpipeCliVersion, mod.ShortName, mod.Require.Flowpipe.MinVersionString))
		}
	}

	m.triggers = workspace.GetWorkspaceResourcesOfType[*flowpipe2.Trigger](w)

	cache.GetCache().SetWithTTL("#rootmod.name", mod.ShortName, 24*7*52*99*time.Hour)
	err = m.cacheModData(mod)
	if err != nil {
		return err
	}

	slog.Info("loaded mod", "mod", mod.Name())
	return nil
}
