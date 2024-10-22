package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/fperr"
	"github.com/turbot/flowpipe/internal/output"
	serviceConfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
)

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Args:  cobra.NoArgs,
		Short: "Run the Flowpipe server, including triggers, integrations, and the API",
		Long:  `Run the Flowpipe server, including triggers, integrations, and the API.`,
		Run:   startServerFunc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {

			serviceConfig.Initialize()
			return nil
		},
	}

	cmdconfig.
		OnCmd(cmd).
		AddIntFlag(constants.ArgPort, localconstants.DefaultServerPort, "Server port.").
		AddStringFlag(constants.ArgListen, localconstants.DefaultListen, "Listen address port.").
		AddStringFlag(constants.ArgBaseUrl, localconstants.DefaultFlowpipeHost, "Base URL for the webhook triggers and http input ("+localconstants.DefaultFlowpipeHost+").").
		AddBoolFlag(constants.ArgWatch, true, "Watch mod files for changes when running Flowpipe server").
		AddBoolFlag(constants.ArgVerbose, false, "Enable verbose output")

	return cmd
}

// TODO: revisit exit codes
func startServerFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		output.IsServerMode = true

		// Check if the port is already in use
		if isPortInUse(viper.GetInt(constants.ArgPort)) {
			errMsg := fmt.Sprintf("the designated port (%d) is already in use", viper.GetInt(constants.ArgPort))
			output.RenderServerOutput(ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "unable to start server", errors.New(errMsg)))
			os.Exit(constants.ExitCodeBindPortUnavailable)
		}

		outputMode := viper.GetString(constants.ArgOutput)
		if outputMode == constants.OutputFormatJSON || outputMode == constants.OutputFormatYAML {
			errMsg := "server command currently only supports '--output' for 'pretty' or 'plain'"
			output.RenderServerOutput(ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "unable to start server", errors.New(errMsg)))
			os.Exit(constants.ExitCodeInsufficientOrWrongInputs)
		}

		// start manager, passing server config
		// (this will ensure manager starts API, ES, Scheduling and docker services
		m, err := manager.NewManager(ctx,
			manager.WithServerConfig(viper.GetString(constants.ArgListen), viper.GetInt(constants.ArgPort)),
		).Start()
		if err != nil {
			output.RenderServerOutput(ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "unable to start server", err))
			fperr.FailOnError(err, nil, "")
		}

		// Block until we receive a signal
		m.InterruptHandler()
	}
}

// Function to check if a port is in use
func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}
