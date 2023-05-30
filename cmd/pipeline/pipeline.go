package pipeline

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func PipelineCmd(ctx context.Context) (*cobra.Command, error) {

	pipelineCmd := &cobra.Command{
		Use:   "pipeline",
		Short: "pipeline commands",
	}

	pipelineListCmd, err := PipelineListCmd(ctx)
	if err != nil {
		return nil, err
	}
	pipelineCmd.AddCommand(pipelineListCmd)

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

func listPipelineFunc(ctx context.Context) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		fmt.Println("pipeline list")
	}
}
