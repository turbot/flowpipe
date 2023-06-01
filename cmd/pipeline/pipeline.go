package pipeline

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
	openapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/printers"
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
		limit := int32(25) // int32 | The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25. (optional) (default to 25)
		nextToken := ""    // string | When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data. (optional)

		configuration := openapiclient.NewConfiguration()

		c := config.Config(ctx)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.Viper.GetBool("api.tls_insecure")}, //nolint:gosec // user defined
		}

		configuration.Servers[0].URL = c.Viper.GetString("api.host") + ":" + strconv.Itoa(c.Viper.GetInt("api.port")) + "/api/v0"
		configuration.HTTPClient = &http.Client{Transport: tr}

		apiClient := openapiclient.NewAPIClient(configuration)
		resp, r, err := apiClient.PipelineApi.List(context.Background()).Limit(limit).NextToken(nextToken).Execute()
		if err != nil {
			fplog.Logger(ctx).Error("Error when calling `PipelineApi.List`", "error", err, "httpResponse", r)
		}

		fplog.Logger(ctx).Info("Response from `PipelineApi.List`", "response", resp)

		if resp != nil {
			jsonPrinter := printers.JsonPrinter{}
			jsonPrinter.PrintObj(ctx, resp, cmd.OutOrStdout())
		}
	}
}
