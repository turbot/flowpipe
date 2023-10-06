package process

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/internal/cmd/common"
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
