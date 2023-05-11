package manager

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/api"
	"github.com/turbot/flowpipe/service/raft"
	"github.com/turbot/flowpipe/util"
)

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	apiService  *api.APIService
	raftService *raft.RaftService

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

// Start starts services managed by the Manager.
func (m *Manager) Start() error {

	fplog.Logger(m.ctx).Debug("manager starting")
	defer fplog.Logger(m.ctx).Debug("manager started")

	// Define the Raft service
	r, err := raft.NewRaftService(m.ctx,
		raft.WithNodeID(m.RaftNodeID),
		raft.WithBootstrap(m.RaftBootstrap),
		raft.WithAddress(m.RaftAddress))
	if err != nil {
		panic(err)
	}
	m.raftService = r

	// Define the API service
	a, err := api.NewAPIService(m.ctx,
		api.WithHTTPSAddress(m.HTTPSAddress),
		api.WithRaftService(m.raftService))
	if err != nil {
		return err
	}
	m.apiService = a

	// Start API
	err = a.Start()
	if err != nil {
		return err
	}

	// Start raft
	err = r.Start()
	if err != nil {
		return err
	}

	m.StartedAt = util.TimeNowPtr()
	m.Status = "running"
	return nil
}

// Stop stops services managed by the Manager.
func (m *Manager) Stop() error {

	fplog.Logger(m.ctx).Debug("manager stopping")
	defer fplog.Logger(m.ctx).Debug("manager stopped")

	// Ensure any log messages are synced before we exit
	logger := fplog.Logger(m.ctx)
	defer logger.Sync()

	err := m.apiService.Stop()
	if err != nil {
		// Log and continue stopping other services
		fplog.Logger(m.ctx).Error("error stopping api service", "error", err)
	}

	err = m.raftService.Stop()
	if err != nil {
		// Log and continue stopping other services
		fplog.Logger(m.ctx).Error("error stopping raft service", "error", err)
	}

	m.StoppedAt = util.TimeNowPtr()

	return nil
}

func (m *Manager) InterruptHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		fplog.Logger(m.ctx).Debug("manager exiting", "signal", sig)
		m.Stop()
		done <- true
	}()
	<-done
	fplog.Logger(m.ctx).Debug("manager exited")
}
