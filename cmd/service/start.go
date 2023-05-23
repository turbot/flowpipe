package service

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/config"
	serviceConfig "github.com/turbot/flowpipe/service/config"
	"github.com/turbot/flowpipe/service/manager"
)

func ServiceStartCmd(ctx context.Context) (*cobra.Command, error) {

	var serviceStartCmd = &cobra.Command{
		Use:  "start",
		Args: cobra.NoArgs,
		Run:  startManagerFunc(ctx),
	}

	c := config.Config(ctx)

	// Command flags
	serviceStartCmd.Flags().StringVar(&c.DataPath, "data-path", "/Users/nathan/.flowpipe/raft", "path to data directory")
	serviceStartCmd.Flags().StringVar(&c.HTTPAddr, "https-address", "", "host:port of the HTTPS server")
	serviceStartCmd.Flags().StringVar(&c.RaftAddr, "raft-address", "", "host:port of the raft server")
	serviceStartCmd.Flags().StringVar(&c.JoinAddr, "join", "", "comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	serviceStartCmd.Flags().BoolVar(&c.RaftBootstrap, "raft-bootstrap", false, "if true, then bootstrap the cluster")
	serviceStartCmd.Flags().StringVar(&c.RaftNodeID, "raft-node-id", "", "unique ID for the node")

	// Bind flags to config
	err := c.Viper.BindPFlag("server.data_path", serviceStartCmd.Flags().Lookup("data-path"))
	if err != nil {
		panic(err)
	}

	err = c.Viper.BindPFlag("server.https_address", serviceStartCmd.Flags().Lookup("https-address"))
	if err != nil {
		panic(err)
	}

	err = c.Viper.BindPFlag("server.raft_address", serviceStartCmd.Flags().Lookup("raft-address"))
	if err != nil {
		panic(err)
	}

	err = c.Viper.BindPFlag("server.join", serviceStartCmd.Flags().Lookup("join"))
	if err != nil {
		panic(err)
	}

	err = c.Viper.BindPFlag("server.raft_bootstrap", serviceStartCmd.Flags().Lookup("raft-bootstrap"))
	if err != nil {
		panic(err)
	}

	err = c.Viper.BindPFlag("server.raft_node_id", serviceStartCmd.Flags().Lookup("raft-node-id"))
	if err != nil {
		panic(err)
	}

	return serviceStartCmd, nil
}

func startManagerFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		c := config.Config(ctx)
		serviceConfig.Initialize(ctx)
		m, err := manager.NewManager(ctx,
			manager.WithHTTPAddress(c.HTTPAddr),
			manager.WithRaftNodeID(c.RaftNodeID),
			manager.WithRaftBootstrap(c.RaftBootstrap),
			manager.WithRaftAddress(c.RaftAddr))
		if err != nil {
			panic(err)
		}
		// Start the manager
		err = m.Start()
		if err != nil {
			panic(err)
		}
		// Block until we receive a signal
		m.InterruptHandler()
	}
}
