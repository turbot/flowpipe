package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
)

func PipelineCmd(ctx context.Context) (*cobra.Command, error) {

	pipelineCmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline commands",
	}

	pipelineListCmd, err := PipelineListCmd(ctx)
	if err != nil {
		return nil, err
	}
	pipelineCmd.AddCommand(pipelineListCmd)

	pipelineShowCmd, err := PipelineShowCmd(ctx)
	if err != nil {
		return nil, err
	}
	pipelineCmd.AddCommand(pipelineShowCmd)

	pipelineRunCmd, err := PipelineRunCmd(ctx)
	if err != nil {
		return nil, err
	}
	pipelineCmd.AddCommand(pipelineRunCmd)

	return pipelineCmd, nil

}

func PipelineListCmd(ctx context.Context) (*cobra.Command, error) {

	var serviceStartCmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listPipelineFunc(ctx),
	}

	return serviceStartCmd, nil
}

func PipelineShowCmd(ctx context.Context) (*cobra.Command, error) {

	var serviceStartCmd = &cobra.Command{
		Use:  "show <pipeline-name>",
		Args: cobra.ExactArgs(1),
		Run:  showPipelineFunc(ctx),
	}

	return serviceStartCmd, nil
}

func PipelineRunCmd(ctx context.Context) (*cobra.Command, error) {

	var pipelineRunCmd = &cobra.Command{
		Use:  "run <pipeline-name>",
		Args: cobra.ExactArgs(1),
		Run:  runPipelineFunc(ctx),
	}

	// Add the pipeline arg flag
	pipelineRunCmd.Flags().StringArray(constants.ArgPipelineArg, nil, "Specify the value of a pipeline argument. Multiple --pipeline-arg may be passed.")
	pipelineRunCmd.Flags().String(constants.ArgPipelineExecutionMode, "synchronous", "Specify the pipeline execution mode. Supported values: asynchronous, synchronous.")
	pipelineRunCmd.Flags().Int(constants.ArgPipelineWaitTime, 60, "Specify how long the pipeline should wait (in seconds) when run in synchronous execution mode.")

	return pipelineRunCmd, nil
}

func runPipelineFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		// API client
		apiClient := common.GetApiClient()
		cmdPipelineRun := flowpipeapiclient.NewCmdPipeline("run")

		// Get the pipeline args from the flag
		pipelineArgs := map[string]string{}
		pipeLineArgValues, err := cmd.Flags().GetStringArray(constants.ArgPipelineArg)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error getting the value of pipeline-arg flag")
			return
		}

		// validate the pipeline arg input
		err = validatePipelineArgs(pipeLineArgValues)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Pipeline argument validation failed")
			return
		}

		for _, value := range pipeLineArgValues {
			splitData := strings.SplitN(value, "=", 2)
			pipelineArgs[splitData[0]] = splitData[1]
		}

		// Set the pipeline args
		cmdPipelineRun.ArgsString = &pipelineArgs

		// Get the pipeline execution mode from the flag
		executionMode := "asynchronous"
		pipelineExecutionMode, err := cmd.Flags().GetString(constants.ArgPipelineExecutionMode)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error getting the value of execution-mode flag")
			return
		}

		if pipelineExecutionMode != "" {
			err = validatePipelineExecutionMode(pipelineExecutionMode)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Pipeline execution mode validation failed")
				return
			}
			executionMode = pipelineExecutionMode
		}

		// Set the pipeline execution mode
		cmdPipelineRun.ExecutionMode = &executionMode

		// Get the pipeline wait time from the flag
		waitTime := int32(60)
		pipelineWaitTime, err := cmd.Flags().GetInt(constants.ArgPipelineWaitTime)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, fmt.Sprintf("Error getting the value of %s flag", constants.ArgPipelineWaitTime))
			return
		}
		if pipelineWaitTime != 0 {
			err = validatePipelineWaitTime(pipelineWaitTime)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Pipeline wait time validation failed")
				return
			}

			waitTime = int32(pipelineWaitTime)
		}

		// Set the pipeline wait time
		cmdPipelineRun.WaitRetry = &waitTime

		request := apiClient.PipelineApi.Cmd(ctx, args[0]).Request(*cmdPipelineRun)
		resp, _, err := request.Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if executionMode == "synchronous" {
			s, err := prettyjson.Marshal(resp)
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when calling `colorjson.Marshal`")
				return
			}
			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future
		} else {
			var executionId string
			var rootPipelineId string

			if resp != nil && resp["flowpipe"] != nil {
				contents := resp["flowpipe"].(map[string]interface{})
				executionId = contents["execution_id"].(string)
				rootPipelineId = contents["pipeline_execution_id"].(string)
			}

			err := pollEventLogAndRender(ctx, apiClient, executionId, rootPipelineId, cmd.OutOrStdout())
			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when polling event log")
				return
			}
		}
	}
}

func listPipelineFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
		nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

		apiClient := common.GetApiClient()
		resp, _, err := apiClient.PipelineApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
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
}

func showPipelineFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		apiClient := common.GetApiClient()
		resp, _, err := apiClient.PipelineApi.Get(context.Background(), args[0]).Execute()
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
				output += "\nName:  " + *resp.Name
			} else {
				output += "Name: " + *resp.Name
			}
			if resp.Tags != nil {
				if resp.Title != nil {
					output += "\nTags:  "
				} else {
					output += "\nTags: "
				}
				isFirstTag := true
				for k, v := range *resp.Tags {
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
			if resp.Outputs != nil {
				output += formatSection("\nOutputs:", resp.Outputs)
			}
			output += "\nUsage:" + "\n"
			if resp.Params != nil {
				var pArg string

				// show the minimal required pipeline args
				for _, param := range resp.Params {
					if (param.Default != nil && len(param.Default) > 0) || (param.Optional != nil && *param.Optional) {
						continue
					}
					pArg += " --pipeline-arg " + *param.Name + "=<value>"
				}
				output += "  flowpipe pipeline run " + *resp.Name + pArg
			} else {
				output += "  flowpipe pipeline run " + *resp.Name
			}
			//nolint:forbidigo // CLI console output
			fmt.Println(output)
		}
	}
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

func validatePipelineExecutionMode(mode string) error {
	if !helpers.StringSliceContains([]string{"asynchronous", "synchronous"}, mode) {
		return fmt.Errorf("invalid execution mode: %s", mode)
	}
	return nil
}

func validatePipelineWaitTime(wait int) error {
	// TODO: Verify if we want min/max validation and appropriate durations
	mn := 1
	mx := 3600
	if wait < mn || wait > mx {
		return fmt.Errorf("invalid wait time: %d - should be between %d and %d", wait, mn, mx)
	}
	return nil
}

func pollEventLogAndRender(ctx context.Context, client *flowpipeapiclient.APIClient, executionId, rootPipelineId string, w io.Writer) error {
	isComplete := false
	lastIndexRead := -1
	printer := printers.LogLinePrinter{}
	for {
		logs, _, err := client.ProcessApi.GetLog(ctx, executionId).Execute()
		if err != nil {
			return err
		}

		printableResource := types.PrintableLogLine{}
		printableResource.Items, err = printableResource.Transform(logs)
		if err != nil {
			error_helpers.ShowErrorWithMessage(ctx, err, "Error when transforming")
		}

		var render []types.LogLine
		for logIndex, logEntry := range printableResource.Items.([]types.LogLine) {
			if logIndex > lastIndexRead {
				lastIndexRead = logIndex
				render = append(render, logEntry)
			}
		}

		printableResource.Items = render
		err = printer.PrintResource(ctx, printableResource, w)
		if err != nil {

		}

		// Check logs received for termination/completion of execution
		for _, logEntry := range logs.Items {
			if *logEntry.EventType == event.HandlerPipelineFinished || *logEntry.EventType == event.HandlerPipelineFailed {
				if logEntry.Payload != nil {
					payload := make(map[string]any)
					if err := json.Unmarshal([]byte(*logEntry.Payload), &payload); err != nil {
						return err
					}

					if payload["pipeline_execution_id"] != nil && payload["pipeline_execution_id"] == rootPipelineId {
						isComplete = true
						break
					}
				}
			}
		}
		if isComplete {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
