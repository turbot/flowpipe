package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

// variable commands
func variableCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "variable",
		Short: "Variable commands",
	}

	cmd.AddCommand(variableListCmd())

	return cmd
}

// list
func variableListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listVariableFunc,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func listVariableFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListVariableResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listVariableRemote()
	} else {
		resp, err = listVariableLocal(cmd, args)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer := printers.GetPrinter[types.Variable](cmd)
		printableResource := types.NewPrintableVariable(resp)
		err := printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func listVariableRemote() (*types.ListVariableResponse, error) {
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.VariableApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err
	}
	// map the API data type into the internal data type
	return types.ListVariableResponseFromAPIResponse(resp)
}

func listVariableLocal(cmd *cobra.Command, args []string) (*types.ListVariableResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	// TODO: figure out if better approach
	var vars []types.Variable
	if m.RootMod != nil && m.RootMod.ResourceMaps != nil && len(m.RootMod.ResourceMaps.Variables) > 0 {
		for k, m := range m.RootMod.ResourceMaps.Variables {
			vars = append(vars, types.Variable{
				Name:        k,
				Type:        m.TypeString,
				Description: m.Description,
				Value:       m.ValueGo,
				Default:     m.DefaultGo,
			})
		}
		return &types.ListVariableResponse{
			Items:     vars,
			NextToken: nil,
		}, nil
	}

	return nil, nil
}
