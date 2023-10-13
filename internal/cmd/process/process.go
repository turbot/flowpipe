package process

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"

	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

func ProcessCmd(ctx context.Context) (*cobra.Command, error) {

	processCmd := &cobra.Command{
		Use:   "process",
		Short: "Process commands",
	}

	processGetCmd, err := ProcessGetCmd(ctx)
	if err != nil {
		return nil, err
	}
	processCmd.AddCommand(processGetCmd)

	processListCmd, err := ProcessListCmd(ctx)
	if err != nil {
		return nil, err
	}
	processCmd.AddCommand(processListCmd)

	return processCmd, nil

}

func ProcessGetCmd(ctx context.Context) (*cobra.Command, error) {

	var processGetCmd = &cobra.Command{
		Use:  "get <execution-id>",
		Args: cobra.ExactArgs(1),
		Run:  getProcessFunc(ctx),
	}

	processGetCmd.Flags().BoolP("output-only", "", false, "Get pipeline execution output only")

	return processGetCmd, nil
}

func ProcessListCmd(ctx context.Context) (*cobra.Command, error) {

	var processGetCmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listProcessFunc(ctx),
	}

	return processGetCmd, nil
}

func getProcessFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {

		apiClient := common.GetApiClient()

		outputOnly, _ := cmd.Flags().GetBool("output-only")

		if outputOnly {
			output, _, err := apiClient.ProcessApi.GetOutput(ctx, args[0]).Execute()
			if err != nil {
				error_helpers.ShowError(ctx, err)
				return
			}

			s, err := prettyjson.Marshal(output)

			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when calling `colorjson.Marshal`")
				return
			}

			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future
		} else {
			ex, _, err := apiClient.ProcessApi.Get(ctx, args[0]).Execute()
			if err != nil {
				error_helpers.ShowError(ctx, err)
				return
			}

			s, err := prettyjson.Marshal(ex)

			if err != nil {
				error_helpers.ShowErrorWithMessage(ctx, err, "Error when calling `colorjson.Marshal`")
				return
			}

			fmt.Println(string(s)) //nolint:forbidigo // console output, but we may change it to a different formatter in the future
		}
	}
}

func listProcessFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
		nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

		apiClient := common.GetApiClient()

		processes, _, err := apiClient.ProcessApi.List(ctx).Limit(limit).NextToken(nextToken).Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if processes != nil {
			printer := printers.GetPrinter(cmd)

			printableResource := types.PrintableProcess{}
			printableResource.Items, err = printableResource.Transform(processes)
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
