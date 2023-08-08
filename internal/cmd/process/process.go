package process

import (
	"context"
	"fmt"

	"github.com/turbot/flowpipe/internal/cmd/common"
	"github.com/turbot/flowpipe/internal/fplog"

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

		logger := fplog.Logger(ctx)
		apiClient := common.GetApiClient()

		outputOnly, _ := cmd.Flags().GetBool("output-only")

		if outputOnly {
			output, _, err := apiClient.ProcessApi.GetOutput(ctx, args[0]).Execute()
			if err != nil {
				logger.Error("Error when calling `ProcessApi.GetOutput`", "error", err)
				return
			}

			s, err := prettyjson.Marshal(output)

			if err != nil {
				logger.Error("Error when calling `colorjson.Marshal`", "error", err)
				return
			}

			fmt.Println(string(s))
		} else {
			ex, _, err := apiClient.ProcessApi.Get(ctx, args[0]).Execute()
			if err != nil {
				logger.Error("Error when calling `ProcessApi.Get`", "error", err)
				return
			}

			s, err := prettyjson.Marshal(ex)

			if err != nil {
				logger.Error("Error when calling `colorjson.Marshal`", "error", err)
				return
			}

			fmt.Println(string(s))
		}
	}
}
