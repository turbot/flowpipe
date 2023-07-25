package primitive

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

const (
	// HTTPRequestDefaultTimeoutMs is the default timeout for HTTP requests
	// For now the value is hardcoded to 3000 milliseconds
	// TODO: Make this configurable
	HTTPRequestDefaultTimeoutMs = 3000
)

type HTTPRequest struct {
	Input types.Input
}

type HTTPInput struct {
	URL              string
	Method           string
	RequestBody      string
	RequestHeaders   map[string]interface{}
	RequestTimeoutMs int
	CaCertPem        string
	Insecure         bool
}

func (h *HTTPRequest) ValidateInput(ctx context.Context, i types.Input) error {
	if i[schema.AttributeTypeUrl] == nil {
		return fperr.BadRequestWithMessage("HTTPRequest input must define a url")
	}
	u := i[schema.AttributeTypeUrl].(string)
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return fperr.BadRequestWithMessage("invalid url: " + u)
	}

	requestBody := i[schema.AttributeTypeRequestBody]
	if requestBody != nil {
		// Try to unmarshal the request body into JSON
		var requestBodyJSON map[string]interface{}
		unmarshalErr := json.Unmarshal([]byte(requestBody.(string)), &requestBodyJSON)
		if unmarshalErr != nil {
			// If unmarshaling fails, assume it's a plain string
			requestBodyJSON = nil
		}

		// If the request body is a JSON object
		if requestBodyJSON != nil {
			_, marshalErr := json.Marshal(requestBodyJSON)
			if marshalErr != nil {
				return fperr.BadRequestWithMessage("error marshaling request body JSON: " + marshalErr.Error())
			}
		}
	}
	return nil
}

func (h *HTTPRequest) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	// Validate the inputs
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO
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
func doRequest(ctx context.Context, inputParams *HTTPInput) (*types.StepOutput, error) {
	// Create the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest(strings.ToUpper(inputParams.Method), inputParams.URL, bytes.NewBuffer([]byte(inputParams.RequestBody)))
	if err != nil {
		return nil, fperr.BadRequestWithMessage("Error creating request: " + err.Error())
	}

	// Set the request headers
	for k, v := range inputParams.RequestHeaders {
		req.Header.Set(k, v.(string))
	}

	if inputParams.CaCertPem != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(inputParams.CaCertPem))

		client.Transport = &http.Transport{
			// #nosec G402
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: inputParams.Insecure,
			},
		}
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
	output := types.StepOutput{
		OutputVariables: map[string]interface{}{},
	}
	output.OutputVariables[schema.AttributeTypeStatus] = resp.Status
	output.OutputVariables[schema.AttributeTypeStatusCode] = resp.StatusCode
	output.OutputVariables[schema.AttributeTypeResponseHeaders] = headers
	output.OutputVariables[schema.AttributeTypeStartedAt] = start
	output.OutputVariables[schema.AttributeTypeFinishedAt] = finish

	if body != nil {
		output.OutputVariables[schema.AttributeTypeResponseBody] = string(body)
	}

	if resp.StatusCode >= 400 {
		message := resp.Status
		output.Errors = &types.StepErrors{
			types.StepError{
				Message:   message,
				ErrorCode: resp.StatusCode,
			},
		}
	}

	return &output, nil
}

// buildHTTPInput builds the HTTPInput struct from the input parameters
func buildHTTPInput(input types.Input) (*HTTPInput, error) {
	// Check for method
	method, ok := input[schema.AttributeTypeMethod].(string)
	if !ok {
		// If not provided, default to GET
		method = types.HttpMethodGet
	}

	// Method should be case insensitive
	method = strings.ToLower(method)

	// Build the input parameters
	inputParams := &HTTPInput{
		URL:    input["url"].(string),
		Method: method,

		// TODO: Make it configurable
		RequestTimeoutMs: HTTPRequestDefaultTimeoutMs,
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

	// Get the request body
	requestBody := input[schema.AttributeTypeRequestBody]
	if requestBody != nil {
		// Try to unmarshal the request body into JSON
		var requestBodyJSON map[string]interface{}
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
		if requestBodyJSON != nil {
			// Set the JSON encoding of the request body
			requestBodyJSONBytes, _ := json.Marshal(requestBodyJSON)
			inputParams.RequestBody = string(requestBodyJSONBytes)

			// Also, set the content type header to application/json
			requestHeaders["Content-Type"] = "application/json"
		}
	}
	inputParams.RequestHeaders = requestHeaders

	return inputParams, nil
}

// mapResponseHeaders maps the response headers to a simpler key-value pair
func mapResponseHeaders(resp *http.Response) map[string]interface{} {
	headers := map[string]interface{}{}
	// But, well known multi-value fields (e.g. Set-Cookie) should be maintained
	// in array form
	headersAsArrays := map[string]bool{"Set-Cookie": true}

	for k, v := range resp.Header {
		if headersAsArrays[k] {
			// It's a known multi-value header
			headers[k] = v
		} else {
			// Otherwise, just use the first value for simplicity
			headers[k] = v[0]
		}
	}
	return headers
}
