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
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

const (
	HTTPRequestGet  = "get"
	HttpRequestPost = "post"

	// HTTPRequestDefaultTimeoutMs is the default timeout for HTTP requests
	// For now the value is hardcoded to 3000 milliseconds
	// TODO: Make this configurable
	HTTPRequestDefaultTimeoutMs = 3000
)

type HTTPRequest struct {
	Input types.Input
}

type HTTPPOSTInput struct {
	URL              string
	RequestBody      string
	RequestHeaders   map[string]interface{}
	RequestTimeoutMs int
	CaCertPem        string
	Insecure         bool
}

func (h *HTTPRequest) ValidateInput(ctx context.Context, i types.Input) error {
	if i["url"] == nil {
		return fperr.BadRequestWithMessage("HTTPRequest input must define a url")
	}
	u := i["url"].(string)
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return fperr.BadRequestWithMessage("invalid url: " + u)
	}

	requestBody := i["body"]
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
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO
	// * Currently the primitive only supports GET and POST requests. Add support for other methods.
	// * Test SSL vs non-SSL
	// * Compare to features in https://www.tines.com/docs/actions/types/http-request#configuration-options

	method := input["method"].(string)
	inputURL := input["url"].(string)

	var output *types.StepOutput
	var err error
	switch method {
	case HTTPRequestGet:
		output, err = get(ctx, inputURL)
	case HttpRequestPost:
		// build the input for the POST request
		postInput, inputErr := buildHTTPPostInput(input)
		if inputErr != nil {
			return nil, inputErr
		}

		output, err = post(ctx, postInput)
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func get(ctx context.Context, inputURL string) (*types.StepOutput, error) {
	logger := fplog.Logger(ctx)

	start := time.Now().UTC()
	resp, err := http.Get(inputURL)
	finish := time.Now().UTC()
	if err != nil {
		logger.Error("error making request", "error", err, "response", resp)
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

	output := types.StepOutput{
		"status":      resp.Status,
		"status_code": resp.StatusCode,
		"headers":     headers,
		"started_at":  start,
		"finished_at": finish,
	}

	if body != nil {
		output["body"] = string(body)
	}

	var bodyJSON interface{}

	// Process the response body only if the status code is 200
	if resp != nil && resp.StatusCode == http.StatusOK {
		// The unmarshalling is only done if the content type is JSON,
		// otherwise the unmashalling will fail.
		// Hence, the body_json field will only be populated if the content type is JSON.
		if resp.Header.Get("Content-Type") == "application/json" {
			err = json.Unmarshal(body, &bodyJSON)
			if err != nil {
				logger.Error("error unmarshalling body: %s", err)
				return nil, err
			}
			output["body_json"] = bodyJSON
		}
	}

	return &output, nil
}

func post(ctx context.Context, inputParams *HTTPPOSTInput) (*types.StepOutput, error) {
	logger := fplog.Logger(ctx)

	// Create the HTTP client
	client := &http.Client{}
	req, err := http.NewRequest("POST", inputParams.URL, bytes.NewBuffer([]byte(inputParams.RequestBody)))
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

	output := types.StepOutput{
		"status":      resp.Status,
		"status_code": resp.StatusCode,
		"headers":     headers,
		"started_at":  start,
		"finished_at": finish,
	}

	if body != nil {
		output["body"] = string(body)
	}

	var bodyJSON interface{}
	// Just ignore errors

	// Process the response body only if the status code is 200
	if resp != nil && resp.StatusCode == http.StatusOK {
		// The unmarshalling is only done if the content type is JSON,
		// otherwise the unmashalling will fail.
		// Hence, the body_json field will only be populated if the content type is JSON.
		if resp.Header.Get("Content-Type") == "application/json" {
			err = json.Unmarshal(body, &bodyJSON)
			if err != nil {
				logger.Error("error unmarshalling body: %s", err)
				return nil, err
			}
			output["body_json"] = bodyJSON
		}
	}

	return &output, nil
}

// builsHTTPPostInput builds the HTTPPOSTInput struct from the input parameters
func buildHTTPPostInput(input types.Input) (*HTTPPOSTInput, error) {
	// Get the inputs from the pipeline
	inputParams := &HTTPPOSTInput{
		URL: input["url"].(string),

		// TODO: Make it configurable
		RequestTimeoutMs: HTTPRequestDefaultTimeoutMs,
	}

	// Set the certificate, if provided
	if input["ca_cert_pem"] != nil {
		inputParams.CaCertPem = input["ca_cert_pem"].(string)
	}

	// Set value for insecureSkipVerify, if provided
	if input["insecure"] != nil {
		inputParams.Insecure = input["insecure"].(bool)
	}

	// Set the request headers, if provided
	requestHeaders := map[string]interface{}{}
	if input["request_headers"] != nil {
		requestHeaders = input["request_headers"].(map[string]interface{})
	}

	// Get the request body
	requestBody := input["body"]

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
		}
	}
	inputParams.RequestHeaders = requestHeaders

	return inputParams, nil
}
