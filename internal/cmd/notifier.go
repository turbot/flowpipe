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

func notifierCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notifier",
		Short: "Notifier commands",
	}

	cmd.AddCommand(notifierListCmd())
	cmd.AddCommand(notifierShowCmd())

	return cmd
}

// list
func notifierListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   listNotifierFunc,
		Short: "List notifier from the current mod",
		Long:  `List notifier from the current mod.`,
	}

	return cmd
}

func listNotifierFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListNotifierResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listNotifierRemote(ctx)
	} else {
		resp, err = listNotifierLocal(cmd, args)
	}
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Error listing triggers transforming")
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpNotifier](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableNotifier(resp)

		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func listNotifierLocal(cmd *cobra.Command, args []string) (*types.ListNotifierResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		_ = m.Stop()
	}()

	return api.ListNotifiers()
}

func listNotifierRemote(ctx context.Context) (*types.ListNotifierResponse, error) {
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.NotifierApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err

	}
	// map the API data type into the internal data type
	return types.ListNotifierResponseFromAPI(resp), err
}

// show
func notifierShowCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show",
		Args:  cobra.ExactArgs(1),
		Run:   showNotifierFunc,
		Short: "Show notifier from the current mod",
		Long:  `Show notifier from the current mod.`,
	}

	return cmd
}

func showNotifierFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.FpNotifier
	var err error
	notifierName := args[0]
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = getNotifierRemote(ctx, notifierName)
	} else {
		resp, err = getNotifierLocal(ctx, notifierName)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpNotifier](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableNotifierFromSingle(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func getNotifierRemote(ctx context.Context, name string) (*types.FpNotifier, error) {
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.NotifierApi.Get(ctx, name).Execute()
	if err != nil {
		return nil, err
	}
	t := types.FpNotifierFromAPI(*resp)
	return &t, nil
}

func getNotifierLocal(ctx context.Context, notifierName string) (*types.FpNotifier, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	return api.GetNotifier(notifierName)
}
