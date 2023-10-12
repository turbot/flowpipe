package trigger

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/printers"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
)

func TriggerCmd(ctx context.Context) (*cobra.Command, error) {

	triggerCmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger commands",
	}

	triggerListCmd, err := TriggerListCmd(ctx)
	if err != nil {
		return nil, err
	}
	triggerCmd.AddCommand(triggerListCmd)

	triggerShowCmd, err := TriggerShowCmd(ctx)
	if err != nil {
		return nil, err
	}
	triggerCmd.AddCommand(triggerShowCmd)

	return triggerCmd, nil

}

func TriggerListCmd(ctx context.Context) (*cobra.Command, error) {

	var triggerListCmd = &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run:  listTriggerFunc(ctx),
	}

	return triggerListCmd, nil
}

func TriggerShowCmd(ctx context.Context) (*cobra.Command, error) {

	var triggerShowCmd = &cobra.Command{
		Use:  "show <trigger-name>",
		Args: cobra.ExactArgs(1),
		Run:  showTriggerFunc(ctx),
	}

	return triggerShowCmd, nil
}

func listTriggerFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
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
}

func showTriggerFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		apiClient := common.GetApiClient()
		resp, _, err := apiClient.TriggerApi.Get(context.Background(), args[0]).Execute()
		if err != nil {
			error_helpers.ShowError(ctx, err)
			return
		}

		if resp != nil {
			output := "\n"

			if resp.Title != nil {
				output += "Title:    " + *resp.Title + "\n"
			}
			output += "Name:     " + *resp.Name + "\n"
			output += "Pipeline: " + *resp.Pipeline + "\n"
			output += "Type:     " + *resp.Type + "\n"
			if resp.Tags != nil {
				output += "Tags:   "
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
			fmt.Println(output)
		}
	}
}
