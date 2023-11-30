package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/color"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
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
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listPipelineFunc,
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
		printer := printers.GetPrinter(cmd)

		printableResource := types.PrintablePipeline{}
		printableResource.Items, err = printableResource.Transform(resp)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when transforming")
		}

		err := printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
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
		Use:  "show <pipeline-name>",
		Args: cobra.ExactArgs(1),
		Run:  showPipelineFunc,
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
		output := ""
		if resp.Title != nil {
			output += "Title: " + *resp.Title
		}
		if resp.Title != nil {
			output += "\nName:  " + resp.Name
		} else {
			output += "Name: " + resp.Name
		}
		if resp.Tags != nil {
			if resp.Title != nil {
				output += "\nTags:  "
			} else {
				output += "\nTags: "
			}
			isFirstTag := true
			for k, v := range resp.Tags {
				if isFirstTag {
					output += k + " = " + v
					isFirstTag = false
				} else {
					output += ", " + k + " = " + v
				}
			}
		}
		if resp.Description != nil {
			output += "\n\nDescription:\n" + *resp.Description + "\n"
		}
		if resp.Params != nil {
			output += formatSection("\nParams:", resp.Params)
		}
		if resp.OutputConfig != nil {
			output += formatSection("\nOutputs:", resp.OutputConfig)
		}
		output += "\nUsage:" + "\n"
		if resp.Params != nil {
			var pArg string

			// show the minimal required pipeline args
			for _, param := range resp.Params {
				if param.Default != nil || (param.Optional != nil && *param.Optional) {
					continue
				}
				pArg += " --pipeline-arg " + param.Name + "=<value>"
			}
			output += "  flowpipe pipeline run " + resp.Name + pArg
		} else {
			output += "  flowpipe pipeline run " + resp.Name
		}
		//nolint:forbidigo // CLI console output
		fmt.Println(output)
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
		Use:  "run <pipeline-name>",
		Args: cobra.ExactArgs(1),
		Run:  runPipelineFunc,
	}

	// Add the pipeline arg flag
	cmdconfig.OnCmd(cmd).
		AddStringArrayFlag(constants.ArgArg, nil, "Specify the value of a pipeline argument. Multiple --pipeline-arg may be passed.")

	return cmd
}

// func used to poll event logs
type pollEventLogFunc func(ctx context.Context, exId, plId string, last int) (bool, int, types.ProcessEventLogs, error)

func runPipelineFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	var resp map[string]any
	var err error

	var pollLogFunc pollEventLogFunc

	// if a host is set, use it to connect to API server
	if viper.IsSet(constants.ArgHost) {
		// run pipeline on server
		resp, err = runPipelineRemote(cmd, args)
		pollLogFunc = pollServerEventLog
	} else {
		// run pipeline in-process
		var m *manager.Manager
		resp, m, err = runPipelineLocal(cmd, args)
		// ensure to shut the manager when we are done
		defer func() {
			if m != nil {
				_ = m.Stop()
			}
		}()

		pollLogFunc = pollLocalEventLog

	}
	if err != nil {
		error_helpers.ShowErrorWithMessage(ctx, err, "Error executing pipeline")
	}

	displayStreamingLogs(ctx, cmd, resp, pollLogFunc)
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

	resp, _, err := apiClient.PipelineApi.Cmd(ctx, pipelineName).Request(*cmdPipelineRun).Execute()

	return resp, err
}

func runPipelineLocal(cmd *cobra.Command, args []string) (map[string]any, *manager.Manager, error) {
	ctx := cmd.Context()

	// create and start the manager with ES service, and Docker, but no API server
	m, err := manager.NewManager(ctx, manager.WithESService(), manager.WithDocker()).Start()
	error_helpers.FailOnError(err)

	// construct the pipeline name _after_ initializing so the cache is initialized
	pipelineName := api.ConstructPipelineFullyQualifiedName(args[0])

	//extract the pipeline args from the flags
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

func displayStreamingLogs(ctx context.Context, cmd *cobra.Command, resp map[string]any, pollLogFunc pollEventLogFunc) {
	if resp != nil && resp["flowpipe"] != nil {
		contents := resp["flowpipe"].(map[string]any)
		executionId := ""
		pipelineId := ""
		stale := false
		lastLoaded := ""

		if s, ok := contents["execution_id"].(string); !ok {
			error_helpers.ShowError(ctx, fmt.Errorf("Error obtaining execution_id"))
			return
		} else {
			executionId = s
		}
		if s, ok := contents["pipeline_execution_id"].(string); !ok {
			error_helpers.ShowError(ctx, fmt.Errorf("Error obtaining pipeline_execution_id"))
			return
		} else {
			pipelineId = s
		}
		if contents["is_stale"] != nil {
			stale = true
			lastLoaded = contents["last_loaded"].(string)
		}

		lastIndex := -1
		// printer := printers.GetPrinter(cmd) // TODO: Use once we can utilise multiple printers with StringPrinter default
		cg, err := color.NewDynamicColorGenerator(0, 16)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error creating ColorGenerator")
			return
		}
		printer := printers.StringPrinter{}
		printableResource := types.PrintableParsedEvent{}
		printableResource.Registry = make(map[string]types.ParsedEventRegistryItem)
		printableResource.ColorGenerator = cg

		// print execution_id / stale info
		var header []any
		header = append(header, types.ParsedHeader{
			ExecutionId: executionId,
			IsStale:     stale,
			LastLoaded:  lastLoaded,
		})
		printableResource.Items = header
		err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error writing execution header")
			return
		}

		// TODO: should we time out?
		// poll logs & print
		for {

			exit, i, logs, err := pollEventLog(ctx, executionId, pipelineId, lastIndex, pollLogFunc)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error polling event logs")
				return
			}

			printableResource.Items, err = printableResource.Transform(logs)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error parsing logs")
				return
			}

			err = printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error printing logs")
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

// Helper function to format a section
func formatSection(sectionName string, items interface{}) string {
	var output string
	if items != nil {
		output += sectionName + "\n"
		switch v := items.(type) {
		case []flowpipeapiclient.FpPipelineParam:
			for _, item := range v {
				output += "  " + paramToString(item) + "\n"
			}
		case []flowpipeapiclient.ModconfigPipelineOutput:
			for _, item := range v {
				output += "  " + outputToString(item) + "\n"
			}
		}
	}
	return output
}

// Helper function to convert Param to string
func paramToString(param flowpipeapiclient.FpPipelineParam) string {
	var strOutput string
	if param.Optional != nil && *param.Optional {
		strOutput = *param.Name + "[" + *param.Type + ",Optional]"
	} else {
		strOutput = *param.Name + "[" + *param.Type + "]"
	}

	if param.Description != nil && len(*param.Description) > 0 {
		strOutput += ": " + *param.Description
	}
	return strOutput
}

// Helper function to convert Output to string
func outputToString(output flowpipeapiclient.ModconfigPipelineOutput) string {
	strOutput := *output.Name
	if output.Description != nil && len(*output.Description) > 0 {
		strOutput += ": " + *output.Description
	}
	return strOutput
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
	logger := fplog.Logger(ctx)
	ex, err := execution.NewExecution(ctx, execution.WithID(exId))
	if err != nil {
		return true, 0, nil, err
	}

	eventStoreFilePath := filepaths.EventStorePath(ex.ID)

	logger.Debug("Opening file", "event store file", eventStoreFilePath)
	file, err := os.Open(eventStoreFilePath)
	if err != nil {
		// TODO KAI use perr? wrap?
		return true, 0, nil, err
	}

	logger.Debug("File opened", "event store file", eventStoreFilePath)
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
	// 	logger.Info("Returning here because of error", "error", err)
	// 	return complete, int(lastSize), res, nil
	// }

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize*40), bufio.MaxScanTokenSize*40)
	currentIndex := 0
	for scanner.Scan() {
		if currentIndex <= last {
			currentIndex++
			continue
		}

		line := scanner.Bytes()
		var entry types.EventLogEntry

		if err := json.Unmarshal(line, &entry); err != nil {
			logger.Warn("Error loading event entry", "error", err)
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
		logger.Error("Unable to scan event store file", "error", err)
		return complete, last, res, nil
	}

	return complete, currentIndex, res, nil
}
