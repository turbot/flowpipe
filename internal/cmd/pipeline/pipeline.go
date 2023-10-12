package pipeline

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/go-kit/helpers"
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
	pipelineRunCmd.Flags().String(constants.ArgPipelineExecutionMode, "asynchronous", "Specify the pipeline execution mode. Supported values: asynchronous, synchronous.")

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

		request := apiClient.PipelineApi.Cmd(ctx, args[0]).Request(*cmdPipelineRun)
		resp, _, err := request.Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if resp != nil {
			s, err := prettyjson.Marshal(resp)

			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when calling `colorjson.Marshal`")
				return
			}

			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future

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
			output := "\n"

			if resp.Title != nil {
				output += "Title: " + *resp.Title + "\n"
			}
			output += "Name:  " + *resp.Name + "\n"
			if resp.Tags != nil {
				output += "Tags:"
				isFirstTag := true
				for k, v := range *resp.Tags {
					if isFirstTag {
						output += "  " + k + " = " + v
						isFirstTag = false
					} else {
						output += ", " + k + " = " + v
					}
				}
				output += "\n"
			}
			if resp.Description != nil {
				output += "\nDescription:\n" + *resp.Description + "\n"
			}
			output += formatSection("\nParams:", resp.Params)
			output += formatSection("\nOutputs:", resp.Outputs)
			output += "\nUsage:" + "\n"
			if resp.Params != nil {
				var pArg string
				for _, param := range *resp.Params {
					pArg += " --pipeline-arg " + *param.Name + "=<value>"
				}
				output += "  flowpipe pipeline run " + *resp.Name + pArg
			} else {
				output += "  flowpipe pipeline run " + *resp.Name
			}
			output += "\n"
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
		case *map[string]flowpipeapiclient.ModconfigPipelineParam:
			for _, item := range *v {
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
func paramToString(param flowpipeapiclient.ModconfigPipelineParam) string {
	return *param.Name + "[*param.Type]: " + *param.Description
}

// Helper function to convert Output to string
func outputToString(output flowpipeapiclient.ModconfigPipelineOutput) string {
	return *output.Name + ": " + *output.Description
}
func validatePipelineArgs(pipelineArgs []string) error {
	validFormat := regexp.MustCompile(`^[^=]+=.+$`)
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
