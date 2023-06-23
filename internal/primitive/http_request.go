package primitive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/pipeline_hcl"
	"github.com/turbot/flowpipe/internal/types"
)

type HTTPRequest struct{}

func NewHTTPRequest(step *pipeline_hcl.PipelineHclStep) *HTTPRequest {
	return &HTTPRequest{}
}

func (h *HTTPRequest) ValidateInput(ctx context.Context, i types.Input) error {
	if i["url"] == nil {
		return errors.New("HTTPRequest input must define a url")
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
		logger.Error("error making request #1", "error", err, "resp", resp)
		if resp != nil {
			return nil, fperr.FromHttpError(err, resp.StatusCode)
		}
		return nil, err
	} else if resp != nil && resp.StatusCode != http.StatusOK {
		logger.Error("error making request #2", "error", err, "resp.StatusCode", resp.StatusCode, "resp.Status", resp.Status)
		return nil, fperr.FromHttpError(fmt.Errorf("%s", resp.Status), resp.StatusCode)
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

	var bodyJSON interface{}
	// Just ignore errors
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		fplog.Logger(ctx).Error("error unmarshalling body: %s", err)
		return nil, err
	}

	output := &types.StepOutput{
		"status":      resp.Status,
		"status_code": resp.StatusCode,
		"headers":     headers,
		"body":        string(body),
		"body_json":   bodyJSON,
		"started_at":  start,
		"finished_at": finish,
	}

	return output, nil
}
