package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/printers"
)

func integrationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integration",
		Short: "Integration commands",
	}

	cmd.AddCommand(integrationListCmd())
	cmd.AddCommand(integrationShowCmd())

	return cmd
}

// list
func integrationListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   listIntegrationFunc,
		Short: "List integrations from the current mod",
		Long:  `List integrations from the current mod.`,
	}

	return cmd
}

func listIntegrationFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListIntegrationResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listIntegrationRemote(ctx)
	} else {
		resp, err = listIntegrationLocal(cmd, args)
	}
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Error listing triggers transforming")
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpIntegration](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableIntegration(resp)

		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func listIntegrationLocal(cmd *cobra.Command, args []string) (*types.ListIntegrationResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// ignore shutdown error?
		_ = m.Stop()
	}()

	// now list the pipelines
	return api.ListIntegrations()
}

func listIntegrationRemote(ctx context.Context) (*types.ListIntegrationResponse, error) {
	return nil, nil
}

// show
func integrationShowCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show",
		Args:  cobra.ExactArgs(1),
		Run:   showIntegrationFunc,
		Short: "Show integration from the current mod",
		Long:  `Show integration from the current mod.`,
	}

	return cmd
}

func showIntegrationFunc(cmd *cobra.Command, args []string) {

}
