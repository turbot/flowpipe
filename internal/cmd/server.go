package cmd

import (
	"fmt"
	"github.com/turbot/flowpipe/internal/output"
	"net"
	"strconv"

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

		// Check if the port is already in use
		if isPortInUse(viper.GetInt(constants.ArgPort)) {
			error_helpers.FailOnError(perr.InternalWithMessage("The designated port (" + strconv.Itoa(viper.GetInt(constants.ArgPort)) + ") is already in use"))
			return
		}

		// Error on unsupported JSON/YAML outputs, convert pretty to plain (no color)
		switch viper.GetString(constants.ArgOutput) {
		case constants.OutputFormatJSON, constants.OutputFormatYAML:
			error_helpers.FailOnError(perr.BadRequestWithMessage("Currently '--output' is not supported for json or yaml"))
			return
		case constants.OutputFormatPretty:
			viper.Set(constants.ArgOutput, constants.OutputFormatPlain)
		}

		output.IsServerMode = true

		// start manager, passing server config
		// (this will ensure manager starts API, ES, Scheduling and docker services
		m, err := manager.NewManager(ctx,
			manager.WithServerConfig(viper.GetString(constants.ArgListen), viper.GetInt(constants.ArgPort)),
		).Start()
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
