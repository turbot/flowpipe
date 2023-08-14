package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"

	"github.com/turbot/flowpipe/internal/cmd/mod"
	"github.com/turbot/flowpipe/internal/cmd/pipeline"
	"github.com/turbot/flowpipe/internal/cmd/process"
	"github.com/turbot/flowpipe/internal/cmd/service"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/types"
)

// ④ Now use the FooMode enum flag. If you want a non-zero default, then
// simply set it here, such as in "foomode = Bar".
var outputMode types.OutputMode

// Build the cobra command that handles our command line tool.
func RootCommand(ctx context.Context) (*cobra.Command, error) {

	// Define our command
	rootCmd := &cobra.Command{
		Use:     constants.Name,
		Short:   constants.ShortDescription,
		Long:    constants.LongDescription,
		Version: constants.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	rootCmd.SetVersionTemplate("Flowpipe v{{.Version}}\n")

	c := config.GetConfigFromContext(ctx)

	// Command flags
	rootCmd.Flags().StringVar(&c.ConfigPath, "config-path", "", "config file (default is $HOME/.flowpipe/flowpipe.yaml)")

	rootCmd.PersistentFlags().String(constants.CmdOptionApiHost, "http://localhost", "API server host")
	rootCmd.PersistentFlags().Int(constants.CmdOptionApiPort, 7103, "API server port")
	rootCmd.PersistentFlags().Bool(constants.CmdOptionTlsInsecure, false, "Skip TLS verification")

	// ⑤ Define the CLI flag parameters for your wrapped enum flag.
	rootCmd.PersistentFlags().Var(
		enumflag.New(&outputMode, constants.CmdOptionsOutput, types.OutputModeIds, enumflag.EnumCaseInsensitive),
		constants.CmdOptionsOutput,
		"Output format; one of: table, yaml, json")

	err := viper.BindPFlag("api.host", rootCmd.PersistentFlags().Lookup(constants.CmdOptionApiHost))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("api.port", rootCmd.PersistentFlags().Lookup(constants.CmdOptionApiPort))
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("api.tls_insecure", rootCmd.PersistentFlags().Lookup(constants.CmdOptionTlsInsecure))
	if err != nil {
		log.Fatal(err)
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

	processCmd, err := process.ProcessCmd(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(processCmd)

	modCmd, err := mod.ModCmd(ctx)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(modCmd)

	return rootCmd, nil
}
