package cmd

import (
	"context"

	"github.com/turbot/pipe-fittings/printers"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
)

func triggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger commands",
	}

	cmd.AddCommand(triggerListCmd())
	cmd.AddCommand(triggerShowCmd())

	return cmd
}

// list
func triggerListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   listTriggerFunc,
		Short: "List triggers from the current mod",
		Long:  `List triggers from the current mod.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func listTriggerFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListTriggerResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listTriggerRemote(ctx)
	} else {
		resp, err = listTriggerLocal(cmd, args)
	}
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Error listing triggers transforming")
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpTrigger](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableTrigger(resp)

		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func listTriggerRemote(ctx context.Context) (*types.ListTriggerResponse, error) {
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.TriggerApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err

	}
	// map the API data type into the internal data type
	return types.ListTriggerResponseFromAPI(resp), err
}

func listTriggerLocal(cmd *cobra.Command, args []string) (*types.ListTriggerResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// ignore shutdown error?
		_ = m.Stop()
	}()

	// now list the pipelines
	return api.ListTriggers()
}

// show

func triggerShowCmd() *cobra.Command {
	var triggerShowCmd = &cobra.Command{
		Use:   "show <trigger-name>",
		Args:  cobra.ExactArgs(1),
		Run:   showTriggerFunc,
		Short: "Show details of a trigger from the current mod",
		Long:  `Show details of a trigger from the current mod.`,
	}

	// initialize hooks
	cmdconfig.OnCmd(triggerShowCmd)

	return triggerShowCmd
}

func showTriggerFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.FpTrigger
	var err error
	triggerName := args[0]
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = getTriggerRemote(ctx, triggerName)
	} else {
		resp, err = getTriggerLocal(ctx, triggerName)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpTrigger](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableTriggerFromSingle(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func getTriggerRemote(ctx context.Context, name string) (*types.FpTrigger, error) {
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.TriggerApi.Get(ctx, name).Execute()
	if err != nil {
		return nil, err
	}
	t := types.FpTriggerFromAPI(*resp)
	return &t, nil
}

func getTriggerLocal(ctx context.Context, triggerName string) (*types.FpTrigger, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	// try to fetch the pipeline from the cache
	return api.GetTrigger(triggerName)
}
