//nolint:forbidigo // CLI console output
package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
)

func triggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger commands",
	}

	cmd.AddCommand(triggerListCmd())
	cmd.AddCommand(TriggerShowCmd())

	return cmd
}

func triggerListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listTriggerFunc,
	}

	return cmd
}

func TriggerShowCmd() *cobra.Command {
	var triggerShowCmd = &cobra.Command{
		Use:  "show <trigger-name>",
		Args: cobra.ExactArgs(1),
		Run:  showTriggerFunc,
	}

	return triggerShowCmd
}

func listTriggerFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
	nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

	apiClient := common.GetApiClient()
	resp, _, err := apiClient.TriggerApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		printer := printers.GetPrinter(cmd)

		printableResource := types.PrintableTrigger{}
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

func showTriggerFunc(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	apiClient := common.GetApiClient()
	resp, _, err := apiClient.TriggerApi.Get(context.Background(), args[0]).Execute()
	if err != nil {
		error_helpers.ShowError(ctx, err)
		return
	}

	if resp != nil {
		output := ""
		if resp.Title != nil {
			output += "Title:    " + *resp.Title + "\n"
		}

		output += "Name:     " + *resp.Name
		output += "\nPipeline: " + *resp.Pipeline
		output += "\nType:     " + *resp.Type
		if resp.Url != nil {
			output += "\nUrl:      " + *resp.Url
		}

		if resp.Tags != nil {
			output += "\nTags:   "
			isFirstTag := true
			for k, v := range *resp.Tags {
				if isFirstTag {
					output += "  " + k + " = " + v
					isFirstTag = false
				} else {
					output += ", " + k + " = " + v
				}
			}
		}
		if resp.Description != nil {
			output += "\n\nDescription:\n" + *resp.Description
		}
		//nolint:forbidigo // CLI console output
		fmt.Println(output)
	}
}
