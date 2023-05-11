package raft

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	_ "github.com/swaggo/swag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/fsm"
	"github.com/turbot/flowpipe/util"
)

// RaftService represents the Raft service.
type RaftService struct {
	// Ctx is the context used by the Raft service.
	ctx context.Context

	grpcServer *grpc.Server
	Raft       *raft.Raft
	Storage    *fsm.KeyValue

	Bootstrap bool   `json:"bootstrap"`
	NodeID    string `json:"node_id"`
	Host      string `json:"host"`
	Port      string `json:"port"`

	// Status tracking for the Raft service.
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

// RaftServiceOption defines a type of function to configures the RaftService.
type RaftServiceOption func(*RaftService) error

// NewRaftService creates a new RaftService.
func NewRaftService(ctx context.Context, opts ...RaftServiceOption) (*RaftService, error) {
	// Defaults
	r := &RaftService{
		ctx:    ctx,
		Status: "initialized",
		Host:   "localhost",
		Port:   "7104",
	}
	// Set options
	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return r, err
		}
	}
	// If a node ID has not been passed, then generate a unique consistent one
	// for the machine and port we're running on.
	// TODO - Is this good? Do we want the same ID each time? What if they run
	// with different mods? Does the mod path need to be included in the stable
	// ID generation?
	if r.NodeID == "" {
		nid, err := util.NodeID(r.Port)
		if err != nil {
			return nil, err
		}
		r.NodeID = nid
	}
	return r, nil
}

func WithNodeID(nodeID string) RaftServiceOption {
	return func(r *RaftService) error {
		if nodeID != "" {
			r.NodeID = nodeID
		}
		return nil
	}
}

func WithBootstrap(bootstrap bool) RaftServiceOption {
	return func(r *RaftService) error {
		r.Bootstrap = bootstrap
		return nil
	}
}

// WithAddress sets the host and port of the raft service from the given address string
// in host:port format. If the port is not specified, the default port 50051 is used.
func WithAddress(addr string) RaftServiceOption {
	return func(r *RaftService) error {
		if addr == "" {
			return nil
		}
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return err
		}
		if host != "" {
			r.Host = host
		}
		if port != "" {
			r.Port = port
		}
		return nil
	}
}

func (r *RaftService) StorageDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".flowpipe", "raft", r.NodeID), nil
}

func (r *RaftService) ServerAddress() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// Start starts services managed by the Manager.
func (r *RaftService) StartExample() error {

	fplog.Logger(r.ctx).Debug("raft starting")
	defer fplog.Logger(r.ctx).Debug("raft started")

	if r.NodeID == "" {
		return errors.New("node ID is required")
	}

	sock, err := net.Listen("tcp", fmt.Sprintf(":%s", r.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	r.Storage = fsm.NewKeyValue()

	raftService, tm, err := r.Initialize(r.Storage)
	if err != nil {
		return fmt.Errorf("failed to start raft: %v", err)
	}
	r.Raft = raftService

	r.grpcServer = grpc.NewServer()
	tm.Register(r.grpcServer)
	leaderhealth.Setup(raftService, r.grpcServer, []string{"Example"})
	raftadmin.Register(r.grpcServer, raftService)
	reflection.Register(r.grpcServer)

	go func() {
		if err := r.grpcServer.Serve(sock); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	r.StartedAt = util.TimeNowPtr()
	r.Status = "running"
	return nil
}

// Stop stops services managed by the Manager.
func (r *RaftService) Stop() error {

	fplog.Logger(r.ctx).Debug("raft stopping")
	defer fplog.Logger(r.ctx).Debug("raft stopped")

	if r.grpcServer != nil {
		r.grpcServer.GracefulStop()
	}
	r.StoppedAt = util.TimeNowPtr()
	r.Status = "stopped"
	return nil
}

func (r *RaftService) Initialize(fsm raft.FSM) (*raft.Raft, *transport.Manager, error) {

	c := raft.DefaultConfig()
	c.LocalID = raft.ServerID(r.NodeID)

	baseDir, err := r.StorageDir()
	if err != nil {
		return nil, nil, err
	}
	err = os.MkdirAll(baseDir, os.ModePerm)
	if err != nil {
		// TODO - wrap error
		return nil, nil, err
	}

	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err)
	}

	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.dat"), err)
	}

	fss, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	// Create new insecure credentials
	creds := insecure.NewCredentials()

	// Create gRPC dial options with secure credentials
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	// Create transport manager with gRPC dial options
	tm := transport.New(raft.ServerAddress(r.ServerAddress()), dialOpts)

	raftService, err := raft.NewRaft(c, fsm, ldb, sdb, fss, tm.Transport())
	if err != nil {
		return nil, nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	if r.Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(r.NodeID),
					Address:  raft.ServerAddress(r.ServerAddress()),
				},
			},
		}
		f := raftService.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			return nil, nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}

	r.Raft = raftService

	return raftService, tm, nil
}
