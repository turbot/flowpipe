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

func variableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Variable commands",
	}

	cmd.AddCommand(variableListCmd())
	cmd.AddCommand(variableShowCmd())

	return cmd
}

func variableShowCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show",
		Args:  cobra.ExactArgs(1),
		Run:   showVariableFunc,
		Short: "Show variable from the current mod",
		Long:  `Show variable from the current mod.`,
	}

	return cmd
}

func variableListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   listVariableFunc,
		Short: "List variables from the current mod",
		Long:  `List variables from the current mod.`,
	}

	return cmd
}

func listVariableFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListVariableResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listVariableRemote(ctx)
	} else {
		resp, err = listVariableLocal(cmd, args)
	}
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Error listing variables transforming")
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpVariable](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableVariable(resp)

		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func listVariableLocal(cmd *cobra.Command, args []string) (*types.ListVariableResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		_ = m.Stop()
	}()

	return api.ListVariables()
}

func listVariableRemote(ctx context.Context) (*types.ListVariableResponse, error) {
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.VariableApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err

	}

	return types.ListVariableResponseFromAPI(resp), err
}

func showVariableFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.FpVariable
	var err error
	variableName := args[0]
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = getVariableRemote(ctx, variableName)
	} else {
		resp, err = getVariableLocal(ctx, variableName)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpVariable](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintableVariableFromSingle(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}
func getVariableLocal(ctx context.Context, variableName string) (*types.FpVariable, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	return api.GetVariable(variableName)
}

func getVariableRemote(ctx context.Context, name string) (*types.FpVariable, error) {
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.VariableApi.Get(ctx, name).Execute()
	if err != nil {
		return nil, err
	}
	t := types.FpVariableFromApi(*resp)
	return t, nil
}
