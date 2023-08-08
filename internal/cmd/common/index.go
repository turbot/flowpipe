package common

import (
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/spf13/viper"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
)

func GetApiClient() *flowpipeapiclient.APIClient {
	configuration := flowpipeapiclient.NewConfiguration()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("api.tls_insecure")}, //nolint:gosec // user defined
	}

	configuration.Servers[0].URL = viper.GetString("api.host") + ":" + strconv.Itoa(viper.GetInt("api.port")) + "/api/v0"
	configuration.HTTPClient = &http.Client{Transport: tr}

	apiClient := flowpipeapiclient.NewAPIClient(configuration)

	return apiClient
}
