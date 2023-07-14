package primitive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

type HTTPRequest struct {
	Input types.Input
}

func (h *HTTPRequest) ValidateInput(ctx context.Context, i types.Input) error {
	if i["url"] == nil {
		return fperr.BadRequestWithMessage("HTTPRequest input must define a url")
	}
	u := i["url"].(string)
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return fmt.Errorf("invalid url: %s", u)
	}
	return nil
}

func (h *HTTPRequest) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	logger := fplog.Logger(ctx)

	// TODO
	// * POST and other methods
	// * Handle server not found errors - https://steampipe.notfound/
	// * Test SSL vs non-SSL
	// * Compare to features in https://www.tines.com/docs/actions/types/http-request#configuration-options

	start := time.Now().UTC()
	resp, err := http.Get(input["url"].(string))
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
