package raft

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rqlite/rqlite-disco-clients/consul"
	"github.com/rqlite/rqlite-disco-clients/dns"
	"github.com/rqlite/rqlite-disco-clients/dnssrv"
	"github.com/rqlite/rqlite-disco-clients/etcd"
	"github.com/rqlite/rqlite/auth"
	"github.com/rqlite/rqlite/cluster"
	"github.com/rqlite/rqlite/disco"
	httpd "github.com/rqlite/rqlite/http"
	"github.com/rqlite/rqlite/rtls"
	"github.com/rqlite/rqlite/tcp"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/store"
)

const (
	DiscoModeNone     = ""
	DiscoModeConsulKV = "consul-kv"
	DiscoModeEtcdKV   = "etcd-kv"
	DiscoModeDNS      = "dns"
	DiscoModeDNSSRV   = "dns-srv"

	HTTPAddrFlag    = "http-addr"
	HTTPAdvAddrFlag = "http-adv-addr"
	RaftAddrFlag    = "raft-addr"
	RaftAdvAddrFlag = "raft-adv-addr"

	HTTPx509CertFlag = "http-cert"
	HTTPx509KeyFlag  = "http-key"
	NodeX509CertFlag = "node-cert"
	NodeX509KeyFlag  = "node-key"
)

// Start starts services managed by the Manager.
func (r *RaftService) Start() error {

	fplog.Logger(r.ctx).Debug("raft starting")
	defer fplog.Logger(r.ctx).Debug("raft started")

	cfg := config.Config(r.ctx)
	cfg.RaftNodeID = r.NodeID

	// Create internode network mux and configure.
	muxLn, err := net.Listen("tcp", cfg.RaftAddr)
	if err != nil {
		return fperr.WrapWithMessage(err, "failed to listen on %s", cfg.RaftAddr)
	}
	mux, err := startNodeMux(cfg, muxLn)
	if err != nil {
		//return fperr.WrapWithMessage(err, "failed to start node mux")
		return err
	}
	raftTn := mux.Listen(cluster.MuxRaftHeader)
	log.Printf("Raft TCP mux Listener registered with byte header %d", cluster.MuxRaftHeader)

	// Create the store.
	str, err := createStore(cfg, raftTn)
	if err != nil {
		//errors.New("failed to create store: %s", err.Error())
		return err
	}

	// Get any credential store.
	credStr, err := credentialStore(cfg)
	if err != nil {
		//errors.New("failed to get credential store: %s", err.Error())
		return err
	}

	// Create cluster service now, so nodes will be able to learn information about each other.
	clstrServ, err := clusterService(cfg, mux.Listen(cluster.MuxClusterHeader), str, str, credStr)
	if err != nil {
		//errors.New("failed to create cluster service: %s", err.Error())
		return err
	}
	log.Printf("cluster TCP mux Listener registered with byte header %d", cluster.MuxClusterHeader)

	// Create the HTTP service.
	//
	// We want to start the HTTP server as soon as possible, so the node is responsive and external
	// systems can see that it's running. We still have to open the Store though, so the node won't
	// be able to do much until that happens however.
	_, err = createClusterClient(cfg, clstrServ)
	//clstrClient, err := createClusterClient(cfg, clstrServ)
	if err != nil {
		//errors.New("failed to create cluster client: %s", err.Error())
		return err
	}
	/*
		httpServ, err := startHTTPService(cfg, str, clstrClient, credStr)
		if err != nil {
			//errors.New("failed to start HTTP server: %s", err.Error())
			return err
		}
		log.Printf("HTTP server started")
	*/
	var httpServ *httpd.Service

	fplog.Logger(r.ctx).Debug("store opening")

	// Now, open store. How long this takes does depend on how much data is being stored by rqlite.
	if err := str.Open(); err != nil {
		fplog.Logger(r.ctx).Error("store opening error", "error", err)
		//errors.New("failed to open store: %s", err.Error())
		return err
	}

	fplog.Logger(r.ctx).Debug("store opened")

	/*
		// Register remaining status providers.
		httpServ.RegisterStatus("cluster", clstrServ)
	*/

	// Create the cluster!
	nodes, err := str.Nodes()
	if err != nil {
		//errors.New("failed to get nodes %s", err.Error())
		return err
	}
	fplog.Logger(r.ctx).Debug("cluster creating")
	if err := createCluster(cfg, len(nodes) > 0, str, httpServ, credStr); err != nil {
		//errors.New("clustering failure: %s", err.Error())
		return err
	}
	fplog.Logger(r.ctx).Debug("cluster created")

	return nil

}

// startNodeMux starts the TCP mux on the given listener, which should be already
// bound to the relevant interface.
func startNodeMux(cfg *config.Configuration, ln net.Listener) (*tcp.Mux, error) {
	var err error
	adv := tcp.NameAddress{
		Address: cfg.RaftAdv,
	}

	var mux *tcp.Mux
	if cfg.NodeX509Cert != "" {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("enabling node-to-node encryption with cert: %s, key: %s",
			cfg.NodeX509Cert, cfg.NodeX509Key))
		if cfg.NodeX509CACert != "" {
			b.WriteString(fmt.Sprintf(", CA cert %s", cfg.NodeX509CACert))
		}
		if cfg.NodeVerifyClient {
			b.WriteString(", mutual TLS disabled")
		} else {
			b.WriteString(", mutual TLS enabled")
		}
		log.Println(b.String())
		mux, err = tcp.NewTLSMux(ln, adv, cfg.NodeX509Cert, cfg.NodeX509Key, cfg.NodeX509CACert,
			cfg.NoNodeVerify, cfg.NodeVerifyClient)
	} else {
		mux, err = tcp.NewMux(ln, adv)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create node-to-node mux: %s", err.Error())
	}
	go mux.Serve()

	return mux, nil
}

func createStore(cfg *config.Configuration, ln *tcp.Layer) (*store.Store, error) {
	dataPath, err := filepath.Abs(cfg.DataPath)
	if err != nil {
		return nil, fperr.WrapWithMessage(err, "failed to determine absolute data path")
	}
	dbConf := store.NewDBConfig(!cfg.OnDisk)
	dbConf.OnDiskPath = cfg.OnDiskPath
	dbConf.FKConstraints = cfg.FKConstraints

	str := store.New(ln, &store.Config{
		DBConf: dbConf,
		Dir:    cfg.DataPath,
		ID:     cfg.RaftNodeID,
	})

	// Set optional parameters on store.
	str.StartupOnDisk = cfg.OnDiskStartup
	str.SetRequestCompression(cfg.CompressionBatch, cfg.CompressionSize)
	str.RaftLogLevel = cfg.RaftLogLevel
	str.NoFreeListSync = cfg.RaftNoFreelistSync
	str.ShutdownOnRemove = cfg.RaftShutdownOnRemove
	str.SnapshotThreshold = cfg.RaftSnapThreshold
	str.SnapshotInterval = cfg.RaftSnapInterval
	str.LeaderLeaseTimeout = cfg.RaftLeaderLeaseTimeout
	str.HeartbeatTimeout = cfg.RaftHeartbeatTimeout
	str.ElectionTimeout = cfg.RaftElectionTimeout
	str.ApplyTimeout = cfg.RaftApplyTimeout
	str.BootstrapExpect = cfg.BootstrapExpect
	str.ReapTimeout = cfg.RaftReapNodeTimeout
	str.ReapReadOnlyTimeout = cfg.RaftReapReadOnlyNodeTimeout

	isNew := store.IsNewNode(dataPath)
	if isNew {
		log.Printf("no preexisting node state detected in %s, node may be bootstrapping", dataPath)
	} else {
		log.Printf("preexisting node state detected in %s", dataPath)
	}

	return str, nil
}

func credentialStore(cfg *config.Configuration) (*auth.CredentialsStore, error) {
	if cfg.AuthFile == "" {
		return nil, nil
	}

	f, err := os.Open(cfg.AuthFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open authentication file %s: %s", cfg.AuthFile, err.Error())
	}

	cs := auth.NewCredentialsStore()
	if cs.Load(f); err != nil {
		return nil, err
	}
	return cs, nil
}

func createDiscoService(cfg *config.Configuration, str *store.Store) (*disco.Service, error) {
	var c disco.Client
	var err error

	rc := cfg.DiscoConfigReader()
	defer func() {
		if rc != nil {
			rc.Close()
		}
	}()
	if cfg.DiscoMode == DiscoModeConsulKV {
		var consulCfg *consul.Config
		consulCfg, err = consul.NewConfigFromReader(rc)
		if err != nil {
			return nil, fmt.Errorf("create Consul config: %s", err.Error())
		}

		c, err = consul.New(cfg.DiscoKey, consulCfg)
		if err != nil {
			return nil, fmt.Errorf("create Consul client: %s", err.Error())
		}
	} else if cfg.DiscoMode == DiscoModeEtcdKV {
		var etcdCfg *etcd.Config
		etcdCfg, err = etcd.NewConfigFromReader(rc)
		if err != nil {
			return nil, fmt.Errorf("create etcd config: %s", err.Error())
		}

		c, err = etcd.New(cfg.DiscoKey, etcdCfg)
		if err != nil {
			return nil, fmt.Errorf("create etcd client: %s", err.Error())
		}
	} else {
		return nil, fmt.Errorf("invalid disco service: %s", cfg.DiscoMode)
	}

	return disco.NewService(c, str), nil
}

/*
func startHTTPService(cfg *config.Config, str *store.Store, cltr *cluster.Client, credStr *auth.CredentialsStore) (*httpd.Service, error) {
	// Create HTTP server and load authentication information.
	s := httpd.New(cfg.HTTPAddr, str, cltr, credStr)

	s.CACertFile = cfg.HTTPx509CACert
	s.CertFile = cfg.HTTPx509Cert
	s.KeyFile = cfg.HTTPx509Key
	s.TLS1011 = cfg.TLS1011
	s.ClientVerify = cfg.HTTPVerifyClient
	s.Expvar = cfg.Expvar
	s.Pprof = cfg.PprofEnabled
	s.DefaultQueueCap = cfg.WriteQueueCap
	s.DefaultQueueBatchSz = cfg.WriteQueueBatchSz
	s.DefaultQueueTimeout = cfg.WriteQueueTimeout
	s.DefaultQueueTx = cfg.WriteQueueTx
	s.BuildInfo = map[string]interface{}{
		"commit":     cmd.Commit,
		"branch":     cmd.Branch,
		"version":    cmd.Version,
		"compiler":   runtime.Compiler,
		"build_time": cmd.Buildtime,
	}
	return s, s.Start()
}
*/

func clusterService(cfg *config.Configuration, tn cluster.Transport, db cluster.Database, mgr cluster.Manager, credStr *auth.CredentialsStore) (*cluster.Service, error) {
	c := cluster.New(tn, db, mgr, credStr)
	c.SetAPIAddr(cfg.HTTPAdv)
	c.EnableHTTPS(cfg.HTTPx509Cert != "" && cfg.HTTPx509Key != "") // Conditions met for an HTTPS API
	if err := c.Open(); err != nil {
		return nil, err
	}
	return c, nil
}

func createClusterClient(cfg *config.Configuration, clstr *cluster.Service) (*cluster.Client, error) {
	var dialerTLSConfig *tls.Config
	var err error
	if cfg.NodeX509Cert != "" || cfg.NodeX509CACert != "" {
		dialerTLSConfig, err = rtls.CreateClientConfig(cfg.NodeX509Cert, cfg.NodeX509Key,
			cfg.NodeX509CACert, cfg.NoNodeVerify, cfg.TLS1011)
		if err != nil {
			log.Fatalf("failed to create TLS config for cluster dialer: %s", err.Error())
		}
	}
	clstrDialer := tcp.NewDialer(cluster.MuxClusterHeader, dialerTLSConfig)
	clstrClient := cluster.NewClient(clstrDialer, cfg.ClusterConnectTimeout)
	if err := clstrClient.SetLocal(cfg.RaftAdv, clstr); err != nil {
		log.Fatalf("failed to set cluster client local parameters: %s", err.Error())
	}
	return clstrClient, nil
}

func createCluster(cfg *config.Configuration, hasPeers bool, str *store.Store,
	httpServ *httpd.Service, credStr *auth.CredentialsStore) error {
	var tlsConfig *tls.Config
	var err error
	if cfg.HTTPx509Cert != "" || cfg.HTTPx509CACert != "" {
		tlsConfig, err = rtls.CreateClientConfig(cfg.HTTPx509Cert, cfg.HTTPx509Key, cfg.HTTPx509CACert,
			cfg.NoHTTPVerify, cfg.TLS1011)
		if err != nil {
			return fmt.Errorf("failed to create TLS client config for cluster: %s", err.Error())
		}
	}

	joins := cfg.JoinAddresses()
	if joins == nil && cfg.DiscoMode == "" && !hasPeers {
		if cfg.RaftNonVoter {
			return fmt.Errorf("cannot create a new non-voting node without joining it to an existing cluster")
		}

		// Brand new node, told to bootstrap itself. So do it.
		log.Println("bootstraping single new node")
		if err := str.Bootstrap(store.NewServer(str.ID(), cfg.RaftAdv, true)); err != nil {
			return fmt.Errorf("failed to bootstrap single new node: %s", err.Error())
		}
		return nil
	}

	// Prepare the Joiner
	joiner := cluster.NewJoiner(cfg.JoinSrcIP, cfg.JoinAttempts, cfg.JoinInterval, tlsConfig)
	if cfg.JoinAs != "" {
		pw, ok := credStr.Password(cfg.JoinAs)
		if !ok {
			return fmt.Errorf("user %s does not exist in credential store", cfg.JoinAs)
		}
		joiner.SetBasicAuth(cfg.JoinAs, pw)
	}

	// Prepare definition of being part of a cluster.
	isClustered := func() bool {
		leader, _ := str.LeaderAddr()
		return leader != ""
	}

	if joins != nil && cfg.BootstrapExpect == 0 {
		// Explicit join operation requested, so do it.
		j, err := joiner.Do(joins, str.ID(), cfg.RaftAdv, !cfg.RaftNonVoter)
		if err != nil {
			return fmt.Errorf("failed to join cluster: %s", err.Error())
		}
		log.Println("successfully joined cluster at", j)
		return nil
	}

	if joins != nil && cfg.BootstrapExpect > 0 {
		if hasPeers {
			log.Println("preexisting node configuration detected, ignoring bootstrap request")
			return nil
		}

		// Bootstrap with explicit join addresses requests.
		bs := cluster.NewBootstrapper(cluster.NewAddressProviderString(joins), tlsConfig)
		if cfg.JoinAs != "" {
			pw, ok := credStr.Password(cfg.JoinAs)
			if !ok {
				return fmt.Errorf("user %s does not exist in credential store", cfg.JoinAs)
			}
			bs.SetBasicAuth(cfg.JoinAs, pw)
		}
		return bs.Boot(str.ID(), cfg.RaftAdv, isClustered, cfg.BootstrapExpectTimeout)
	}

	if cfg.DiscoMode == "" {
		// No more clustering techniques to try. Node will just sit, probably using
		// existing Raft state.
		return nil
	}

	log.Printf("discovery mode: %s", cfg.DiscoMode)
	switch cfg.DiscoMode {
	case DiscoModeDNS, DiscoModeDNSSRV:
		if hasPeers {
			log.Printf("preexisting node configuration detected, ignoring %s option", cfg.DiscoMode)
			return nil
		}
		rc := cfg.DiscoConfigReader()
		defer func() {
			if rc != nil {
				rc.Close()
			}
		}()

		var provider interface {
			cluster.AddressProvider
			httpd.StatusReporter
		}
		if cfg.DiscoMode == DiscoModeDNS {
			dnsCfg, err := dns.NewConfigFromReader(rc)
			if err != nil {
				return fmt.Errorf("error reading DNS configuration: %s", err.Error())
			}
			provider = dns.New(dnsCfg)

		} else {
			dnssrvCfg, err := dnssrv.NewConfigFromReader(rc)
			if err != nil {
				return fmt.Errorf("error reading DNS configuration: %s", err.Error())
			}
			provider = dnssrv.New(dnssrvCfg)
		}

		bs := cluster.NewBootstrapper(provider, tlsConfig)
		if cfg.JoinAs != "" {
			pw, ok := credStr.Password(cfg.JoinAs)
			if !ok {
				return fmt.Errorf("user %s does not exist in credential store", cfg.JoinAs)
			}
			bs.SetBasicAuth(cfg.JoinAs, pw)
		}
		httpServ.RegisterStatus("disco", provider)
		return bs.Boot(str.ID(), cfg.RaftAdv, isClustered, cfg.BootstrapExpectTimeout)

	case DiscoModeEtcdKV, DiscoModeConsulKV:
		discoService, err := createDiscoService(cfg, str)
		if err != nil {
			return fmt.Errorf("failed to start discovery service: %s", err.Error())
		}

		if !hasPeers {
			log.Println("no preexisting nodes, registering with discovery service")

			leader, addr, err := discoService.Register(str.ID(), cfg.HTTPURL(), cfg.RaftAdv)
			if err != nil {
				return fmt.Errorf("failed to register with discovery service: %s", err.Error())
			}
			if leader {
				log.Println("node registered as leader using discovery service")
				if err := str.Bootstrap(store.NewServer(str.ID(), str.Addr(), true)); err != nil {
					return fmt.Errorf("failed to bootstrap single new node: %s", err.Error())
				}
			} else {
				for {
					log.Printf("discovery service returned %s as join address", addr)
					if j, err := joiner.Do([]string{addr}, str.ID(), cfg.RaftAdv, !cfg.RaftNonVoter); err != nil {
						log.Printf("failed to join cluster at %s: %s", addr, err.Error())

						time.Sleep(time.Second)
						_, addr, err = discoService.Register(str.ID(), cfg.HTTPURL(), cfg.RaftAdv)
						if err != nil {
							log.Printf("failed to get updated leader: %s", err.Error())
						}
						continue
					} else {
						log.Println("successfully joined cluster at", j)
						break
					}
				}
			}
		} else {
			log.Println("preexisting node configuration detected, not registering with discovery service")
		}
		go discoService.StartReporting(cfg.NodeID, cfg.HTTPURL(), cfg.RaftAdv)
		httpServ.RegisterStatus("disco", discoService)

	default:
		return fmt.Errorf("invalid disco mode %s", cfg.DiscoMode)
	}
	return nil
}
