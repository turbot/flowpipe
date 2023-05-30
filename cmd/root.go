package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/turbot/flowpipe/cmd/pipeline"
	"github.com/turbot/flowpipe/cmd/service"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/constants"
)

// Build the cobra command that handles our command line tool.
func RootCommand(ctx context.Context) (*cobra.Command, error) {

	// Define our command
	rootCmd := &cobra.Command{
		Use:     constants.Name,
		Short:   constants.ShortDescription,
		Long:    constants.LongDescription,
		Version: constants.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			c := config.Config(ctx)
			// Initialize Viper once we start running the command. That means we've
			// got all the flags and can use them to override config values. In
			// particular, we need to know the --config-path (if any) before we
			// can initialize Viper.
			return c.InitializeViper()
		},
	}
	rootCmd.SetVersionTemplate("Flowpipe v{{.Version}}\n")

	c := config.Config(ctx)

	// Command flags
	rootCmd.Flags().StringVar(&c.ConfigPath, "config-path", "", "config file (default is $HOME/.flowpipe/flowpipe.yaml)")
	rootCmd.Flags().StringVar(&c.Workspace, "workspace", "default", "The workspace profile to use")

	// Bind flags to config
	err := c.Viper.BindPFlag("workspace.profile", rootCmd.Flags().Lookup("workspace"))
	if err != nil {
		panic(err)
	}

	// flowpipe service
	serviceCmd, err := service.ServiceCmd(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(serviceCmd)

	pipelineCmd, err := pipeline.PipelineCmd(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(pipelineCmd)

	return rootCmd, nil
}
