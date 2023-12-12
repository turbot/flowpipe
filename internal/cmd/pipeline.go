package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/sanitize"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	fpconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/manager"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/perr"
)

// pipeline commands
func pipelineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline commands",
	}

	cmd.AddCommand(pipelineListCmd())
	cmd.AddCommand(pipelineShowCmd())
	cmd.AddCommand(pipelineRunCmd())

	return cmd
}

// list
func pipelineListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Run:   listPipelineFunc,
		Short: "List pipelines from the current mod and its direct dependents",
		Long:  `List pipelines from the current mod and its direct dependents.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func listPipelineFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.ListPipelineResponse
	var err error
	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = listPipelineRemote()
	} else {
		resp, err = listPipelineLocal(cmd, args)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpPipeline](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintablePipeline(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func listPipelineRemote() (*types.ListPipelineResponse, error) {
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.PipelineApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		return nil, err
	}
	// map the API data type into the internal data type
	return types.ListPipelineResponseFromAPIResponse(resp)
}

func listPipelineLocal(cmd *cobra.Command, args []string) (*types.ListPipelineResponse, error) {
	ctx := cmd.Context()
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	// now list the pipelines
	return api.ListPipelines()
}

// show
func pipelineShowCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show <pipeline-name>",
		Args:  cobra.ExactArgs(1),
		Run:   showPipelineFunc,
		Short: "Show details of a pipeline from the current mod or its direct dependents",
		Long:  `Show details of a pipeline from the current mod or its direct dependents.`,
	}
	// initialize hooks
	cmdconfig.OnCmd(cmd)

	return cmd
}

func showPipelineFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp *types.FpPipeline
	var err error
	pipelineName := args[0]

	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		resp, err = getPipelineRemote(pipelineName)
	} else {
		resp, err = getPipelineLocal(ctx, pipelineName)
	}

	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer, err := printers.GetPrinter[types.FpPipeline](cmd)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error obtaining printer")
		}
		printableResource := types.NewPrintablePipelineFromSingle(resp)
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when printing")
		}
	}
}

func getPipelineRemote(pipelineName string) (*types.FpPipeline, error) {
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.PipelineApi.Get(context.Background(), pipelineName).Execute()
	if err != nil {
		return nil, err
	}
	// map the API data type into the internal data type
	return types.FpPipelineFromAPIResponse(*resp)
}

func getPipelineLocal(ctx context.Context, pipelineName string) (*types.FpPipeline, error) {
	// create and start the manager in local mode (i.e. do not set listen address)
	m, err := manager.NewManager(ctx).Start()
	error_helpers.FailOnError(err)
	defer func() {
		// TODO ignore shutdown error?
		_ = m.Stop()
	}()

	// try to fetch the pipeline from the cache
	return api.GetPipeline(pipelineName)
}

// run
func pipelineRunCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run <pipeline-name>",
		Args:  cobra.ExactArgs(1),
		Run:   runPipelineFunc,
		Short: "Run a pipeline from the current mod or its direct dependents or from a Flowpipe server instance",
		Long:  `Run a pipeline from the current mod or its direct dependents or from a Flowpipe server instance.`,
	}

	// Add the pipeline arg flag
	cmdconfig.OnCmd(cmd).
		AddStringArrayFlag(constants.ArgArg, nil, "Specify the value of a pipeline argument. Multiple --arg may be passed.").
		AddBoolFlag(constants.ArgVerbose, false, "Enable verbose output.").
		AddBoolFlag(constants.ArgDetach, false, "Run the pipeline in detached mode.")

	return cmd
}

// func used to poll event logs
type pollEventLogFunc func(ctx context.Context, exId, plId string, last int) (bool, int, types.ProcessEventLogs, error)

func runPipelineFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp map[string]any
	var err error
	var pollLogFunc pollEventLogFunc

	isDetach := viper.GetBool(constants.ArgDetach)
	isRemote := viper.IsSet(constants.ArgHost)
	if !isRemote && isDetach {
		error_helpers.ShowError(ctx, fmt.Errorf("unable to use --detach with local execution"))
		return
	}

	// if a host is set, use it to connect to API server
	resp, pollLogFunc, err = executePipeline(cmd, args, isRemote)
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "failed executing pipeline")
		return
	}

	output := viper.GetString(constants.ArgOutput)
	streamLogs := output == "plain" || output == "pretty"
	switch {
	case isDetach:
		err := displayDetached(ctx, cmd, resp)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed printing execution information")
			return
		}
	case streamLogs:
		displayStreamingLogs(ctx, cmd, resp, pollLogFunc)
	default:
		displayBasicOutput(ctx, cmd, resp, pollLogFunc)
	}
}

func executePipeline(cmd *cobra.Command, args []string, isRemote bool) (map[string]any, pollEventLogFunc, error) {
	if isRemote {
		// run pipeline on server
		resp, err := runPipelineRemote(cmd, args)
		pollLogFunc := pollServerEventLog
		return resp, pollLogFunc, err
	}
	// run pipeline in-process
	var m *manager.Manager
	resp, m, err := runPipelineLocal(cmd, args)
	// ensure to shut the manager when we are done
	defer func() {
		if m != nil {
			_ = m.Stop()
		}
	}()

	pollLogFunc := pollLocalEventLog
	return resp, pollLogFunc, err
}

func runPipelineRemote(cmd *cobra.Command, args []string) (map[string]interface{}, error) {
	ctx := cmd.Context()

	pipelineName := args[0]
	// extract the pipeline args from the flags
	pipelineArgs := getPipelineArgs(cmd)

	// API client
	apiClient := common.GetApiClient()
	cmdPipelineRun := flowpipeapiclient.NewCmdPipeline("run")

	// Set the pipeline args
	cmdPipelineRun.ArgsString = &pipelineArgs

	resp, _, err := apiClient.PipelineApi.Command(ctx, pipelineName).Request(*cmdPipelineRun).Execute()

	return resp, err
}

func runPipelineLocal(cmd *cobra.Command, args []string) (map[string]any, *manager.Manager, error) {
	ctx := cmd.Context()

	// create and start the manager with ES service, and Docker, but no API server
	m, err := manager.NewManager(ctx, manager.WithESService(), manager.WithDocker()).Start()
	error_helpers.FailOnError(err)

	// construct the pipeline name _after_ initializing so the cache is initialized
	pipelineName := api.ConstructPipelineFullyQualifiedName(args[0])

	// extract the pipeline args from the flags
	pipelineArgs := getPipelineArgs(cmd)

	input := types.CmdPipeline{
		Command:    "run",
		ArgsString: pipelineArgs,
	}

	resp, _, err := api.ExecutePipeline(input, pipelineName, m.ESService)
	if err != nil {
		return nil, nil, err
	}

	return resp, m, err
}

func displayDetached(ctx context.Context, cmd *cobra.Command, resp map[string]any) error {
	exec, err := types.FpPipelineExecutionFromAPIResponse(resp)
	if err != nil {
		return err
	}
	err = displayPipelineExecution(ctx, exec, cmd)
	if err != nil {
		return err
	}
	return nil
}

func displayStreamingLogs(ctx context.Context, cmd *cobra.Command, resp map[string]any, pollLogFunc pollEventLogFunc) {
	if resp != nil && resp["flowpipe"] != nil {
		contents := resp["flowpipe"].(map[string]any)
		executionId := ""
		pipelineId := ""
		stale := false
		lastLoaded := ""

		if s, ok := contents["execution_id"].(string); !ok {
			error_helpers.ShowError(ctx, fmt.Errorf("failed obtaining execution_id"))
			return
		} else {
			executionId = s
		}
		if s, ok := contents["pipeline_execution_id"].(string); !ok {
			error_helpers.ShowError(ctx, fmt.Errorf("failed obtaining pipeline_execution_id"))
			return
		} else {
			pipelineId = s
		}
		if contents["is_stale"] != nil {
			stale = true
			lastLoaded = contents["last_loaded"].(string)
		}

		lastIndex := -1

		printer, err := printers.NewStringPrinter[types.SanitizedStringer]()
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed instantiating string printer")
			return
		}
		printer.Sanitizer = sanitize.Instance
		printableResource := types.NewPrintableParsedEvent(pipelineId)

		// print execution_id / stale info
		var header []types.SanitizedStringer
		header = append(header, types.ParsedHeader{
			ExecutionId: executionId,
			IsStale:     stale,
			LastLoaded:  lastLoaded,
		})
		printableResource.Items = header
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed writing execution header")
			return
		}

		// TODO: should we time out?
		// poll logs & print
		for {
			exit, i, logs, err := pollEventLog(ctx, executionId, pipelineId, lastIndex, pollLogFunc)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "failed polling events")
				return
			}

			err = printableResource.SetEvents(logs)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "failed parsing events")
				return
			}

			err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "failed printing events")
				return
			}

			lastIndex = i

			if exit {
				break
			}

			// TODO: make this configurable
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func displayBasicOutput(ctx context.Context, cmd *cobra.Command, resp map[string]any, pollLogFunc pollEventLogFunc) {
	exec, err := types.FpPipelineExecutionFromAPIResponse(resp)
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "failed obtaining execution")
		return
	}

	lastIndex := -1

	for {
		exit, i, logs, err := pollEventLog(ctx, exec.ExecutionId, exec.PipelineExecutionId, lastIndex, pollLogFunc)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "failed polling events")
			return
		}
		lastIndex = i

		for _, log := range logs {
			switch log.EventType {
			case event.HandlerPipelineQueued:
				var e event.PipelineQueued
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("failed unmarshalling %s event", e.HandlerName()))
					return
				}
				if e.PipelineExecutionID == exec.PipelineExecutionId {
					exec.PipelineName = &e.Name
					exec.CreatedAt = &e.Event.CreatedAt
					exec.Status = "queued"
				}
			case event.HandlerPipelineStarted:
				var e event.PipelineStarted
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("failed unmarshalling %s event", e.HandlerName()))
					return
				}
				if e.PipelineExecutionID == exec.PipelineExecutionId {
					exec.Status = "started"
				}
			case event.HandlerPipelineFinished:
				var e event.PipelineFinished
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("failed unmarshalling %s event", e.HandlerName()))
					return
				}
				if e.PipelineExecutionID == exec.PipelineExecutionId {
					exec.Status = "finished"
					exec.Outputs = e.PipelineOutput
				}
			case event.HandlerPipelineFailed:
				var e event.PipelineFailed
				err := json.Unmarshal([]byte(log.Payload), &e)
				if err != nil {
					error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("failed unmarshalling %s event", e.HandlerName()))
					return
				}
				if e.PipelineExecutionID == exec.PipelineExecutionId {
					exec.Status = "failed"
					exec.Outputs = e.PipelineOutput
					exec.Errors = e.Errors
				}
			default:
				// ignore other events
			}
		}

		if exit {
			break
		}

		// TODO: make this configurable
		time.Sleep(500 * time.Millisecond)
	}

	err = displayPipelineExecution(ctx, exec, cmd)
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}
}

func displayPipelineExecution(ctx context.Context, pe *types.FpPipelineExecution, cmd *cobra.Command) error {
	printer, err := printers.GetPrinter[types.FpPipelineExecution](cmd)
	if err != nil {
		return fmt.Errorf("error obtaining printer\n%v", err)
	}
	printableResource := types.PrintablePipelineExecution{
		Items: []types.FpPipelineExecution{*pe},
	}
	err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
	if err != nil {
		return fmt.Errorf("error when printing\n%v", err)
	}

	return nil
}

func getPipelineArgs(cmd *cobra.Command) map[string]string {
	pipelineArgs := map[string]string{}
	pipeLineArgValues, err := cmd.Flags().GetStringArray(constants.ArgArg)
	error_helpers.FailOnErrorWithMessage(err, "Error getting the value of pipeline-arg flag")

	// validate the pipeline arg input
	err = validatePipelineArgs(pipeLineArgValues)
	error_helpers.FailOnErrorWithMessage(err, "Pipeline argument validation failed")

	for _, value := range pipeLineArgValues {
		splitData := strings.SplitN(value, "=", 2)
		pipelineArgs[splitData[0]] = splitData[1]
	}
	return pipelineArgs
}

func validatePipelineArgs(pipelineArgs []string) error {
	validFormat := regexp.MustCompile(`^[\w-]+=[\S\s]+$`)
	for _, arg := range pipelineArgs {
		if !validFormat.MatchString(arg) {
			return fmt.Errorf("invalid format: %s", arg)
		}
	}
	return nil
}

func pollEventLog(ctx context.Context, executionId, rootPipelineId string, lastIndex int, pollFunc func(ctx context.Context, exId, plId string, last int) (bool, int, types.ProcessEventLogs, error)) (bool, int, types.ProcessEventLogs, error) {
	return pollFunc(ctx, executionId, rootPipelineId, lastIndex)
}

func pollServerEventLog(ctx context.Context, exId, plId string, last int) (bool, int, types.ProcessEventLogs, error) {
	complete := false
	var out types.ProcessEventLogs
	client := common.GetApiClient()
	logs, _, err := client.ProcessApi.GetLog(ctx, exId).Execute()
	if err != nil {
		return false, last, nil, err
	}

	// check we have new event logs to parse
	if len(logs.Items)-1 > last {
		for index, item := range logs.Items {
			if index > last {
				ts, err := time.Parse(time.RFC3339Nano, *item.Ts)
				if err != nil {
					return false, 0, nil, fmt.Errorf("error parsing timestamp from %s", *item.Ts)
				}
				out = append(out, types.ProcessEventLog{
					EventType: *item.EventType,
					Timestamp: &ts,
					Payload:   *item.Payload,
				})

				last = index

				// check to see if event logs complete
				if item.EventType != nil && (*item.EventType == event.HandlerPipelineFinished || *item.EventType == event.HandlerPipelineFailed) {
					payload := make(map[string]any)
					if err := json.Unmarshal([]byte(*item.Payload), &payload); err != nil {
						return false, 0, nil, fmt.Errorf("error parsing payload from %s", *item.Payload)
					}
					complete = payload["pipeline_execution_id"] != nil && payload["pipeline_execution_id"] == plId
				}
			}
		}
	}

	return complete, last, out, nil
}

func pollLocalEventLog(ctx context.Context, exId, plId string, last int) (bool, int, types.ProcessEventLogs, error) {
	ex, err := execution.NewExecution(ctx, execution.WithID(exId))
	if err != nil {
		return true, 0, nil, err
	}

	eventStoreFilePath := filepaths.EventStoreFilePath(ex.ID)

	slog.Debug("Opening file", "event store file", eventStoreFilePath)
	file, err := os.Open(eventStoreFilePath)
	if err != nil {
		// TODO KAI use perr? wrap?
		return true, 0, nil, err
	}

	slog.Debug("File opened", "event store file", eventStoreFilePath)
	defer func() {
		// ensure we close the file
		_ = file.Close()
	}()

	// var lastSize int64
	// if last != -1 {
	// 	lastSize = int64(last)
	// }

	var res types.ProcessEventLogs
	var complete bool

	// TODO: this code section is buggy, the json unmarshall errors out maybe 50% of the time
	// TODO: with unexpected error in JSON file
	// Seek to the last read position
	// if _, err := file.Seek(lastSize, 0); err != nil {
	// 	//nolint:nilerr // just return without passing error - we will try again next time
	// 	slog.Info("Returning here because of error", "error", err)
	// 	return complete, int(lastSize), res, nil
	// }

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, fpconstants.MaxScanSize), fpconstants.MaxScanSize)
	currentIndex := 0
	for scanner.Scan() {
		if currentIndex < last {
			currentIndex++
			continue
		}

		line := scanner.Bytes()
		var entry types.EventLogEntry

		if err := json.Unmarshal(line, &entry); err != nil {
			slog.Warn("Error loading event entry", "error", err)
			return complete, last, res, nil
		}

		res = append(res, types.ProcessEventLog{
			EventType: entry.EventType,
			Timestamp: entry.Timestamp,
			Payload:   string(entry.Payload),
		})

		// check to see if event logs complete
		if entry.EventType == event.HandlerPipelineFinished || entry.EventType == event.HandlerPipelineFailed {
			payload := make(map[string]any)
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				return false, 0, nil, perr.InternalWithMessage(fmt.Sprintf("error parsing payload from %s", entry.Payload))
			}
			complete = payload["pipeline_execution_id"] != nil && payload["pipeline_execution_id"] == plId
		}

		currentIndex++
	}

	if err := scanner.Err(); err != nil {
		if err.Error() == bufio.ErrTooLong.Error() {
			return false, 0, nil, perr.InternalWithMessageAndType(perr.ErrorCodeInternalTokenTooLarge, "Event log entry too large. Max size is "+strconv.Itoa(fpconstants.MaxScanSize))
		}

		slog.Error("Unable to scan event store file", "error", err)
		return complete, last, res, nil
	}

	return complete, currentIndex, res, nil
}
