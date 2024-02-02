package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/templates"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) InputRegisterAPI(router *gin.RouterGroup) {
	// router.POST("/input/:input/:hash", api.runInputPost)
	router.POST("/input/slack/:input/:hash", api.runSlackInputPost)
	router.GET("/input/email/:input/:hash", api.runInputEmailGet)
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

type ParsedSlackResponse struct {
	Prompt   string
	UserName string
	Value    any
}

func (api *APIService) runInputEmailGet(c *gin.Context) {
	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	_ = validateInputHash(inputUri)
	// TODO: uncomment if hash validation required
	// if err != nil {
	//   common.AbortWithError(c, err)
	//   return
	// }

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}

	slog.Info("executionMode", "executionMode", executionMode)

	_, pipeExec, stepExec, err := getExecutions(inputQuery.ExecutionID, inputQuery.PipelineExecutionID, inputQuery.StepExecutionID)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	if pipeExec.Status == "finished" {
		alreadyAcknowledgedInputTemplate, err := templates.HTMLTemplate("already-acknowledged-input.html")
		if err != nil {
			slog.Error("error reading the template file", "error", err)
			common.AbortWithError(c, err)
			return
		}
		renderHTMLWithValues(c, string(alreadyAcknowledgedInputTemplate), gin.H{})
	} else {
		acknowledgeInputTemplate, err := templates.HTMLTemplate("acknowledge-input.html")
		if err != nil {
			slog.Error("error reading the template file", "error", err)
			common.AbortWithError(c, err)
			return
		}
		renderHTMLWithValues(c, string(acknowledgeInputTemplate), gin.H{"response": inputQuery.Value})
	}

	err = finishInputStep(api, inputQuery.ExecutionID, inputQuery.PipelineExecutionID, stepExec, inputQuery.Value)
	if err != nil {
		common.AbortWithError(c, err)
	}
}

func (api *APIService) runSlackInputPost(c *gin.Context) {
	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// TODO: Figure out if required, removed validation to make testing easier
	// err := validateInputHash(inputUri)
	// if err != nil {
	// 	common.AbortWithError(c, err)
	// 	return
	// }

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}
	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}
	slog.Info("executionMode", "executionMode", executionMode)

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	decodedBody, err := url.QueryUnescape(string(bodyBytes))
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	decodedBody = decodedBody[8:] // strip non-json prefix

	var jsonBody map[string]any
	err = json.Unmarshal([]byte(decodedBody), &jsonBody)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	var encodedPayload string
	var slackBlockType bool
	if try, ok := jsonBody["callback_id"].(string); ok {
		encodedPayload = try
		slackBlockType = false
	} else if !helpers.IsNil(jsonBody["actions"]) {
		encodedPayload = jsonBody["actions"].([]any)[0].(map[string]any)["action_id"].(string)
		slackBlockType = true
	}

	payload, err := decodePayload(encodedPayload)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slackResponse, err := parseSlackData(jsonBody)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// respond to slack
	c.String(http.StatusOK, fmt.Sprintf("%s <@%s> has selected `%v`", slackResponse.Prompt, slackResponse.UserName, slackResponse.Value))
	if slackBlockType {
		slog.Warn("Slack message not yet updated, therefore may receive future events from it")
		// TODO: figure out how to determine correct integration to call an update message method on
	}

	// restart the pipeline execution
	_, _, stepExec, err := getExecutions(payload.ExecutionID, payload.PipelineExecutionID, payload.StepExecutionID)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	err = finishInputStep(api, payload.ExecutionID, payload.PipelineExecutionID, stepExec, slackResponse.Value)
	if err != nil {
		common.AbortWithError(c, err)
	}
}

func validateInputHash(inputUri types.InputRequestUri) error {
	inputName := inputUri.Input
	inputHash := inputUri.Hash

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		slog.Error("salt not found")
		return perr.InternalWithMessage("salt not found")
	}

	hashString := util.CalculateHash(inputName, salt.(string))
	if hashString != inputHash {
		slog.Error("invalid hash", "hash", inputHash, "input_name", inputName, "expected", hashString)
		return perr.UnauthorizedWithMessage("invalid hash for " + inputName)
	}

	return nil
}

// Custom function to render HTML with values
func renderHTMLWithValues(c *gin.Context, templateContent string, data interface{}) {
	tmpl, err := template.New("html").Parse(templateContent)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse template")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Failed to execute template")
		return
	}
}

func decodePayload(input string) (JSONPayload, error) {
	b64decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return JSONPayload{}, err
	}
	var out JSONPayload
	err = json.Unmarshal(b64decoded, &out)
	if err != nil {
		return JSONPayload{}, err
	}

	return out, nil
}

func parseSlackData(input map[string]any) (ParsedSlackResponse, error) {
	var out ParsedSlackResponse

	// prompt
	if oMsg, ok := input["original_message"].(map[string]any); ok {
		if attachments, ok := oMsg["attachments"].([]any); ok {
			for _, attachment := range attachments {
				out.Prompt = attachment.(map[string]any)["text"].(string)
				break
			}
		}
	} else if oMsg, ok := input["message"].(map[string]any); ok {
		if blocks, ok := oMsg["blocks"].([]any); ok {
			for _, block := range blocks {
				out.Prompt = block.(map[string]any)["text"].(map[string]any)["text"].(string)
				break
			}
		}
	}

	// username
	if user, ok := input["user"].(map[string]any); ok {
		out.UserName = user["name"].(string) // TODO: establish if this should be name or id
	}

	// value
	var values []string
	for _, a := range input["actions"].([]any) {
		action := a.(map[string]any)
		actionType := action["type"].(string)

		switch actionType {
		case constants.InputTypeButton:
			values = append(values, action["value"].(string))
		case constants.InputTypeSelect, "multi_static_select":
			selectedOptions := action["selected_options"].([]any)
			for _, selectedOption := range selectedOptions {
				values = append(values, selectedOption.(map[string]any)["value"].(string))
			}
		}
	}

	switch len(values) {
	case 0:
		out.Value = ""
	case 1:
		out.Value = values[0]
	default:
		out.Value = values
	}

	return out, nil
}

func getExecutions(execId string, pipelineId string, stepId string) (*execution.ExecutionInMemory, *execution.PipelineExecution, *execution.StepExecution, error) {
	ex, err := execution.GetExecution(execId)
	if err != nil {
		return nil, nil, nil, err
	}

	pipelineExecution := ex.PipelineExecutions[pipelineId]
	if pipelineExecution == nil {
		return nil, nil, nil, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution %s not found", pipelineId))
	}

	stepExecution := pipelineExecution.StepExecutions[stepId]
	if stepExecution == nil {
		return nil, nil, nil, perr.NotFoundWithMessage(fmt.Sprintf("step execution %s not found", stepId))
	}

	return ex, pipelineExecution, stepExecution, nil
}

func finishInputStep(api *APIService, execId string, pipelineId string, stepExecution *execution.StepExecution, value any) error {
	evt := &event.Event{ExecutionID: execId, CreatedAt: time.Now()}

	// TODO: decide if we return an error if step already finished

	stepFinishedEvent, err := event.NewStepFinished()
	if err != nil {
		return perr.InternalWithMessage("unable to create step finished event")
	}

	out := modconfig.Output{
		Data: map[string]any{
			"value": value,
		},
		Status: "finished",
	}

	stepFinishedEvent.Event = evt
	stepFinishedEvent.PipelineExecutionID = pipelineId
	stepFinishedEvent.StepExecutionID = stepExecution.ID
	stepFinishedEvent.StepForEach = stepExecution.StepForEach
	stepFinishedEvent.StepLoop = stepExecution.StepLoop
	stepFinishedEvent.StepRetry = stepExecution.StepRetry
	stepFinishedEvent.StepOutput = map[string]any{}
	stepFinishedEvent.Output = &out
	err = api.EsService.Raise(stepFinishedEvent)
	return err
}
