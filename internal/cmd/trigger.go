package cmd

import (
	"context"
	"fmt"
	"time"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/utils"

	"github.com/spf13/viper"
	o "github.com/turbot/flowpipe/internal/output"
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
	cmd.AddCommand(triggerRunCmd())

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

func triggerRunCmd() *cobra.Command {
	// only for local for now

	var cmd = &cobra.Command{
		Use:   "run <trigger-name>",
		Args:  cobra.ExactArgs(1),
		Run:   runTriggerFunc,
		Short: "Run a trigger from the current mod.",
		Long:  `Run a trigger from the current mod.`,
	}

	// Add the pipeline arg flag
	cmdconfig.OnCmd(cmd).
		AddStringArrayFlag(constants.ArgArg, nil, "Specify the value of a trigger argument. Multiple --arg may be passed.").
		AddBoolFlag(constants.ArgVerbose, false, "Enable verbose output.").
		AddBoolFlag(constants.ArgDetach, false, "Run the trigger in detached mode.").
		AddStringFlag(constants.ArgExecutionId, "", "Specify trigger execution id. Execution id will generated if not provided.")

	return cmd
}

func runTriggerLocal(cmd *cobra.Command, args []string) (types.TriggerExecutionResponse, *manager.Manager, error) {
	ctx := cmd.Context()

	response := types.TriggerExecutionResponse{}

	// create and start the manager with ES service, and Docker, but no API server
	// Move all this code to "run local"
	// create and start the manager with ES service, and Docker, but no API server
	m, err := manager.NewManager(ctx, manager.WithESService()).Start()
	if err != nil {
		error_helpers.FailOnError(err)
		return response, nil, err
	}

	triggerName := api.ConstructTriggerFullyQualifiedName(args[0])

	// extract the trigger args from the flags
	triggerArgs := getPipelineArgs(cmd)

	executionId, err := cmd.Flags().GetString(constants.ArgExecutionId)
	if err != nil {
		return response, nil, err
	}

	input := types.CmdTrigger{
		Command:    "run",
		ArgsString: triggerArgs,
	}

	response, _, err = api.ExecuteTrigger(ctx, input, executionId, triggerName, m.ESService)

	return response, m, err
}

func runTriggerRemote(cmd *cobra.Command, args []string) (types.TriggerExecutionResponse, error) {
	ctx := cmd.Context()

	executionId, err := cmd.Flags().GetString(constants.ArgExecutionId)
	if err != nil {
		return types.TriggerExecutionResponse{}, err
	}

	triggerName := api.ConstructTriggerFullyQualifiedName(args[0])

	// extract the trigger args from the flags
	triggerArgs := getPipelineArgs(cmd)

	// API client
	apiClient := common.GetApiClient()
	cmdTriggerRun := flowpipeapiclient.NewCmdTrigger("run")

	// Set the pipeline args
	cmdTriggerRun.ArgsString = &triggerArgs
	cmdTriggerRun.ExecutionId = &executionId

	apiResponse, _, err := apiClient.TriggerApi.Command(ctx, triggerName).Request(*cmdTriggerRun).Execute()
	if err != nil {
		return types.TriggerExecutionResponse{}, err
	}

	resp := TriggerExecutionResponseFromAPIResponse(*apiResponse)
	return resp, err
}

func TriggerExecutionResponseFromAPIResponse(apiResponse flowpipeapiclient.TriggerExecutionResponse) types.TriggerExecutionResponse {
	var results []types.PipelineExecutionResponse
	for _, apiResult := range apiResponse.Results {
		results = append(results, PipelineExecutionResponseFromAPIResponse(apiResult))
	}

	resp := types.TriggerExecutionResponse{
		Results: results,
		Flowpipe: types.FlowpipeTriggerResponseMetadata{
			IsStale:   apiResponse.Flowpipe.IsStale,
			ProcessID: utils.Deref(apiResponse.Flowpipe.ProcessId, ""),
			Name:      utils.Deref(apiResponse.Flowpipe.Name, ""),
			Type:      utils.Deref(apiResponse.Flowpipe.Type, ""),
		},
	}

	if apiResponse.Flowpipe.LastLoaded != nil {
		time, _ := time.Parse(utils.RFC3339WithMS, *apiResponse.Flowpipe.LastLoaded)
		resp.Flowpipe.LastLoaded = &time
	}

	return resp
}

func executeTrigger(cmd *cobra.Command, args []string, isRemote bool) (*manager.Manager, types.TriggerExecutionResponse, pollEventLogFunc, error) {
	if isRemote {
		// run trigger on server
		resp, err := runTriggerRemote(cmd, args)
		pollLogFunc := pollServerEventLog
		return nil, resp, pollLogFunc, err
	}
	// run trigger in-process
	var m *manager.Manager
	resp, m, err := runTriggerLocal(cmd, args)

	pollLogFunc := pollLocalEventLog
	return m, resp, pollLogFunc, err
}

func runTriggerFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp types.TriggerExecutionResponse
	var err error
	var pollLogFunc pollEventLogFunc

	isDetach := viper.GetBool(constants.ArgDetach)
	isRemote := viper.IsSet(constants.ArgHost)
	isVerbose := viper.IsSet(constants.ArgVerbose)
	if !isRemote && isDetach {
		error_helpers.ShowError(ctx, fmt.Errorf("unable to use --detach with local execution"))
		return
	}

	output := viper.GetString(constants.ArgOutput)
	streamLogs := (output == "plain" || output == "pretty") && (o.IsServerMode || isRemote || isVerbose)
	progressLogs := (output == "plain" || output == "pretty") && !o.IsServerMode && !isRemote && !isVerbose
	if progressLogs {
		o.PipelineProgress = o.NewProgress("Initializing...")
	}

	// if a host is set, use it to connect to API server
	var m *manager.Manager
	m, resp, pollLogFunc, err = executeTrigger(cmd, args, isRemote)
	if err != nil {
		error_helpers.FailOnErrorWithMessage(err, "failed executing pipeline")
		return
	}

	// ensure to shut the manager when we are done
	defer func() {
		if m != nil {
			_ = m.Stop()
		}
	}()

	for _, pipelineExecutionResponse := range resp.Results {
		switch {
		case isDetach:
			err := displayDetached(ctx, cmd, pipelineExecutionResponse)
			if err != nil {
				error_helpers.FailOnErrorWithMessage(err, "failed printing execution information")
				return
			}
		case streamLogs:
			displayStreamingLogs(ctx, cmd, pipelineExecutionResponse, pollLogFunc)
		case progressLogs:
			displayProgressLogs(ctx, cmd, pipelineExecutionResponse, pollLogFunc)
		default:
			displayBasicOutput(ctx, cmd, pipelineExecutionResponse, pollLogFunc)
		}
	}
}
