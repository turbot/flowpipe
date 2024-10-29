package primitive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/resources"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

const EnvFlowpipeInputRouter = "FLOWPIPE_INPUT_ROUTER"

type RoutedInput struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
	PipelineName        string
	StepName            string
	StepType            string
	RoutedUrl           string

	endFunc RoutedInputEndStepFunc
}

type RoutedInputCreatePayload struct {
	ExecutionID         string                         `json:"execution_id"`
	PipelineExecutionID string                         `json:"pipeline_execution_id"`
	StepExecutionID     string                         `json:"step_execution_id"`
	NotifierName        string                         `json:"notifier_name"`
	StepType            string                         `json:"step_type"`
	Inputs              map[string]RoutedInputFormData `json:"inputs,omitempty"`
	Message             *string                        `json:"message,omitempty"`
	Overrides           *RoutedInputOverrides          `json:"overrides,omitempty"`
}

type RoutedInputResponse struct {
	ID              string                         `json:"id"`
	TenantID        string                         `json:"tenant_id"`
	IdentityID      string                         `json:"identity_id"`
	WorkspaceID     string                         `json:"workspace_id"`
	NotifierID      string                         `json:"notifier_id"`
	Notifier        map[string]any                 `json:"notifier"`
	ProcessID       string                         `json:"process_id"`
	StepExecutionID string                         `json:"step_execution_id"`
	RandomID        string                         `json:"random_id"`
	State           string                         `json:"state"`
	StateReason     string                         `json:"state_reason"`
	StepType        string                         `json:"step_type"`
	Inputs          map[string]RoutedInputFormData `json:"inputs"`
	Message         *string                        `json:"message,omitempty"`
	Overrides       *RoutedInputOverrides          `json:"overrides,omitempty"`
}

type RoutedInputListResponse struct {
	Items     []RoutedInputResponse `json:"items"`
	NextToken *string               `json:"next_token,omitempty"`
}

type RoutedInputFormData struct {
	// TODO: #refactor - can we use a shared struct with pipes for this?
	Prompt    string                           `json:"prompt"`
	InputType string                           `json:"input_type"`
	Options   []InputIntegrationResponseOption `json:"options,omitempty"`
	Response  any                              `json:"response,omitempty"`
}

type RoutedInputOverrides struct {
	To      []string `json:"to,omitempty"`
	Cc      []string `json:"cc,omitempty"`
	Bcc     []string `json:"bcc,omitempty"`
	Subject *string  `json:"subject,omitempty"`
	Channel *string  `json:"channel,omitempty"`
}

// RoutedInputEndStepFunc is a function that ends a step
type RoutedInputEndStepFunc func(stepExecution *execution.StepExecution, out *resources.Output) error

func NewRoutedInput(executionID, pipelineExecutionID, stepExecutionID, pipelineName, stepName, stepType, url string, endStepFunc RoutedInputEndStepFunc) *RoutedInput {
	return &RoutedInput{
		ExecutionID:         executionID,
		PipelineExecutionID: pipelineExecutionID,
		StepExecutionID:     stepExecutionID,
		PipelineName:        pipelineName,
		StepName:            stepName,
		StepType:            stepType,
		RoutedUrl:           url,
		endFunc:             endStepFunc,
	}
}

func NewRoutedInputHttpPayload(executionID, pipelineExecutionID, stepExecutionID, notifierName, stepType string, inputs map[string]RoutedInputFormData, message *string, overrides *RoutedInputOverrides) *RoutedInputCreatePayload {
	return &RoutedInputCreatePayload{
		ExecutionID:         executionID,
		PipelineExecutionID: pipelineExecutionID,
		StepExecutionID:     stepExecutionID,
		NotifierName:        notifierName,
		StepType:            stepType,
		Inputs:              inputs,
		Message:             message,
		Overrides:           overrides,
	}
}

func NewRoutedInputHttpPayloadInput(prompt, inputType string, options []InputIntegrationResponseOption) *RoutedInputFormData {
	return &RoutedInputFormData{
		Prompt:    prompt,
		InputType: inputType,
		Options:   options,
	}
}

func IsInputRouted() bool {
	_, isSet := os.LookupEnv(EnvFlowpipeInputRouter)
	return isSet
}

func GetInputRouter() (string, bool) {
	return os.LookupEnv(EnvFlowpipeInputRouter)
}

func (r *RoutedInput) GetShortStepName() string {
	return strings.Split(r.StepName, ".")[len(strings.Split(r.StepName, "."))-1]
}

func (r *RoutedInput) ValidateInput(ctx context.Context, i resources.Input) error {
	switch r.StepType {
	case "input":
		err := validateInputStepInput(ctx, i)
		if err != nil {
			return err // will already be perr
		}
	case "message":
		// no additional validation required
	}

	return nil
}

func (r *RoutedInput) Run(ctx context.Context, i resources.Input) (*resources.Output, error) {
	// Validate
	if e := r.ValidateInput(ctx, i); e != nil {
		return nil, e
	}

	// Variables
	var inputType, prompt string
	var message *string
	var payload *RoutedInputCreatePayload

	// Notifier
	notifierName := "default"
	if notifier, ok := i[schema.AttributeTypeNotifier].(map[string]any); ok {
		if name, hasName := notifier[schema.AttributeTypeNotifierName].(string); hasName {
			if name == "" {
				return nil, perr.BadRequestWithMessage("Notifier name can not be empty string.")
			}
			notifierName = name
		}
	}

	// Notifier Overrides
	overrides := r.GetOverrides(i)

	switch r.StepType {
	case schema.BlockTypePipelineStepMessage:
		if t, ok := i[schema.AttributeTypeText].(string); ok {
			message = &t
		}
		payload = NewRoutedInputHttpPayload(
			r.ExecutionID,
			r.PipelineExecutionID,
			r.StepExecutionID,
			notifierName,
			r.StepType,
			make(map[string]RoutedInputFormData),
			message,
			overrides)
	case schema.BlockTypePipelineStepInput:
		if it, ok := i[schema.AttributeTypeType].(string); ok {
			inputType = it
		}

		if p, ok := i[schema.AttributeTypePrompt].(string); ok {
			prompt = p
		}

		opts := parseOptionsFromInput(i)
		payload = NewRoutedInputHttpPayload(
			r.ExecutionID,
			r.PipelineExecutionID,
			r.StepExecutionID,
			notifierName,
			r.StepType,
			map[string]RoutedInputFormData{r.GetShortStepName(): *NewRoutedInputHttpPayloadInput(prompt, inputType, opts)},
			nil,
			overrides)
	}

	output, err := r.execute(ctx, payload)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (r *RoutedInput) execute(ctx context.Context, payload *RoutedInputCreatePayload) (*resources.Output, error) {
	output := &resources.Output{}

	if payload == nil {
		return nil, perr.BadRequestWithMessage("missing payload")
	}

	// TODO: #refactor #question is this a requirement for all routed inputs? Will they always go to Pipes?
	token := os.Getenv(app_specific.EnvPipesToken)
	if token == "" {
		return nil, perr.BadRequestWithMessage("missing token for input router")
	}

	client := &http.Client{}

	id, err := r.initialCreate(ctx, client, token, payload)
	if err != nil {
		return nil, err
	}

	slog.Info("RoutedInput created .. running the poller", "id", id)
	r.Poll(ctx, client, token, id)

	return output, nil
}

func (r *RoutedInput) initialCreate(ctx context.Context, client *http.Client, token string, payload *RoutedInputCreatePayload) (string, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal payload", "error", err)
		return "", perr.InternalWithMessage("failed to marshal payload")
	}

	req, err := http.NewRequest("POST", r.RoutedUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error("failed to create request", "error", err)
		return "", perr.InternalWithMessage("failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to execute request", "error", err, "routedUrl", r.RoutedUrl)
		return "", perr.InternalWithMessage("failed to execute request")
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body", "error", err)
		return "", perr.InternalWithMessage("failed to read response body")
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		slog.Error("failed to create routed input", "status", resp.StatusCode, "body", string(resBody))
		return "", perr.InternalWithMessage(fmt.Sprintf("failed to create routed input input, received status code: %d", resp.StatusCode))
	}

	var response RoutedInputResponse
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		slog.Error("failed to unmarshal response body", "error", err)
		return "", perr.InternalWithMessage("failed to unmarshal input router response body")
	}

	return response.ID, nil
}

func (r *RoutedInput) Poll(ctx context.Context, client *http.Client, token string, inputID string) {
	go func() {
		pollUrl, _ := url.JoinPath(r.RoutedUrl, inputID)
		req, _ := http.NewRequest("GET", pollUrl, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		slog.Info("RoutedInput polling start", "url", pollUrl)

		count := 0

		for {
			var err error
			time.Sleep(2 * time.Second) // TODO: #refactor better approach - this is at loop initialisation to handle continue from err delay before retry

			count++
			slog.Debug("RoutedInput polling ..", "url", pollUrl, "count", count)
			resp, err := client.Do(req)
			if err != nil {
				// TODO: #error handle errors in polling?
				continue
			}
			defer resp.Body.Close()

			resBody, err := io.ReadAll(resp.Body)
			if err != nil {
				// TODO: #error handle errors in polling?
				slog.Error("failed to read response body", "error", err)
				continue
			}
			var response RoutedInputResponse
			err = json.Unmarshal(resBody, &response)
			if err != nil {
				// TODO: #error handle errors in polling?
				slog.Error("failed to unmarshal response body", "error", err, "body", string(resBody))
				continue
			}

			var out resources.Output
			switch response.State {
			case "finished":
				switch {
				case r.StepType == schema.BlockTypePipelineStepInput:
					if form, ok := response.Inputs[r.GetShortStepName()]; ok {
						if form.Response != nil {
							out = resources.Output{
								Data: map[string]any{
									"value": form.Response,
								},
								Status: constants.StateFinished,
							}
						}
					}
				case r.StepType == schema.BlockTypePipelineStepMessage:
					out = resources.Output{
						Data:   make(map[string]any),
						Status: constants.StateFinished,
					}
				}

				var ex *execution.ExecutionInMemory
				ex, err = execution.GetExecution(r.ExecutionID)
				if err != nil {
					// TODO: #error handle errors in polling?
					continue
				}
				pipelineExecution := ex.PipelineExecutions[r.PipelineExecutionID]
				stepExecution := pipelineExecution.StepExecutions[r.StepExecutionID]

				err = r.endFunc(stepExecution, &out)
				if err != nil {
					// TODO: #error handle errors in polling?
					slog.Error("failed to end step", "error", err)
					continue
				}
				return
			case "error":
				stateErr := perr.InternalWithMessage(response.StateReason)
				out = resources.Output{
					Status:      constants.StateFailed,
					FailureMode: constants.FailureModeFatal,
					Errors: []resources.StepError{{
						Error:               stateErr,
						PipelineExecutionID: r.PipelineExecutionID,
						StepExecutionID:     r.StepExecutionID,
						Pipeline:            r.PipelineName,
						Step:                r.StepName,
					}},
				}
				var ex *execution.ExecutionInMemory
				ex, err = execution.GetExecution(r.ExecutionID)
				if err != nil {
					// TODO: #error handle errors in polling?
					continue
				}
				pipelineExecution := ex.PipelineExecutions[r.PipelineExecutionID]
				stepExecution := pipelineExecution.StepExecutions[r.StepExecutionID]

				err = r.endFunc(stepExecution, &out)
				if err != nil {
					// TODO: #error handle errors in polling?
					slog.Error("failed to end step", "error", err)
					continue
				}
				return
			}

		}
	}()
}

func (r *RoutedInput) GetOverrides(stepInput resources.Input) *RoutedInputOverrides {
	overrides := &RoutedInputOverrides{}
	hasOverrides := false

	if to, ok := stepInput[schema.AttributeTypeTo].([]any); ok {
		hasOverrides = true
		for _, t := range to {
			if email, ok := t.(string); ok {
				overrides.To = append(overrides.To, email)
			}
		}
	}

	if cc, ok := stepInput[schema.AttributeTypeCc].([]any); ok {
		hasOverrides = true
		for _, c := range cc {
			if email, ok := c.(string); ok {
				overrides.Cc = append(overrides.Cc, email)
			}
		}
	}

	if bcc, ok := stepInput[schema.AttributeTypeBcc].([]any); ok {
		hasOverrides = true
		for _, b := range bcc {
			if email, ok := b.(string); ok {
				overrides.Bcc = append(overrides.Bcc, email)
			}
		}
	}

	if subject, ok := stepInput[schema.AttributeTypeSubject].(string); ok {
		hasOverrides = true
		overrides.Subject = &subject
	}

	if channel, ok := stepInput[schema.AttributeTypeChannel].(string); ok {
		hasOverrides = true
		overrides.Channel = &channel
	}

	if hasOverrides {
		return overrides
	}

	return nil
}
