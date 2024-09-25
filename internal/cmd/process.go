package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cmd/common"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/printers"
)

// process commands
func processCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process commands",
	}

	cmd.AddCommand(processShowCmd())
	cmd.AddCommand(processListCmd())
	cmd.AddCommand(processTailCmd())
	cmd.AddCommand(processResumeCmd())

	return cmd
}

func processResumeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "resume <execution-id>",
		Args:  cobra.ExactArgs(1),
		Run:   resumeProcessFunc,
		Short: "Resume a process",
		Long:  `Resume a process.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func processShowCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <execution-id>",
		Args:  cobra.ExactArgs(1),
		Run:   showProcessFunc,
		Short: "Show details for a single process",
		Long:  `Show details for a single process.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func resumeProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var err error
	executionId := args[0]

	var m *manager.Manager

	var resp types.PipelineExecutionResponse
	var pollLogFunc pollEventLogFunc

	if viper.IsSet(constants.ArgHost) {
		_, err = getProcessRemote(executionId)
	} else {
		m, resp, err = resumeProcessLocal(ctx, executionId)
		pollLogFunc = pollLocalEventLog
	}
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	defer func() {
		if m != nil {
			_ = m.Stop()
		}
	}()

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
		o.PipelineProgress = o.NewProgress("Resuming...")
	}

	switch {
	case isDetach:
		err := displayDetached(ctx, cmd, resp)
		if err != nil {
			error_helpers.FailOnErrorWithMessage(err, "failed printing execution information")
			return
		}
	case streamLogs:
		displayStreamingLogs(ctx, cmd, resp, pollLogFunc)
	case progressLogs:
		displayProgressLogs(ctx, cmd, resp, pollLogFunc)
	default:
		displayBasicOutput(ctx, cmd, resp, pollLogFunc)
	}

}

func showProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.Process
	var err error
	executionId := args[0]

	if viper.IsSet(constants.ArgHost) {
		resp, err = getProcessRemote(executionId)
	} else {
		resp, err = getProcessLocal(ctx, executionId)
	}
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.Process](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed obtaining printer")
			return
		}
		printableResource := types.NewPrintableProcessFromSingle(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed when printing")
			return
		}
	}
}

func getProcessRemote(executionId string) (*types.Process, error) {
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.ProcessApi.Get(context.Background(), executionId).Execute()
	if err != nil {
		return nil, err
	}
	// map the API data type into the internal data type
	return types.ProcessFromAPIResponse(*resp)
}

func getProcessLocal(ctx context.Context, executionId string) (*types.Process, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		_ = m.Stop()
	}()
	return api.GetProcess(executionId)
}

func resumeProcessLocal(ctx context.Context, executionId string) (*manager.Manager, types.PipelineExecutionResponse, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx, manager.WithESService()).Start()
	error_helpers.FailOnError(err)

	pipelineExecutionId, pipelineName, err := api.ResumeProcess(executionId, m.ESService)

	if err != nil {
		return nil, types.PipelineExecutionResponse{}, err
	}

	response := types.PipelineExecutionResponse{}
	response.Flowpipe = types.FlowpipeResponseMetadata{
		ExecutionID:         executionId,
		PipelineExecutionID: pipelineExecutionId,
		Pipeline:            pipelineName,
	}

	return m, response, nil
}

// list
func processListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   listProcessFunc,
		Short: "List processes",
		Long:  `List processes.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func listProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListProcessResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listProcessRemote()
	} else {
		resp, err = listProcessLocal(cmd, args)
	}
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.Process](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
			return
		}
		printableResource := types.NewPrintableProcess(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
			return
		}
	}
}

func listProcessRemote() (*types.ListProcessResponse, error) {
	ctx := context.Background()
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()

	resp, _, err := apiClient.ProcessApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err
	}
	// map the API data type into the internal data type
	return types.ListProcessResponseFromAPIResponse(resp)
}

func listProcessLocal(cmd *cobra.Command, args []string) (*types.ListProcessResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	return api.ListProcesses()
}

// tail
func processTailCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "tail <execution-id>",
		Args:  cobra.ExactArgs(1),
		Run:   tailProcessFunc,
		Short: "Tail a single process",
		Long:  `Tail a single process.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgVerbose, false, "Enable verbose output.")

	return cmd
}

func tailProcessFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	executionId := args[0]
	isRemote := viper.IsSet(constants.ArgHost)
	tailStart := time.Now()
	var err error

	if isRemote {
		err = tailProcessRemote(ctx, cmd, executionId, tailStart)
	} else {
		err = tailProcessLocal()
	}
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}
}

func tailProcessRemote(ctx context.Context, cmd *cobra.Command, execId string, tailStart time.Time) error {
	// TODO: Need a better approach to get outer pipeline id
	var pipelineExecutionId string
	var pipelineStatus string
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.ProcessApi.GetExecution(ctx, execId).Execute()
	if err != nil {
		return err
	}
	if resp != nil && resp.HasPipelineExecutions() {
		for _, pl := range *resp.PipelineExecutions {
			if pl.GetParentExecutionId() == "" && pl.GetParentStepExecutionId() == "" {
				pipelineExecutionId = *pl.Id
				pipelineStatus = *pl.Status
				continue
			}
		}
	} else {
		return fmt.Errorf("failed to obtain process")
	}

	output := viper.GetString(constants.ArgOutput)
	isStreamingLogs := output == "pretty" || output == "plain"
	pipelineComplete := pipelineStatus == "finished" || pipelineStatus == "failed"
	input := buildLogDisplayInput(execId, pipelineExecutionId)

	switch {
	case !isStreamingLogs && !pipelineComplete:
		return fmt.Errorf("--output %s may only be used to tail a process which has completed", output)
	case !isStreamingLogs && pipelineComplete:
		displayBasicOutput(ctx, cmd, input, pollServerEventLog)
		return nil
	case isStreamingLogs:
		displayStreamingLogs(ctx, cmd, input, pollServerEventLog)
		return nil
	default:
		return nil
	}
}

func tailProcessLocal() error {
	return fmt.Errorf("tail requires a remote server via --host <host>")
}

func buildLogDisplayInput(executionId, pipelineExecutionId string) types.PipelineExecutionResponse {
	return types.PipelineExecutionResponse{
		Flowpipe: types.FlowpipeResponseMetadata{
			ExecutionID:         executionId,
			PipelineExecutionID: pipelineExecutionId,
		},
	}
}
