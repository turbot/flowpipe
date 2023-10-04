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
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
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

func PipelineRunCmd(ctx context.Context) (*cobra.Command, error) {

	var pipelineRunCmd = &cobra.Command{
		Use:  "run <pipeline-name>",
		Args: cobra.ArbitraryArgs,
		Run:  runPipelineFunc(ctx),
	}

	// Add the pipeline arg flag
	pipelineRunCmd.Flags().StringArray(constants.ArgPipelineArg, nil, "Specify the value of a pipeline argument. Multiple --pipeline-arg may be passed.")
	pipelineRunCmd.Flags().String(constants.ArgPipelineExecutionMode, "asynchronous", "Specify the pipeline execution mode. Supported values: asynchronous, synchronous.")

	return pipelineRunCmd, nil
}

func runPipelineFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		logger := fplog.Logger(ctx)

		// API client
		apiClient := common.GetApiClient()
		cmdPipelineRun := flowpipeapiclient.NewCmdPipeline("run")

		// Get the pipeline args from the flag
		pipelineArgs := map[string]string{}
		pipeLineArgValues, err := cmd.Flags().GetStringArray(constants.ArgPipelineArg)
		if err != nil {
			logger.Error("Error getting the value of pipeline-arg flag", "error", err)
			return
		}

		// validate the pipeline arg input
		err = validatePipelineArgs(pipeLineArgValues)
		if err != nil {
			logger.Error("Pipeline argument validation failed", "error", err)
			return
		}

		for _, value := range pipeLineArgValues {
			splitData := strings.Split(value, "=")
			pipelineArgs[splitData[0]] = splitData[1]
		}

		// Set the pipeline args
		cmdPipelineRun.ArgsString = &pipelineArgs

		// Get the pipeline execution mode from the flag
		executionMode := "asynchronous"
		pipelineExecutionMode, err := cmd.Flags().GetString(constants.ArgPipelineExecutionMode)
		if err != nil {
			logger.Error("Error getting the value of execution-mode flag", "error", err)
			return
		}

		if pipelineExecutionMode != "" {
			err = validatePipelineExecutionMode(pipelineExecutionMode)
			if err != nil {
				logger.Error("Pipeline execution mode validation failed", "error", err)
				return
			}
			executionMode = pipelineExecutionMode
		}

		// Set the pipeline execution mode
		cmdPipelineRun.ExecutionMode = &executionMode

		request := apiClient.PipelineApi.Cmd(ctx, args[0]).Request(*cmdPipelineRun)
		resp, _, err := request.Execute()
		if err != nil {
			logger.Error("Error when calling `PipelineApi.Cmd`", "error", err)
			return
		}

		if resp != nil {
			s, err := prettyjson.Marshal(resp)

			if err != nil {
				logger.Error("Error when calling `colorjson.Marshal`", "error", err)
				return
			}

			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future

		}
	}
}

func listPipelineFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		logger := fplog.Logger(ctx)
		limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
		nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

		apiClient := common.GetApiClient()
		resp, r, err := apiClient.PipelineApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
		if err != nil {
			logger.Error("Error when calling `PipelineApi.List`", "error", err, "httpResponse", r)
			return
		}

		if resp != nil {
			printer := printers.GetPrinter(cmd)

			printableResource := types.PrintablePipeline{}
			printableResource.Items, err = printableResource.Transform(resp)
			if err != nil {
				logger.Error("Error when transforming", "error", err)
			}

			err := printer.PrintResource(ctx, printableResource, cmd.OutOrStdout())
			if err != nil {
				logger.Error("Error when printing", "error", err)
			}
		}
	}
}

func validatePipelineArgs(pipelineArgs []string) error {
	validFormat := regexp.MustCompile(`^[^=]+=[^=]+$`)
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
