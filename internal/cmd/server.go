package cmd

import (
	"fmt"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/types"
	"net"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	serviceConfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
)

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Args:  cobra.NoArgs,
		Short: "Run the Flowpipe server, including triggers, integrations, and the API",
		Long:  `Run the Flowpipe server, including triggers, integrations, and the API.`,
		Run:   startServerFunc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {

			// TODO KAI look at whether this is really needed
			serviceConfig.Initialize()
			return nil
		},
	}

	cmdconfig.
		OnCmd(cmd).
		AddIntFlag(constants.ArgPort, localconstants.DefaultServerPort, "Server port.").
		AddStringFlag(constants.ArgListen, localconstants.DefaultListen, "Listen address port.").
		AddBoolFlag(constants.ArgWatch, true, "Watch mod files for changes when running Flowpipe server").
		AddBoolFlag(constants.ArgVerbose, false, "Enable verbose output")

	return cmd
}

func startServerFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		output.IsServerMode = true

		// Check if the port is already in use
		if isPortInUse(viper.GetInt(constants.ArgPort)) {
			errMsg := fmt.Sprintf("the designated port (%d) is already in use", viper.GetInt(constants.ArgPort))
			output.RenderServerOutput(ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "unable to start server", fmt.Errorf(errMsg)))
			error_helpers.FailOnError(perr.InternalWithMessage(errMsg))
			return
		}

		outputMode := viper.GetString(constants.ArgOutput)
		if outputMode == constants.OutputFormatJSON || outputMode == constants.OutputFormatYAML {
			errMsg := "currently '--output' is not supported for json or yaml"
			output.RenderServerOutput(ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "unable to start server", fmt.Errorf(errMsg)))
			error_helpers.FailOnError(perr.BadRequestWithMessage(errMsg))
			return
		}

		// start manager, passing server config
		// (this will ensure manager starts API, ES, Scheduling and docker services
		m, err := manager.NewManager(ctx,
			manager.WithServerConfig(viper.GetString(constants.ArgListen), viper.GetInt(constants.ArgPort)),
		).Start()
		if err != nil {
			output.RenderServerOutput(ctx, types.NewServerOutputError(types.NewServerOutputPrefix(time.Now(), "flowpipe"), "unable to start server", err))
		}
		error_helpers.FailOnError(err)

		// Block until we receive a signal
		m.InterruptHandler()
	}
}

// Function to check if a port is in use
func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}
