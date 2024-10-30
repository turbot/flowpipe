package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/es/command"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fperr"
	o "github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/printers"
	"github.com/turbot/pipe-fittings/schema"
	"golang.org/x/exp/maps"
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

	exitCode := 0
	lastStatus := ""

	defer func() {
		if m != nil {
			_ = m.Stop()
		}
		slog.Debug("Completed execution from resumeProcessFunc", "status", lastStatus, "exitCode", exitCode)
		os.Exit(exitCode)
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
		lastStatus = displayStreamingLogs(ctx, cmd, resp, pollLogFunc)
	case progressLogs:
		lastStatus = displayProgressLogs(ctx, cmd, resp, pollLogFunc)
	default:
		lastStatus = displayBasicOutput(ctx, cmd, resp, pollLogFunc)
	}

	switch lastStatus {
	case event.HandlerExecutionFailed:
		exitCode = fperr.ExitCodeExecutionFailed
	case event.HandlerExecutionCancelled:
		exitCode = fperr.ExitCodeExecutionCancelled
	case event.HandlerExecutionPaused:
		exitCode = fperr.ExitCodeExecutionPaused
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

	// this applies to local only (flowpipe process resume)
	evt := &event.Event{
		ExecutionID: executionId,
	}

	ex, err := execution.LoadExecutionFromProcessDB(evt)
	if err != nil {
		return nil, types.PipelineExecutionResponse{}, err
	}

	// restart all input step poller (if there's any)
	if routerUrl, routed := primitive.GetInputRouter(); routed {
		type definitions struct {
			PipelineDef       *resources.Pipeline
			StepDef           resources.PipelineStep
			PipelineExecution *execution.PipelineExecution
			StepExecution     *execution.StepExecution
		}
		// obtain any unfinished input steps & store pipeline definition as we need this to skip poller and end step if done
		unfinishedInputSteps := make(map[string]definitions)
		for _, pex := range ex.PipelineExecutions {
			pipelineDefn, err := ex.PipelineDefinition(pex.ID)
			if err != nil {
				return nil, types.PipelineExecutionResponse{}, err
			}

			for _, se := range pex.StepExecutions {
				if se.Status == "starting" {
					if pipelineDefn.GetStep(se.Name).GetType() == schema.BlockTypePipelineStepInput {
						unfinishedInputSteps[se.ID] = definitions{
							PipelineDef:       pipelineDefn,
							PipelineExecution: pex,
							StepDef:           pipelineDefn.GetStep(se.Name),
							StepExecution:     se,
						}
					}
				}
			}
		}

		if len(unfinishedInputSteps) > 0 {

			steps := maps.Keys(unfinishedInputSteps)
			token := os.Getenv(app_specific.EnvPipesToken)
			if token == "" {
				return nil, types.PipelineExecutionResponse{}, perr.InternalWithMessage("Missing token for routed input. Please set " + app_specific.EnvPipesToken + " env variable.")
			}

			inputs, e := obtainRoutedInputsForExecution(routerUrl, token, steps)
			if e != nil {
				return nil, types.PipelineExecutionResponse{}, e
			}

			for _, input := range inputs.Items {
				switch input.State {
				case "pending", "started":
					if step, ok := unfinishedInputSteps[input.StepExecutionID]; ok {
						slog.Info("Resuming input step poller", "step", input.StepExecutionID)
						endStepFunc := func(stepExecution *execution.StepExecution, out *resources.Output) error {
							return command.EndStepFromApi(ex, stepExecution, step.PipelineDef, step.StepDef, out, m.ESService.EventBus)
						}
						p := primitive.NewRoutedInput(executionId, step.PipelineExecution.ID, step.StepExecution.ID, step.PipelineDef.PipelineName, step.StepExecution.Name, step.StepDef.GetType(), routerUrl, endStepFunc)
						p.Poll(ctx, &http.Client{}, token, input.ID)
					}
				case "finished":
					if step, ok := unfinishedInputSteps[input.StepExecutionID]; ok {
						slog.Info("Resuming input step with result", "step", input.StepExecutionID)
						stepShortName := strings.Split(step.StepExecution.Name, ".")[len(strings.Split(step.StepExecution.Name, "."))-1]
						out := resources.Output{
							Data: map[string]any{
								"value": input.Inputs[stepShortName].Response,
							},
							Status: "finished",
						}
						e := command.EndStepFromApi(ex, step.StepExecution, step.PipelineDef, step.StepDef, &out, m.ESService.EventBus)
						if e != nil {
							return nil, types.PipelineExecutionResponse{}, perr.InternalWithMessage("Failed finishing input step.")
						}
					}
				case "error":
					// TODO: #refactor #error Handle failed inputs from Pipes
				}
			}
		}
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

func obtainRoutedInputsForExecution(routerUrl string, token string, stepKeys []string) (*primitive.RoutedInputListResponse, error) {
	client := &http.Client{}

	inputItems := make([]primitive.RoutedInputResponse, 0)
	for i := range stepKeys {
		getInputUrl := fmt.Sprintf("%s/%s", routerUrl, stepKeys[i])

		req, e := http.NewRequest(http.MethodGet, getInputUrl, nil)
		if e != nil {
			return nil, perr.InternalWithMessage("Failed to build request to obtain routed inputs.")
		}

		req.Header.Set("Authorization", "Bearer "+token)
		resp, e := client.Do(req)
		if e != nil {
			return nil, perr.InternalWithMessage("Failed to get routed input details.")
		}
		defer resp.Body.Close()

		resBody, e := io.ReadAll(resp.Body)
		if e != nil {
			return nil, perr.InternalWithMessage("Failed reading response body.")
		}

		var input primitive.RoutedInputResponse
		e = json.Unmarshal(resBody, &input)
		if e != nil {
			return nil, perr.InternalWithMessage("Failed to deserialize the response body from input router.")
		}

		inputItems = append(inputItems, input)
	}

	return &primitive.RoutedInputListResponse{Items: inputItems, NextToken: nil}, nil
}
