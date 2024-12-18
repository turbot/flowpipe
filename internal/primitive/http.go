package primitive

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type HTTPRequest struct {
	Input resources.Input
}

type HTTPInput struct {
	URL            string
	Method         string
	RequestBody    string
	RequestHeaders map[string]interface{}
	CaCertPem      string
	Insecure       bool
	Timeout        time.Duration
}

func (h *HTTPRequest) ValidateInput(ctx context.Context, i resources.Input) error {
	if i[schema.AttributeTypeUrl] == nil {
		return perr.BadRequestWithMessage("HTTPRequest input must define a url")
	}
	u := i[schema.AttributeTypeUrl].(string)
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return perr.BadRequestWithMessage("invalid url: " + u)
	}

	requestBody := i[schema.AttributeTypeRequestBody]
	if requestBody != nil && i[schema.AttributeTypeRequestHeaders] != nil {

		headers, ok := i[schema.AttributeTypeRequestHeaders].(map[string]interface{})
		if !ok {
			return perr.BadRequestWithMessage("request headers must be a map")
		}

		if headers["Content-Type"] != nil && strings.Contains(headers["Content-Type"].(string), "application/json") {
			// Try to unmarshal the request body into JSON
			var requestBodyJSON interface{}
			unmarshalErr := json.Unmarshal([]byte(requestBody.(string)), &requestBodyJSON)
			if unmarshalErr != nil {
				stepName := i[schema.AttributeTypeStepName].(string)
				return perr.BadRequestWithMessage("step " + stepName + " error marshaling request body JSON: " + unmarshalErr.Error())
			}
		}
	}

	if i[schema.AttributeTypeRequestHeaders] != nil {
		requestHeaders := i[schema.AttributeTypeRequestHeaders].(map[string]interface{})
		if requestHeaders["Authorization"] != nil && i[schema.BlockTypePipelineBasicAuth] != nil {
			stepName := i[schema.AttributeTypeStepName].(string)
			return perr.BadRequestWithMessage("step " + stepName + " should have either basic_auth or authorization header but not both")
		}
	}

	if i[schema.AttributeTypeTimeout] != nil {
		switch duration := i[schema.AttributeTypeTimeout].(type) {
		case string:
			_, err := time.ParseDuration(duration)
			if err != nil {
				return perr.BadRequestWithMessage("invalid sleep duration " + duration)
			}
		case int64:
			if duration < 0 {
				return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a positive whole number")
			}
		case float64:
			if duration < 0 {
				return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a positive whole number")
			}
		default:
			return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a string or a whole number")
		}
	}

	return nil
}

func (h *HTTPRequest) Run(ctx context.Context, input resources.Input) (*resources.Output, error) {
	// Validate the inputs
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO:
	// * Test SSL vs non-SSL
	// * Compare to features in https://www.tines.com/docs/actions/types/http-request#configuration-options

	// Constuct the input structure from the input parameters
	httpInput, err := buildHTTPInput(input)
	if err != nil {
		return nil, err
	}

	// Make the HTTP request
	output, err := doRequest(ctx, httpInput)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// doRequest performs the HTTP request based on the inputs provided and returns the output
func doRequest(ctx context.Context, inputParams *HTTPInput) (*resources.Output, error) {
	// Create the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest(strings.ToUpper(inputParams.Method), inputParams.URL, bytes.NewBuffer([]byte(inputParams.RequestBody)))
	if err != nil {
		return nil, perr.BadRequestWithMessage("Error creating request: " + err.Error())
	}

	// Set the request headers
	for k, v := range inputParams.RequestHeaders {
		req.Header.Set(k, v.(string))
	}

	// Initialize the TLSClientConfig with default settings
	tlsConfig := &tls.Config{} // #nosec G402

	// By default the client verifies the server's certificate chain and host name.
	// If the insecure flag is set, the client skips this verification and accepts any certificate presented by the server and any host name in that certificate.
	// Default value of insecure flag is false.
	if inputParams.Insecure {
		tlsConfig.InsecureSkipVerify = inputParams.Insecure
	}

	if inputParams.Timeout > 0 {
		client.Timeout = inputParams.Timeout
	}

	// If the input parameter 'ca_cert_pem' is set, the client verifies the server's certificate chain using the provided PEM encoded CA certificates.
	if inputParams.CaCertPem != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(inputParams.CaCertPem))

		tlsConfig.RootCAs = caCertPool
	}

	// Configure the client's transport with the final TLSClientConfig
	client.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	start := time.Now().UTC()
	resp, err := client.Do(req)
	finish := time.Now().UTC()
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Golang Response.Header is a map[string][]string, which is accurate
	// but complicated for users. We map it to a simpler key-value pair
	// approach.
	headers := mapResponseHeaders(resp)

	// Construct the output
	output := resources.Output{
		Data: map[string]interface{}{},
	}
	output.Data[schema.AttributeTypeStatus] = resp.Status
	output.Data[schema.AttributeTypeStatusCode] = resp.StatusCode
	output.Data[schema.AttributeTypeResponseHeaders] = headers

	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	var bodyString string

	if body != nil {
		bodyString = string(body)
	}

	if headers["Content-Type"] != nil && strings.Contains(headers["Content-Type"].(string), "application/json") && resp.StatusCode < 400 && len(body) > 0 {
		var bodyJSON interface{}
		err := json.Unmarshal(body, &bodyJSON)
		if err != nil {
			slog.Debug("Unable to marshallresponse body to JSON", "error", err)
			output.Data[schema.AttributeTypeResponseBody] = bodyString
		} else {
			output.Data[schema.AttributeTypeResponseBody] = bodyJSON
		}
	} else {
		output.Data[schema.AttributeTypeResponseBody] = bodyString
	}

	if resp.StatusCode >= 400 {
		output.Errors = []resources.StepError{
			{
				Error: perr.FromHttpError(errors.New(resp.Status), resp.StatusCode),
			},
		}
	}

	return &output, nil
}

// buildHTTPInput builds the HTTPInput struct from the input parameters
func buildHTTPInput(input resources.Input) (*HTTPInput, error) {
	// Check for method
	method, ok := input[schema.AttributeTypeMethod].(string)
	if !ok {
		// If not provided, default to GET
		method = resources.HttpMethodGet
	}

	// Method should be case insensitive
	method = strings.ToLower(method)

	// Build the input parameters
	inputParams := &HTTPInput{
		URL:    input["url"].(string),
		Method: method,
	}

	// Set the certificate, if provided
	if input[schema.AttributeTypeCaCertPem] != nil {
		inputParams.CaCertPem = input[schema.AttributeTypeCaCertPem].(string)
	}

	// Set value for insecureSkipVerify, if provided
	if input[schema.AttributeTypeInsecure] != nil {
		inputParams.Insecure = input[schema.AttributeTypeInsecure].(bool)
	}

	// Set the request headers, if provided
	requestHeaders := map[string]interface{}{}
	if input[schema.AttributeTypeRequestHeaders] != nil {
		requestHeaders = input[schema.AttributeTypeRequestHeaders].(map[string]interface{})
	}

	// check for basic_auth
	if input[schema.BlockTypePipelineBasicAuth] != nil {
		basicAuth := input[schema.BlockTypePipelineBasicAuth].(map[string]interface{})
		encodeString := base64.StdEncoding.EncodeToString([]byte(basicAuth["Username"].(string) + ":" + basicAuth["Password"].(string)))
		requestHeaders["Authorization"] = "Basic " + encodeString
	}

	contentType := ""

	if requestHeaders["Content-Type"] != nil {
		contentType = requestHeaders["Content-Type"].(string)
	}

	requestBody := input[schema.AttributeTypeRequestBody]

	if strings.Contains(contentType, "application/json") && requestBody != nil {
		// Get the request body

		if requestBody != nil {
			// Try to unmarshal the request body into JSON
			var requestBodyJSON interface{}
			unmarshalErr := json.Unmarshal([]byte(requestBody.(string)), &requestBodyJSON)
			if unmarshalErr != nil {
				// If unmarshaling fails, assume it's a plain string
				requestBodyJSON = nil

				// Set the request body as a string
				inputParams.RequestBody = requestBody.(string)

				// Also, set the content type header to plain text
				requestHeaders["Content-Type"] = "text/plain"
			}

			// If the request body is a JSON object
			if helpers.IsNil(requestBodyJSON) {
				// Set the JSON encoding of the request body
				requestBodyJSONBytes, _ := json.Marshal(requestBodyJSON)
				inputParams.RequestBody = string(requestBodyJSONBytes)

				// Also, set the content type header to application/json
				requestHeaders["Content-Type"] = "application/json"
			} else {
				inputParams.RequestBody = requestBody.(string)
			}
		}
	} else if requestBody != nil {
		inputParams.RequestBody = requestBody.(string)
	}

	inputParams.RequestHeaders = requestHeaders

	if input[schema.AttributeTypeTimeout] != nil {
		var timeout time.Duration
		switch timeoutDuration := input[schema.AttributeTypeTimeout].(type) {
		case string:
			timeout, _ = time.ParseDuration(timeoutDuration)
		case int64:
			timeout = time.Duration(timeoutDuration) * time.Millisecond // in milliseconds
		case float64:
			timeout = time.Duration(timeoutDuration) * time.Millisecond // in milliseconds
		}
		inputParams.Timeout = timeout
	}

	return inputParams, nil
}

// mapResponseHeaders maps the response headers to a simpler key-value pair
func mapResponseHeaders(resp *http.Response) map[string]interface{} {
	headers := map[string]interface{}{}
	for k, v := range resp.Header {
		if len(v) > 1 {
			headers[k] = strings.Join(v, ",")
		} else {
			headers[k] = v[0]
		}
	}
	return headers
}
