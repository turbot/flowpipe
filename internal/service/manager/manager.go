package manager

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/scheduler"
	"github.com/turbot/flowpipe/pipeparser"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/filepaths"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/utils"
	"github.com/turbot/flowpipe/pipeparser/workspace"
)

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	apiService       *api.APIService
	esService        *es.ESService
	schedulerService *scheduler.SchedulerService

	triggers map[string]*modconfig.Trigger

	RaftNodeID    string `json:"raft_node_id,omitempty"`
	RaftBootstrap bool   `json:"raft_bootstrap"`
	RaftAddress   string `json:"raft_address,omitempty"`

	HTTPSAddress string `json:"https_address,omitempty"`

	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

// ManagerOption defines a type of function to configures the Manager.
type ManagerOption func(*Manager) error

// NewManager creates a new Manager.
func NewManager(ctx context.Context, opts ...ManagerOption) (*Manager, error) {
	// Defaults
	m := &Manager{
		ctx:    ctx,
		Status: "initialized",
	}
	// Set options
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return m, err
		}
	}
	return m, nil
}

func WithRaftNodeID(nodeID string) ManagerOption {
	return func(m *Manager) error {
		m.RaftNodeID = nodeID
		return nil
	}
}

func WithRaftBootstrap(bootstrap bool) ManagerOption {
	return func(m *Manager) error {
		m.RaftBootstrap = bootstrap
		return nil
	}
}

func WithRaftAddress(addr string) ManagerOption {
	return func(m *Manager) error {
		m.RaftAddress = addr
		return nil
	}
}

func WithHTTPAddress(addr string) ManagerOption {
	return func(m *Manager) error {
		m.HTTPSAddress = addr
		return nil
	}
}

// TODO: is there any point to have a separate "Initialize" and "Start"?
func (m *Manager) Initialize() error {
	pipelineDir := viper.GetString("pipeline.dir")

	filepaths.PipesComponentWorkspaceDataDir = ".flowpipe"
	filepaths.PipesComponentModsFileName = "mod.hcl"
	filepaths.PipesComponentDefaultVarsFileName = "flowpipe.pvars"

	constants.PipesComponentModDataExtension = ".hcl"
	constants.PipesComponentVariablesExtension = ".pvars"
	constants.PipesComponentAutoVariablesExtension = ".auto.pvars"
	constants.PipesComponentEnvInputVarPrefix = "P_VAR_"

	var pipelines map[string]*modconfig.Pipeline
	var triggers map[string]*modconfig.Trigger
	var modInfo *modconfig.Mod
	if pipeparser.ModFileExists(pipelineDir, filepaths.PipesComponentModsFileName) {

		w, errorAndWarning := workspace.LoadWithParams(m.ctx, pipelineDir, []string{".hcl", ".sp"})

		err := w.SetupWatcher(m.ctx, func(c context.Context, e error) {
			logger := fplog.Logger(m.ctx)
			logger.Error("error watching workspace", "error", e)
		})

		w.SetOnFileWatcherEventMessages(func() {
			logger := fplog.Logger(m.ctx)
			err := m.Reload(w.Mod.ResourceMaps.Pipelines, w.Mod.ResourceMaps.Triggers)
			if err != nil {
				logger.Error("error reloading pipelines", "error", err)
			}

			if m.schedulerService != nil {
				m.schedulerService.Triggers = w.Mod.ResourceMaps.Triggers
				err := m.schedulerService.ReloadTriggers()
				if err != nil {
					logger.Error("error reloading triggers", "error", err)
				}
			}

		})

		if err != nil {
			return err
		}

		if errorAndWarning.Error != nil {
			return errorAndWarning.Error
		}

		mod := w.Mod
		modInfo = mod

		pipelines = mod.ResourceMaps.Pipelines
		triggers = mod.ResourceMaps.Triggers

		for _, dependendMode := range mod.ResourceMaps.Mods {
			if dependendMode.Name() != mod.Name() {
				for _, pipeline := range dependendMode.ResourceMaps.Pipelines {
					pipelines[pipeline.Name()] = pipeline
				}
				for _, trigger := range dependendMode.ResourceMaps.Triggers {
					triggers[trigger.Name()] = trigger
				}
			}
		}
	} else {
		var err error
		pipelines, triggers, err = pipeparser.LoadPipelines(m.ctx, pipelineDir)
		if err != nil {
			return err
		}
	}

	m.triggers = triggers

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

		// TODO: how do we want to do this?
		inMemoryCache.SetWithTTL(trigger.Name(), trigger, 24*7*52*99*time.Hour)
	}
	inMemoryCache.SetWithTTL("#trigger.names", triggerNames, 24*7*52*99*time.Hour)

	inMemoryCache.SetWithTTL("#rootmod.name", modInfo.ShortName, 24*7*52*99*time.Hour)

	return nil
}

func (m *Manager) Reload(pipelines map[string]*modconfig.Pipeline, triggers map[string]*modconfig.Trigger) error {
	m.triggers = triggers

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

		// TODO: how do we want to do this?
		inMemoryCache.SetWithTTL(trigger.Name(), trigger, 24*7*52*99*time.Hour)
	}
	inMemoryCache.SetWithTTL("#trigger.names", triggerNames, 24*7*52*99*time.Hour)

	return nil
}

// Start starts services managed by the Manager.
func (m *Manager) Start() error {

	fplog.Logger(m.ctx).Debug("Manager starting")
	defer fplog.Logger(m.ctx).Debug("Manager started")

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

	m.esService = esService

	// Define the API service
	a, err := api.NewAPIService(m.ctx, esService,
		api.WithHTTPSAddress(m.HTTPSAddress))

	if err != nil {
		return err
	}
	m.apiService = a

	// Start API
	err = a.Start()
	if err != nil {
		return err
	}

	// Start the scheduler service
	s := scheduler.NewSchedulerService(m.ctx, esService, m.triggers)
	err = s.Start()
	if err != nil {
		return err
	}

	m.schedulerService = s

	m.StartedAt = utils.TimeNow()
	m.Status = "running"

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

	err := m.apiService.Stop()
	if err != nil {
		// Log and continue stopping other services
		fplog.Logger(m.ctx).Error("error stopping api service", "error", err)
	}

	err = m.esService.Stop()
	if err != nil {
		// Log and continue stopping other services
		fplog.Logger(m.ctx).Error("error stopping es service", "error", err)
	}

	// err = m.raftService.Stop()
	// if err != nil {
	// 	// Log and continue stopping other services
	// 	fplog.Logger(m.ctx).Error("error stopping raft service", "error", err)
	// }

	m.StoppedAt = utils.TimeNow()

	return nil
}

func (m *Manager) InterruptHandler() {
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
