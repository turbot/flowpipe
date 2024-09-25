package primitive

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/pipe-fittings/modconfig"
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
	RoutedUrl           string

	endFunc RoutedInputEndStepFunc
}

type RoutedInputCreatePayload struct {
	// TODO: #refactor - can we use a shared struct with pipes for this?
	ExecutionID         string                         `json:"execution_id"`
	PipelineExecutionID string                         `json:"pipeline_execution_id"`
	StepExecutionID     string                         `json:"step_execution_id"`
	NotifierName        string                         `json:"notifier_name"`
	Inputs              map[string]RoutedInputFormData `json:"inputs"`
}

type RoutedInputResponse struct {
	// TODO: #refactor - can we use a shared struct with pipes for this?
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
	Inputs          map[string]RoutedInputFormData `json:"inputs"`
}

type RoutedInputFormData struct {
	// TODO: #refactor - can we use a shared struct with pipes for this?
	Prompt    string                           `json:"prompt"`
	InputType string                           `json:"input_type"`
	Options   []InputIntegrationResponseOption `json:"options,omitempty"`
	Response  any                              `json:"response,omitempty"`
}

// RoutedInputEndStepFunc is a function that ends a step
type RoutedInputEndStepFunc func(stepExecution *execution.StepExecution, out *modconfig.Output) error

func NewRoutedInput(executionID, pipelineExecutionID, stepExecutionID, pipelineName, stepName, url string, endStepFunc RoutedInputEndStepFunc) *RoutedInput {
	return &RoutedInput{
		ExecutionID:         executionID,
		PipelineExecutionID: pipelineExecutionID,
		StepExecutionID:     stepExecutionID,
		PipelineName:        pipelineName,
		StepName:            stepName,
		RoutedUrl:           url,
		endFunc:             endStepFunc,
	}
}

func NewRoutedInputHttpPayload(executionID, pipelineExecutionID, stepExecutionID, notifierName string, inputs map[string]RoutedInputFormData) *RoutedInputCreatePayload {
	return &RoutedInputCreatePayload{
		ExecutionID:         executionID,
		PipelineExecutionID: pipelineExecutionID,
		StepExecutionID:     stepExecutionID,
		NotifierName:        notifierName,
		Inputs:              inputs,
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

func (r *RoutedInput) ValidateInput(ctx context.Context, i modconfig.Input) error {
	err := validateInputStepInput(ctx, i)
	if err != nil {
		return err // will already be perr
	}

	return nil
}

func (r *RoutedInput) Run(ctx context.Context, i modconfig.Input) (*modconfig.Output, error) {
	// Validate
	if e := r.ValidateInput(ctx, i); e != nil {
		return nil, e
	}

	// Variables
	var inputType, prompt, notifierName string

	if it, ok := i[schema.AttributeTypeType].(string); ok {
		inputType = it
	}

	if p, ok := i[schema.AttributeTypePrompt].(string); ok {
		prompt = p
	}

	if _, ok := i[schema.AttributeTypeNotifier].(map[string]any); ok {
		// TODO: #refactor figure out how to extract notifier name... we currently don't pass this into the step when parsing (only notifies/integrations - even those don't have names just types & details)
		notifierName = "workspace_owners"
	}

	opts := parseOptionsFromInput(i)
	payload := NewRoutedInputHttpPayload(
		r.ExecutionID,
		r.PipelineExecutionID,
		r.StepExecutionID,
		notifierName,
		map[string]RoutedInputFormData{r.GetShortStepName(): *NewRoutedInputHttpPayloadInput(prompt, inputType, opts)})

	output, err := r.execute(ctx, payload)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (r *RoutedInput) execute(ctx context.Context, payload *RoutedInputCreatePayload) (*modconfig.Output, error) {
	output := &modconfig.Output{}

	if payload == nil {
		return nil, perr.BadRequestWithMessage("missing payload")
	}

	// TODO: #refactor #question is this a requirement for all routed inputs? Will they always go to Pipes?
	token := os.Getenv("FLOWPIPE_PIPES_TOKEN")
	if token == "" {
		return nil, perr.InternalWithMessage("missing token")
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

	// TODO: remove this log entry before final release
	slog.Info("RoutedInput creating ..", "payload", string(jsonPayload))
	resp, err := client.Do(req)
	if err != nil {
		return "", perr.InternalWithMessage("failed to execute request")
	}
	defer resp.Body.Close()

	// TODO: #refactor see if we can use some standard shared struct(s)
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", perr.InternalWithMessage("failed to read response body")
	}

	var response RoutedInputResponse
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		return "", perr.InternalWithMessage("failed to unmarshal response body")
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
			time.Sleep(5 * time.Second) // TODO: #refactor better approach - this is at loop initialisation to handle continue from err delay before retry

			count++
			slog.Info("RoutedInput polling ..", "url", pollUrl, "count", count)
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
				slog.Error("failed to unmarshal response body", "error", err)
				continue
			}

			if response.State == "finished" {
				if form, ok := response.Inputs[r.GetShortStepName()]; ok {
					if form.Response != nil {
						out := modconfig.Output{
							Data: map[string]any{
								"value": form.Response,
							},
							Status: "finished",
						}

						ex, err := execution.GetExecution(r.ExecutionID)
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

						if err == nil {
							return
						}
					}
				}
			}

		}
	}()
}
