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
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	apiService *api.APIService
	esService  *es.ESService

	triggers map[string]pipeline.ITrigger

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

	fpParseContext, err := pipeline.LoadFlowpipeConfig(m.ctx, pipelineDir)
	if err != nil {
		return err
	}

	pipelines := fpParseContext.PipelineHcls

	m.triggers = fpParseContext.TriggerHcls

	inMemoryCache := cache.GetCache()
	var pipelineNames []string

	for pipelineName := range pipelines {
		pipelineNames = append(pipelineNames, pipelineName)

		// TODO: how do we want to do this?
		inMemoryCache.SetWithTTL(pipelineName, pipelines[pipelineName], 24*7*52*99*time.Hour)
	}

	inMemoryCache.SetWithTTL("#pipeline.names", pipelineNames, 24*7*52*99*time.Hour)
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
	esService.StartedAt = util.TimeNow()

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

	m.StartedAt = util.TimeNow()
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

	m.StoppedAt = util.TimeNow()

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
