package common

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strconv"

	"github.com/spf13/viper"

	flowpipeapiclient "github.com/turbot/flowpipe-sdk-go"
	"github.com/turbot/flowpipe/internal/constants"
)

type customTransport struct {
	Transport http.RoundTripper
}

var ErrUnreachable = errors.New("flowpipe service is unreachable. Is your flowpipe service running?")

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := c.Transport.RoundTrip(req)
	opErr, ok := err.(*net.OpError)
	if ok && opErr.Op == "dial" && opErr.Err.Error() == "connect: connection refused" {
		return nil, ErrUnreachable
	}

	return resp, err
}

func GetApiClient() *flowpipeapiclient.APIClient {
	configuration := flowpipeapiclient.NewConfiguration()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("api.tls_insecure")}, //nolint:gosec // user defined
	}

	// Use the custom transport
	customTransport := &customTransport{Transport: tr}

	configuration.Servers[0].URL = viper.GetString(constants.ArgApiHost) + ":" + strconv.Itoa(viper.GetInt(constants.ArgApiPort)) + "/api/v0"
	configuration.HTTPClient = &http.Client{Transport: customTransport}

	apiClient := flowpipeapiclient.NewAPIClient(configuration)

	return apiClient
}
