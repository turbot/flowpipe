package manager

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/scheduler"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
)

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	RootMod *modconfig.Mod

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
			trigger.CachePipelinesAndTriggers(w.Mod.ResourceMaps.Pipelines, w.Mod.ResourceMaps.Triggers)
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
	err := trigger.CachePipelinesAndTriggers(pipelines, triggers)
	if err != nil {
		return err
	}

	logger.Info("Pipelines and triggers loaded", "pipelines", len(pipelines), "triggers", len(triggers), "rootMod", rootModName)

	m.RootMod = modInfo

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
	esService.RootMod = m.RootMod

	m.esService = esService

	// Define the API service
	apiService, err := api.NewAPIService(m.ctx, esService,
		api.WithHTTPSAddress(m.HTTPSAddress))

	if err != nil {
		return err
	}
	m.apiService = apiService

	// Start API
	err = apiService.Start()
	if err != nil {
		return err
	}

	// Start the scheduler service
	s := scheduler.NewSchedulerService(m.ctx, esService, m.triggers)
	if !viper.GetBool(constants.ArgNoScheduler) {
		err = s.Start()
		if err != nil {
			return err
		}
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

	// Cleanup docker artifacts
	// TODO - Can we remove this since we cleanup per function etc?
	if docker.GlobalDockerClient != nil {
		err = docker.GlobalDockerClient.CleanupArtifacts()
		if err != nil {
			fplog.Logger(m.ctx).Error("Failed to cleanup flowpipe docker artifacts", "error", err)
		}
	}

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
