package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/types"
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
		_ = m.Stop()
	}()

	return api.ListIntegrations()
}

func listIntegrationRemote(ctx context.Context) (*types.ListIntegrationResponse, error) {
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.IntegrationApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err

	}
	// map the API data type into the internal data type
	return types.ListIntegrationResponseFromAPI(resp), err
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
	ctx := cmd.Context()
	var resp *types.FpIntegration
	var err error
	integrationName := args[0]
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = getIntegrationRemote(ctx, integrationName)
	} else {
		resp, err = getIntegrationLocal(ctx, integrationName)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpIntegration](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableIntegrationFromSingle(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func getIntegrationRemote(ctx context.Context, name string) (*types.FpIntegration, error) {
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.IntegrationApi.Get(ctx, name).Execute()
	if err != nil {
		return nil, err
	}
	t := types.FpIntegrationFromAPI(*resp)
	return &t, nil
}

func getIntegrationLocal(ctx context.Context, integrationName string) (*types.FpIntegration, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	return api.GetIntegration(integrationName)
}
