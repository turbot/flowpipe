package primitive

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type HTTPRequest struct{}

func (h *HTTPRequest) ValidateInput(ctx context.Context, i pipeline.StepInput) error {
	if i["url"] == nil {
		return errors.New("HTTPRequest input must define a url")
	}
	return nil
}

func (h *HTTPRequest) Run(ctx context.Context, input pipeline.StepInput) (pipeline.StepOutput, error) {
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO
	// * POST and other methods
	// * Handle server not found errors - https://steampipe.notfound/
	// * Test SSL vs non-SSL
	// * Compare to features in https://www.tines.com/docs/actions/types/http-request#configuration-options

	start := time.Now().UTC()
	resp, err := http.Get(input["url"].(string))
	finish := time.Now().UTC()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

	output := pipeline.StepOutput{
		"status":      resp.Status,
		"status_code": resp.StatusCode,
		"headers":     headers,
		"body":        string(body),
		"started_at":  start,
		"finished_at": finish,
	}

	return output, nil
}
