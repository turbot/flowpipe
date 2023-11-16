package manager

import (
	"context"
	"fmt"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/scheduler"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
)

type ExecutionMode int

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	RootMod *modconfig.Mod

	apiService *api.APIService
	*es.ESService
	schedulerService *scheduler.SchedulerService

	triggers map[string]*modconfig.Trigger

	RaftNodeID    string
	RaftBootstrap bool
	RaftAddress   string

	HTTPAddress string
	HTTPPort    int

	// which services should we start
	startES  bool
	startAPI bool

	Status    string
	StartedAt *time.Time
	StoppedAt *time.Time
}

// NewManager creates a new Manager.
func NewManager(ctx context.Context, opts ...ManagerOption) *Manager {
	// Defaults
	m := &Manager{
		ctx:    ctx,
		Status: "initialized",
	}
	// Set options
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Start initializes tha manage and starts services managed by the Manager.
func (m *Manager) Start() (*Manager, error) {
	fplog.Logger(m.ctx).Debug("Manager starting")
	defer fplog.Logger(m.ctx).Debug("Manager started")

	// initialize - load and cache triggers and pipelines
	// if we are in server mode and there is a modfile, setup the file watcher
	if err := m.initialize(); err != nil {
		return nil, err
	}

	if m.startES {
		err := m.startESService()
		if err != nil {
			return nil, err
		}
	}

	if m.startAPI {
		if err := m.startAPIService(); err != nil {
			return nil, err
		}
		if err := m.startSchedulerService(); err != nil {
			return nil, err
		}
	}

	m.StartedAt = utils.TimeNow()
	m.Status = "running"

	return m, nil
}

// load and cache triggers and pipelines
// if we are in server mode and there is a modfile, setup the file watcher
func (m *Manager) initialize() error {
	logger := fplog.Logger(m.ctx)

	pipelineDir := viper.GetString(constants.ArgModLocation)
	logger.Info("Starting Flowpipe", "pipelineDir", pipelineDir)

	var pipelines = map[string]*modconfig.Pipeline{}
	var triggers = map[string]*modconfig.Trigger{}
	var modInfo *modconfig.Mod

	if load_mod.ModFileExists(pipelineDir, app_specific.ModFileName) {
		w, errorAndWarning := workspace.LoadWorkspacePromptingForVariables(m.ctx, pipelineDir, ".hcl", ".sp")
		if errorAndWarning.Error != nil {
			return errorAndWarning.Error
		}

		// if we are running in server mode, setup the file watcher
		if m.startAPI {
			if err := m.setupWatcher(w); err != nil {
				return err
			}
		}

		mod := w.Mod
		modInfo = mod

		pipelines = workspace.GetWorkspaceResourcesOfType[*modconfig.Pipeline](w)
		triggers = workspace.GetWorkspaceResourcesOfType[*modconfig.Trigger](w)

	} else {
		var err error
		pipelines, triggers, err = load_mod.LoadPipelines(m.ctx, pipelineDir)
		if err != nil {
			return err
		}
	}

	m.triggers = triggers

	var rootModName string
	if modInfo != nil {
		rootModName = modInfo.ShortName
	} else {
		rootModName = "local"
	}

	cache.GetCache().SetWithTTL("#rootmod.name", rootModName, 24*7*52*99*time.Hour)
	err := m.cachePipelinesAndTriggers(pipelines, triggers)
	if err != nil {
		return err
	}

	logger.Info("Pipelines and triggers loaded", "pipelines", len(pipelines), "triggers", len(triggers), "rootMod", rootModName)

	m.RootMod = modInfo

	return nil
}

func (m *Manager) setupWatcher(w *workspace.Workspace) error {
	err := w.SetupWatcher(m.ctx, func(c context.Context, e error) {
		logger := fplog.Logger(m.ctx)
		logger.Error("error watching workspace", "error", e)
		m.apiService.ModMetadata.IsStale = true
	})
	if err != nil {
		return err
	}

	w.SetOnFileWatcherEventMessages(func() {
		logger := fplog.Logger(m.ctx)
		logger.Info("caching pipelines and triggers")
		m.triggers = w.Mod.ResourceMaps.Triggers
		m.cachePipelinesAndTriggers(w.Mod.ResourceMaps.Pipelines, w.Mod.ResourceMaps.Triggers)
		if err != nil {
			logger.Error("error caching pipelines and triggers", "error", err)
		} else {
			logger.Info("cached pipelines and triggers")
			m.apiService.ModMetadata.IsStale = false
			m.apiService.ModMetadata.LastLoaded = time.Now()
		}

		// Reload scheduled triggers
		logger.Info("rescheduling triggers")
		if m.schedulerService != nil {
			m.schedulerService.Triggers = w.Mod.ResourceMaps.Triggers
			err := m.schedulerService.RescheduleTriggers()
			if err != nil {
				logger.Error("error rescheduling triggers", "error", err)
			} else {
				logger.Info("rescheduled triggers")
			}
		}
	})
	return nil
}

func (m *Manager) startESService() error {
	// start event sourcing service
	esService, err := es.NewESService(m.ctx)
	if err != nil {
		return err
	}
	err = esService.Start()
	if err != nil {
		return err
	}
	esService.Status = "running"
	esService.StartedAt = utils.TimeNow()
	esService.RootMod = m.RootMod

	m.ESService = esService
	return nil
}

func (m *Manager) startAPIService() error {
	// Define the API service
	apiService, err := api.NewAPIService(m.ctx, m.ESService,
		api.WithHTTPAddress(m.HTTPAddress),
		api.WithHTTPPort(m.HTTPPort))

	if err != nil {
		return err
	}
	m.apiService = apiService

	// Start API
	return apiService.Start()
}

func (m *Manager) startSchedulerService() error {
	s := scheduler.NewSchedulerService(m.ctx, m.ESService, m.triggers)
	if !viper.GetBool(constants.ArgNoScheduler) {
		if err := s.Start(); err != nil {
			return err
		}
	}

	m.schedulerService = s
	return nil
}

// Stop stops services managed by the Manager.
func (m *Manager) Stop() error {
	fplog.Logger(m.ctx).Debug("manager stopping")
	defer fplog.Logger(m.ctx).Debug("manager stopped")

	// Ensure any log messages are synced before we exit
	logger := fplog.Logger(m.ctx)
	defer func() {
		// this is causing "inappropriate ioctl for device" error: https://github.com/uber-go/zap/issues/880
		//nolint:errcheck // we don't care if this fails
		logger.Sync()
	}()

	if m.apiService != nil {
		if err := m.apiService.Stop(); err != nil {
			// Log and continue stopping other services
			fplog.Logger(m.ctx).Error("error stopping api service", "error", err)
		}
	}

	if m.ESService != nil {
		if err := m.ESService.Stop(); err != nil {
			// Log and continue stopping other services
			fplog.Logger(m.ctx).Error("error stopping es service", "error", err)
		}
	}

	// Cleanup docker artifacts
	// TODO - Can we remove this since we cleanup per function etc?
	if docker.GlobalDockerClient != nil {
		if err := docker.GlobalDockerClient.CleanupArtifacts(); err != nil {
			fplog.Logger(m.ctx).Error("Failed to cleanup flowpipe docker artifacts", "error", err)
		}
	}

	m.StoppedAt = utils.TimeNow()

	return nil
}

func (m *Manager) InterruptHandler() {
	fmt.Println("IN")
	defer func() {
		fmt.Println("OUT")
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		fplog.Logger(m.ctx).Debug("Manager exiting", "signal", sig)
		err := m.Stop()
		if err != nil {
			panic(err)
		}

		done <- true
	}()
	<-done
	fplog.Logger(m.ctx).Debug("Manager exited")
}

func (m *Manager) cachePipelinesAndTriggers(pipelines map[string]*modconfig.Pipeline, triggers map[string]*modconfig.Trigger) error {
	inMemoryCache := cache.GetCache()
	var pipelineNames []string

	for _, p := range pipelines {
		pipelineNames = append(pipelineNames, p.Name())

		// TODO: how do we want to do this?
		inMemoryCache.SetWithTTL(p.Name(), p, 24*7*52*99*time.Hour)
	}

	inMemoryCache.SetWithTTL("#pipeline.names", pipelineNames, 24*7*52*99*time.Hour)

	var triggerNames []string
	for _, trigger := range triggers {
		triggerNames = append(triggerNames, trigger.Name())

		// if it's a webhook trigger, calculate the URL
		_, ok := trigger.Config.(*modconfig.TriggerHttp)
		if ok && !strings.HasPrefix(os.Getenv("RUN_MODE"), "TEST") {
			triggerUrl, err := calculateTriggerUrl(trigger)
			if err != nil {
				return err
			}
			trigger.Config.(*modconfig.TriggerHttp).Url = triggerUrl
		}

		inMemoryCache.SetWithTTL(trigger.Name(), trigger, 24*7*52*99*time.Hour)
	}
	inMemoryCache.SetWithTTL("#trigger.names", triggerNames, 24*7*52*99*time.Hour)

	return nil
}

func calculateTriggerUrl(trigger *modconfig.Trigger) (string, error) {
	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		return "", perr.InternalWithMessage("salt not found")
	}

	hashString := util.CalculateHash(trigger.FullName, salt.(string))

	return "/hook/" + trigger.FullName + "/" + hashString, nil
}
