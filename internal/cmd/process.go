package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"time"
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

	return cmd

}

// get
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
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()
	return api.GetProcess(executionId)
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

func buildLogDisplayInput(executionId, pipelineExecutionId string) map[string]any {
	return map[string]any{
		"flowpipe": map[string]any{
			"execution_id":          executionId,
			"pipeline_execution_id": pipelineExecutionId,
		},
	}
}
