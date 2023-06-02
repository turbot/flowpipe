package config

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
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

// Configuration represents the configuration as set by command-line flags.
// All variables will be set, unless explicitly noted.
type Configuration struct {
	ctx   context.Context
	Viper *viper.Viper

	// TODO: Directory where log files will be written.
	LogDir string `json:"log_dir,omitempty"`

	// Workspace profile to use.
	Workspace string

	// DataPath is path to node data. Always set.
	DataPath string

	// ConfigPath is an optional specific location for the config file.
	ConfigPath string

	RaftNodeID    string `json:"raft_node_id,omitempty"`
	RaftBootstrap bool   `json:"raft_bootstrap,omitempty"`
	RaftHost      string `json:"raft_host,omitempty"`
	RaftPort      string `json:"raft_port,omitempty"`

	// HTTPAddr is the bind network address for the HTTP Server.
	// It never includes a trailing HTTP or HTTPS.
	HTTPAddr string

	// HTTPAdv is the advertised HTTP server network.
	HTTPAdv string

	// TLS1011 indicates whether the node should support deprecated
	// encryption standards.
	TLS1011 bool

	// AuthFile is the path to the authentication file. May not be set.
	AuthFile string

	// HTTPx509CACert is the path to the CA certficate file for when this node verifies
	// other certificates for any HTTP communications. May not be set.
	HTTPx509CACert string

	// HTTPx509Cert is the path to the X509 cert for the HTTP server. May not be set.
	HTTPx509Cert string

	// HTTPx509Key is the path to the private key for the HTTP server. May not be set.
	HTTPx509Key string

	// NoHTTPVerify disables checking other nodes' server HTTP X509 certs for validity.
	NoHTTPVerify bool

	// HTTPVerifyClient indicates whether the HTTP server should verify client certificates.
	HTTPVerifyClient bool

	// NodeEncrypt indicates whether node encryption should be enabled.
	NodeEncrypt bool

	// NodeX509CACert is the path to the CA certficate file for when this node verifies
	// other certificates for any inter-node communications. May not be set.
	NodeX509CACert string

	// NodeX509Cert is the path to the X509 cert for the Raft server. May not be set.
	NodeX509Cert string

	// NodeX509Key is the path to the X509 key for the Raft server. May not be set.
	NodeX509Key string

	// NoNodeVerify disables checking other nodes' Node X509 certs for validity.
	NoNodeVerify bool

	// NodeVerifyClient indicates whether a node should verify client certificates from
	// other nodes.
	NodeVerifyClient bool

	// NodeID is the Raft ID for the node.
	NodeID string

	// RaftAddr is the bind network address for the Raft server.
	RaftAddr string

	// RaftAdv is the advertised Raft server address.
	RaftAdv string

	// JoinSrcIP sets the source IP address during Join request. May not be set.
	JoinSrcIP string

	// JoinAddr is the list addresses to use for a join attempt. Each address
	// will include the proto (HTTP or HTTPS) and will never include the node's
	// own HTTP server address. May not be set.
	JoinAddr string

	// JoinAs sets the user join attempts should be performed as. May not be set.
	JoinAs string

	// JoinAttempts is the number of times a node should attempt to join using a
	// given address.
	JoinAttempts int

	// JoinInterval is the time between retrying failed join operations.
	JoinInterval time.Duration

	// BootstrapExpect is the minimum number of nodes required for a bootstrap.
	BootstrapExpect int

	// BootstrapExpectTimeout is the maximum time a bootstrap operation can take.
	BootstrapExpectTimeout time.Duration

	// DisoMode sets the discovery mode. May not be set.
	DiscoMode string

	// DiscoKey sets the discovery prefix key.
	DiscoKey string

	// DiscoConfig sets the path to any discovery configuration file. May not be set.
	DiscoConfig string

	// Expvar enables go/expvar information. Defaults to true.
	Expvar bool

	// PprofEnabled enables Go PProf information. Defaults to true.
	PprofEnabled bool

	// OnDisk enables on-disk mode.
	OnDisk bool

	// OnDiskPath sets the path to the SQLite file. May not be set.
	OnDiskPath string

	// OnDiskStartup disables the in-memory on-disk startup optimization.
	OnDiskStartup bool

	// FKConstraints enables SQLite foreign key constraints.
	FKConstraints bool

	// RaftLogLevel sets the minimum logging level for the Raft subsystem.
	RaftLogLevel string

	// RaftNonVoter controls whether this node is a voting, read-only node.
	RaftNonVoter bool

	// RaftSnapThreshold is the number of outstanding log entries that trigger snapshot.
	RaftSnapThreshold uint64

	// RaftSnapInterval sets the threshold check interval.
	RaftSnapInterval time.Duration

	// RaftLeaderLeaseTimeout sets the leader lease timeout.
	RaftLeaderLeaseTimeout time.Duration

	// RaftHeartbeatTimeout sets the heartbeat timeout.
	RaftHeartbeatTimeout time.Duration

	// RaftElectionTimeout sets the election timeout.
	RaftElectionTimeout time.Duration

	// RaftApplyTimeout sets the Log-apply timeout.
	RaftApplyTimeout time.Duration

	// RaftShutdownOnRemove sets whether Raft should be shutdown if the node is removed
	RaftShutdownOnRemove bool

	// RaftStepdownOnShutdown sets whether Leadership should be relinquished on shutdown
	RaftStepdownOnShutdown bool

	// RaftNoFreelistSync disables syncing Raft database freelist to disk. When true,
	// it improves the database write performance under normal operation, but requires
	// a full database re-sync during recovery.
	RaftNoFreelistSync bool

	// RaftReapNodeTimeout sets the duration after which a non-reachable voting node is
	// reaped i.e. removed from the cluster.
	RaftReapNodeTimeout time.Duration

	// RaftReapReadOnlyNodeTimeout sets the duration after which a non-reachable non-voting node is
	// reaped i.e. removed from the cluster.
	RaftReapReadOnlyNodeTimeout time.Duration

	// ClusterConnectTimeout sets the timeout when initially connecting to another node in
	// the cluster, for non-Raft communications.
	ClusterConnectTimeout time.Duration

	// WriteQueueCap is the default capacity of Execute queues
	WriteQueueCap int

	// WriteQueueBatchSz is the default batch size for Execute queues
	WriteQueueBatchSz int

	// WriteQueueTimeout is the default time after which any data will be sent on
	// Execute queues, if a batch size has not been reached.
	WriteQueueTimeout time.Duration

	// WriteQueueTx controls whether writes from the queue are done within a transaction.
	WriteQueueTx bool

	// CompressionSize sets request query size for compression attempt
	CompressionSize int

	// CompressionBatch sets request batch threshold for compression attempt.
	CompressionBatch int

	// CPUProfile enables CPU profiling.
	CPUProfile string

	// MemProfile enables memory profiling.
	MemProfile string
}

// ConfigOption defines a type of function to configures the Config.
type ConfigOption func(*Configuration) error

// NewConfig creates a new Config.
func NewConfig(ctx context.Context, opts ...ConfigOption) (*Configuration, error) {
	// Defaults
	c := &Configuration{
		ctx:          ctx,
		Viper:        viper.New(),
		JoinAttempts: 5,
		JoinInterval: 3 * time.Second,
	}
	// Set options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}

func WithRaftNodeID(nodeID string) ConfigOption {
	return func(c *Configuration) error {
		if nodeID != "" {
			c.RaftNodeID = nodeID
		}
		return nil
	}
}

func WithConfigPath(configPath string) ConfigOption {
	return func(c *Configuration) error {
		if configPath != "" {
			c.ConfigPath = configPath
		}
		return nil
	}
}

func (c *Configuration) InitializeViper() error {

	// Convenience
	v := c.Viper

	if c.ConfigPath != "" {
		// User has provided a specific config file location, so use that.
		// We do not look in other (default) locations in this case.
		v.SetConfigFile(c.ConfigPath)
	} else {
		// Look for a config file in standard locations.
		// First, the current working directory.
		v.AddConfigPath(".")
		// Second, the user's home directory.
		v.AddConfigPath("$HOME/.flowpipe")

		// Set the base name of the config file, without the file extension.
		// This means they can use a variety of formats, like HCL or YAML or JSON.
		v.SetConfigName("flowpipe")
	}

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fperr.WrapWithMessage(err, "error reading config file")
		}
	}
	fplog.Logger(c.ctx).Debug("Using config file:", v.ConfigFileUsed())

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable FLOWPIPE_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix("FLOWPIPE")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to FLOWPIPE_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	return nil
}

// WithFlags parses the command line, and populates the configuration.
func WithFlags() ConfigOption {
	return func(c *Configuration) error {
		if flag.Parsed() {
			return fmt.Errorf("command-line flags already parsed")
		}

		flag.StringVar(&c.ConfigPath, "config-path", "~/.flowpipe/flowpipe.yaml", "Location of config file")

		flag.StringVar(&c.DataPath, "data-path", "~/.flowpipe/raft", "Path to data directory")

		flag.StringVar(&c.RaftNodeID, "raft-node-id", "bar", "Unique name for node. If not set, set to advertised Raft address")
		flag.BoolVar(&c.RaftBootstrap, "raft-bootstrap", false, "If true, bootstrap the cluster")

		flag.StringVar(&c.HTTPAddr, HTTPAddrFlag, "localhost:4001", "HTTP server bind address. To enable HTTPS, set X.509 certificate and key")
		flag.StringVar(&c.HTTPAdv, HTTPAdvAddrFlag, "", "Advertised HTTP address. If not set, same as HTTP server bind")
		flag.BoolVar(&c.TLS1011, "tls1011", false, "Support deprecated TLS versions 1.0 and 1.1")
		flag.StringVar(&c.HTTPx509CACert, "http-ca-cert", "", "Path to X.509 CA certificate for HTTPS")
		flag.StringVar(&c.HTTPx509Cert, HTTPx509CertFlag, "", "Path to HTTPS X.509 certificate")
		flag.StringVar(&c.HTTPx509Key, HTTPx509KeyFlag, "", "Path to HTTPS X.509 private key")
		flag.BoolVar(&c.NoHTTPVerify, "http-no-verify", false, "Skip verification of remote node's HTTPS certificate when joining a cluster")
		flag.BoolVar(&c.HTTPVerifyClient, "http-verify-client", false, "Enable mutual TLS for HTTPS")
		flag.BoolVar(&c.NodeEncrypt, "node-encrypt", false, "Ignored, control node-to-node encryption by setting node certificate and key")
		flag.StringVar(&c.NodeX509CACert, "node-ca-cert", "", "Path to X.509 CA certificate for node-to-node encryption")
		flag.StringVar(&c.NodeX509Cert, NodeX509CertFlag, "", "Path to X.509 certificate for node-to-node mutual authentication and encryption")
		flag.StringVar(&c.NodeX509Key, NodeX509KeyFlag, "", "Path to X.509 private key for node-to-node mutual authentication and encryption")
		flag.BoolVar(&c.NoNodeVerify, "node-no-verify", false, "Skip verification of any node-node certificate")
		flag.BoolVar(&c.NodeVerifyClient, "node-verify-client", false, "Enable mutual TLS for node-to-node communication")
		flag.StringVar(&c.AuthFile, "auth", "", "Path to authentication and authorization file. If not set, not enabled")
		flag.StringVar(&c.RaftAddr, RaftAddrFlag, "localhost:4002", "Raft communication bind address")
		flag.StringVar(&c.RaftAdv, RaftAdvAddrFlag, "", "Advertised Raft communication address. If not set, same as Raft bind")
		flag.StringVar(&c.JoinSrcIP, "join-source-ip", "", "Set source IP address during Join request")
		flag.StringVar(&c.JoinAddr, "join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
		flag.StringVar(&c.JoinAs, "join-as", "", "Username in authentication file to join as. If not set, joins anonymously")
		flag.IntVar(&c.JoinAttempts, "join-attempts", 5, "Number of join attempts to make")
		flag.DurationVar(&c.JoinInterval, "join-interval", 3*time.Second, "Period between join attempts")
		flag.IntVar(&c.BootstrapExpect, "bootstrap-expect", 0, "Minimum number of nodes required for a bootstrap")
		flag.DurationVar(&c.BootstrapExpectTimeout, "bootstrap-expect-timeout", 120*time.Second, "Maximum time for bootstrap process")
		flag.StringVar(&c.DiscoMode, "disco-mode", "", "Choose clustering discovery mode. If not set, no node discovery is performed")
		flag.StringVar(&c.DiscoKey, "disco-key", "rqlite", "Key prefix for cluster discovery service")
		flag.StringVar(&c.DiscoConfig, "disco-c", "", "Set discovery c, or path to cluster discovery c file")
		flag.BoolVar(&c.Expvar, "expvar", true, "Serve expvar data on HTTP server")
		flag.BoolVar(&c.PprofEnabled, "pprof", true, "Serve pprof data on HTTP server")
		flag.BoolVar(&c.OnDisk, "on-disk", false, "Use an on-disk SQLite database")
		flag.StringVar(&c.OnDiskPath, "on-disk-path", "", "Path for SQLite on-disk database file. If not set, use file in data directory")
		flag.BoolVar(&c.OnDiskStartup, "on-disk-startup", false, "Do not initialize on-disk database in memory first at startup")
		flag.BoolVar(&c.FKConstraints, "fk", false, "Enable SQLite foreign key constraints")
		// TODO - flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
		flag.BoolVar(&c.RaftNonVoter, "raft-non-voter", false, "Configure as non-voting node")
		flag.DurationVar(&c.RaftHeartbeatTimeout, "raft-timeout", time.Second, "Raft heartbeat timeout")
		flag.DurationVar(&c.RaftElectionTimeout, "raft-election-timeout", time.Second, "Raft election timeout")
		flag.DurationVar(&c.RaftApplyTimeout, "raft-apply-timeout", 10*time.Second, "Raft apply timeout")
		flag.Uint64Var(&c.RaftSnapThreshold, "raft-snap", 8192, "Number of outstanding log entries that trigger snapshot")
		flag.DurationVar(&c.RaftSnapInterval, "raft-snap-int", 30*time.Second, "Snapshot threshold check interval")
		flag.DurationVar(&c.RaftLeaderLeaseTimeout, "raft-leader-lease-timeout", 0, "Raft leader lease timeout. Use 0s for Raft default")
		flag.BoolVar(&c.RaftStepdownOnShutdown, "raft-shutdown-stepdown", true, "Stepdown as leader before shutting down. Enabled by default")
		flag.BoolVar(&c.RaftShutdownOnRemove, "raft-remove-shutdown", false, "Shutdown Raft if node removed")
		flag.BoolVar(&c.RaftNoFreelistSync, "raft-no-freelist-sync", false, "Do not sync Raft log database freelist to disk")
		flag.StringVar(&c.RaftLogLevel, "raft-log-level", "INFO", "Minimum log level for Raft module")
		flag.DurationVar(&c.RaftReapNodeTimeout, "raft-reap-node-timeout", 0*time.Hour, "Time after which a non-reachable voting node will be reaped. If not set, no reaping takes place")
		flag.DurationVar(&c.RaftReapReadOnlyNodeTimeout, "raft-reap-read-only-node-timeout", 0*time.Hour, "Time after which a non-reachable non-voting node will be reaped. If not set, no reaping takes place")
		flag.DurationVar(&c.ClusterConnectTimeout, "cluster-connect-timeout", 30*time.Second, "Timeout for initial connection to other nodes")
		flag.IntVar(&c.WriteQueueCap, "write-queue-capacity", 1024, "Write queue capacity")
		flag.IntVar(&c.WriteQueueBatchSz, "write-queue-batch-size", 128, "Write queue batch size")
		flag.DurationVar(&c.WriteQueueTimeout, "write-queue-timeout", 50*time.Millisecond, "Write queue timeout")
		flag.BoolVar(&c.WriteQueueTx, "write-queue-tx", false, "Use a transaction when writing from queue")
		flag.IntVar(&c.CompressionSize, "compression-size", 150, "Request query size for compression attempt")
		flag.IntVar(&c.CompressionBatch, "compression-batch", 5, "Request batch threshold for compression attempt")
		flag.StringVar(&c.CPUProfile, "cpu-profile", "", "Path to file for CPU profiling information")
		flag.StringVar(&c.MemProfile, "mem-profile", "", "Path to file for memory profiling information")
		flag.Usage = func() {
			//nolint:forbidigo // TODO
			fmt.Fprintf(os.Stderr, "\n%s\n\n", "Pipelines and workflows for DevSecOps.")
			//nolint:forbidigo // TODO
			fmt.Fprintf(os.Stderr, "Usage: %s [flags] <data directory>\n", "flowpipe")
			flag.PrintDefaults()
		}

		flag.Parse()

		/*
			if showVersion {
				msg := fmt.Sprintf("%s %s %s %s %s sqlite%s (commit %s, branch %s, compiler %s)",
					name, build.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), build.SQLiteVersion,
					build.Commit, build.Branch, runtime.Compiler)
				errorExit(0, msg)
			}
		*/

		// Ensure, if set explicitly, that reap times are not too low.
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "raft-reap-node-timeout" || f.Name == "raft-reap-read-only-node-timeout" {
				d, err := time.ParseDuration(f.Value.String())
				if err != nil {
					errorExit(1, fmt.Sprintf("failed to parse duration: %s", err.Error()))
				}
				if d <= 0 {
					errorExit(1, fmt.Sprintf("-%s must be greater than 0", f.Name))
				}
			}
		})

		/*
			// Ensure the data path is set.
			if flag.NArg() < 1 {
				errorExit(1, "no data directory set")
			}
			c.DataPath = flag.Arg(0)

				// Ensure no args come after the data directory.
				if flag.NArg() > 1 {
					errorExit(1, "arguments after data directory are not accepted")
				}
		*/

		/*
			if err := c.Validate(); err != nil {
				errorExit(1, err.Error())
			}
		*/

		return nil
	}
}

func errorExit(code int, msg string) {
	if code != 0 {
		//nolint:forbidigo // TODO
		fmt.Fprintf(os.Stderr, "fatal: ")
	}
	//nolint:forbidigo // TODO
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(code)
}

// JoinAddresses returns the join addresses set at the command line. Returns nil
// if no join addresses were set.
func (c *Configuration) JoinAddresses() []string {
	if c.JoinAddr == "" {
		return nil
	}
	return strings.Split(c.JoinAddr, ",")
}

// HTTPURL returns the fully-formed, advertised HTTP API address for this config, including
// protocol, host and port.
func (c *Configuration) HTTPURL() string {
	apiProto := "http"
	if c.HTTPx509Cert != "" {
		apiProto = "https"
	}
	return fmt.Sprintf("%s://%s", apiProto, c.HTTPAdv)
}

// DiscoConfigReader returns a ReadCloser providing access to the Disco config.
// The caller must call close on the ReadCloser when finished with it. If no
// config was supplied, it returns nil.
func (c *Configuration) DiscoConfigReader() io.ReadCloser {
	var rc io.ReadCloser
	if c.DiscoConfig == "" {
		return nil
	}

	// Open config file. If opening fails, assume string is the literal config.
	cfgFile, err := os.Open(c.DiscoConfig)
	if err != nil {
		rc = io.NopCloser(bytes.NewReader([]byte(c.DiscoConfig)))
	} else {
		rc = cfgFile
	}
	return rc
}
