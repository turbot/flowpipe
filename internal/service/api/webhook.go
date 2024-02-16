package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/turbot/pipe-fittings/utils"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/turbot/pipe-fittings/sanitize"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func (api *APIService) WebhookRegisterAPI(router *gin.RouterGroup) {
	router.POST("/hook/:hook/:hash", api.passHookToProcessor)
	router.GET("/hook/:hook/:hash", api.runTriggerHook)
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

type webformResponse struct {
	ExecutionID         string   `json:"execution_id"`
	PipelineExecutionID string   `json:"pipeline_execution_id"`
	StepExecutionID     string   `json:"step_execution_id"`
	Values              []string `json:"values"`
}

type webformUpdate struct {
	ExecutionID         string `json:"execution_id"`
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	Status              string `json:"status"`
}

type slackResponse struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
	User                string
	Value               any
	ResponseUrl         string
	Prompt              string
	isFinished          bool
}

func (s slackResponse) ValueAsString() string {
	switch t := s.Value.(type) {
	case string:
		return t
	case []string:
		return strings.Join(s.Value.([]string), ", ")
	default:
		return fmt.Sprintf("%v", s.Value)
	}
}

type slackUpdate struct {
	Text            string `json:"text"`
	ReplaceOriginal bool   `json:"replace_original"`
}

func (api *APIService) passHookToProcessor(c *gin.Context) {
	webhookUri := types.WebhookRequestUri{}
	if err := c.ShouldBindUri(&webhookUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	nameParts := strings.Split(webhookUri.Hook, ".")
	switch {
	case len(nameParts) >= 2 && (nameParts[0] == "trigger" || nameParts[1] == "trigger"):
		api.runTriggerHook(c)
	case len(nameParts) >= 2 && nameParts[0] == "integration":
		api.runIntegrationHook(c)
	default:
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("Not Found %s", webhookUri.Hook)))
		return
	}
}

func (api *APIService) runTriggerHook(c *gin.Context) {
	webhookUri := types.WebhookRequestUri{}
	if err := c.ShouldBindUri(&webhookUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	webhookQuery := types.WebhookRequestQuery{}
	if err := c.ShouldBindQuery(&webhookQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	waitRetry := webhookQuery.GetWaitTime()

	webhookTriggerName := webhookUri.Hook
	webhookTriggerHash := webhookUri.Hash

	// Get the trigger from the cache
	triggerCached, found := cache.GetCache().Get(webhookTriggerName)
	if !found {
		common.AbortWithError(c, perr.NotFoundWithMessage("trigger not found"))
		return
	}

	// check if the t is a webhook trigger
	t, ok := triggerCached.(*modconfig.Trigger)
	if !ok {
		common.AbortWithError(c, perr.NotFoundWithMessage("object is not a trigger"))
		return
	}

	// Check if the HTTP trigger is enabled
	// If not enabled, return a 404 error with a custom error type
	if t.Enabled != nil && !*t.Enabled {
		common.AbortWithError(c, perr.NotFoundWithMessageAndType(perr.ErrorCodeTriggerDisabled, "Trigger Disabled"))
		return
	}

	httpTriggerConfig, ok := t.Config.(*modconfig.TriggerHttp)
	if !ok {
		common.AbortWithError(c, perr.NotFoundWithMessage("object is not a webhook trigger"))
		return
	}

	salt, err := util.GetModSaltOrDefault()
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	mod := api.EsService.RootMod
	modFullName := t.GetMetadata().ModFullName

	if modFullName != mod.FullName {
		slog.Error("Trigger can only be run from root mod", "trigger", t.Name(), "mod", modFullName, "root_mod", mod.FullName)
		return
	}

	hashString := util.CalculateHash(webhookTriggerName, salt)

	if hashString != webhookTriggerHash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash for webhook "+webhookTriggerName))
		return
	}

	body := ""
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		body = string(bodyBytes)
	}
	data := map[string]interface{}{}

	data["request_body"] = body
	data["request_headers"] = map[string]string{}
	for k, v := range c.Request.Header {
		data["request_headers"].(map[string]string)[k] = v[0]
	}
	data["url"] = c.Request.RequestURI

	executionVariables := map[string]cty.Value{}

	selfObject := map[string]cty.Value{}
	for k, v := range data {
		ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(v)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		selfObject[k] = ctyVal
	}

	vars := map[string]cty.Value{}
	for _, v := range mod.ResourceMaps.Variables {
		vars[v.GetMetadata().ResourceName] = v.Value
	}

	// "self" is a magic variable that contains the request headers and request body
	// of the webhook.
	//
	// We need to build eval context because we have to use HCL evaluation to get
	// the pipeline args
	executionVariables["self"] = cty.ObjectVal(selfObject)
	executionVariables[schema.AttributeVar] = cty.ObjectVal(vars)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(constants.ArgModLocation)),
	}

	// Get the available methods for the trigger
	var triggerMethods []string
	for method := range httpTriggerConfig.Methods {
		triggerMethods = append(triggerMethods, method)
	}
	requestMethod := strings.ToLower(c.Request.Method)

	// Return error if the request method is not allowed
	if !slices.Contains(triggerMethods, requestMethod) {
		common.AbortWithError(c, perr.MethodNotAllowed())
	}
	triggerMethod := httpTriggerConfig.Methods[requestMethod]

	pipelineArgs, diags := triggerMethod.GetArgs(evalContext)
	if diags.HasErrors() {
		common.AbortWithError(c, error_helpers.HclDiagsToError("trigger", diags))
	}

	pipeline := triggerMethod.Pipeline
	pipelineName := pipeline.AsValueMap()["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
	}

	pipelineCmd.Args = pipelineArgs

	if output.IsServerMode {
		output.RenderServerOutput(c, types.NewServerOutputTriggerExecution(time.Now(), pipelineCmd.Event.ExecutionID, t.Name(), pipelineName))
	}

	if err := api.EsService.Send(pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	if triggerMethod.ExecutionMode == "synchronous" {
		api.waitForPipeline(c, pipelineCmd, waitRetry)
		return
	}

	c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
	c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
	c.String(http.StatusOK, "")
}

func (api *APIService) waitForPipeline(c *gin.Context, pipelineCmd *event.PipelineQueue, waitRetry int) {
	if waitRetry == 0 {
		waitRetry = 60
	}
	waitTime := 1 * time.Second
	expectedState := "finished"

	var pex *execution.PipelineExecution

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < waitRetry; i++ {
		time.Sleep(waitTime)

		ex, err := execution.GetExecution(pipelineCmd.Event.ExecutionID)

		if err != nil {
			if errorModel, ok := err.(perr.ErrorModel); ok {
				response := map[string]interface{}{}

				response["errors"] = []modconfig.StepError{
					{
						PipelineExecutionID: pipelineCmd.PipelineExecutionID,
						Pipeline:            pipelineCmd.Name,
						Error:               errorModel,
					},
				}

				c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
				c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
				c.Header("flowpipe-status", "failed")

				c.JSON(500, response)
				return
			} else {
				common.AbortWithError(c, err)
				return
			}
		}

		pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]
		if pex == nil {
			slog.Warn("Pipeline execution not found", "pipeline_execution_id", pipelineCmd.PipelineExecutionID)
			common.AbortWithError(c, perr.NotFoundWithMessage("pipeline execution not found"))
			return
		}

		if pex.Status == expectedState || pex.Status == "failed" || pex.Status == "finished" {
			break
		}
	}

	if pex == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage("pipeline execution not found"))
		return
	}

	response := pex.PipelineOutput

	if response == nil {
		response = map[string]interface{}{}
	}

	for k, v := range pex.PipelineOutput {
		response[k] = sanitize.Instance.Sanitize(v)
	}

	if response["errors"] != nil {
		response["errors"] = response["errors"].([]modconfig.StepError)
	}

	c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
	c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
	c.Header("flowpipe-status", pex.Status)

	if api.ModMetadata.IsStale {
		response["flowpipe"].(map[string]interface{})["is_stale"] = api.ModMetadata.IsStale
		response["flowpipe"].(map[string]interface{})["last_loaded"] = api.ModMetadata.LastLoaded
		c.Header("flowpipe-mod-is-stale", "true")
		c.Header("flowpipe-mod-last-loaded", api.ModMetadata.LastLoaded.Format(util.RFC3389WithMS))
	}

	if pex.Status == expectedState {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(209, response)
	}
}

func (api *APIService) runIntegrationHook(c *gin.Context) {
	webhookUri := types.WebhookRequestUri{}
	if err := c.ShouldBindUri(&webhookUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// determine integration type: integration.slack.example -> slack
	nameParts := strings.Split(webhookUri.Hook, ".")
	integrationType := nameParts[1]
	integrationName := nameParts[2]

	salt, err := util.GetGlobalSalt()
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}
	hashString := util.CalculateHash(webhookUri.Hook, salt)
	if hashString != webhookUri.Hash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash for integration "+integrationName))
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	switch integrationType {
	case "slack":
		resp, err := parseSlackResponse(bodyBytes)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		if !resp.isFinished { // acknowledge received event but take no further action
			c.Status(200)
			return
		}

		eventPublished, err := api.finishInputStep(resp.ExecutionID, resp.PipelineExecutionID, resp.StepExecutionID, resp.Value)
		if err != nil {
			common.AbortWithError(c, err)
			return
		} else if !eventPublished { // only event this is false & we don't have error is that we've already processed the step
			common.AbortWithError(c, perr.ConflictWithMessage("already processed"))
			replyMsg := fmt.Sprintf("%s\n<@%s> this was already responded to previously", resp.Prompt, resp.User)
			_ = updateSlackMessage(resp.ResponseUrl, replyMsg)
			return
		} else {
			c.Status(200)
			replyMsg := fmt.Sprintf("%s\n<@%s> responded: %s", resp.Prompt, resp.User, resp.ValueAsString())
			_ = updateSlackMessage(resp.ResponseUrl, replyMsg)
			return
		}
	case "webform":
		resp, err := parseWebformResponse(bodyBytes)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		var v any
		switch len(resp.Values) {
		case 0:
			v = ""
		case 1:
			v = resp.Values[0]
		default:
			v = resp.Values
		}

		eventPublished, err := api.finishInputStep(resp.ExecutionID, resp.PipelineExecutionID, resp.StepExecutionID, v)
		if err != nil {
			common.AbortWithError(c, err)
			return
		} else if !eventPublished { // only event this is false & we don't have error is that we've already processed the step
			common.AbortWithError(c, perr.ConflictWithMessage("already processed"))
			return
		}
		c.JSON(200, webformUpdateFromResponse(resp, "finished"))
	default:
		// TODO: handle more gracefully?
		common.AbortWithError(c, perr.BadRequestWithMessage(fmt.Sprintf("Integration type %s is not supported", integrationType)))
		return
	}
}

func (api *APIService) finishInputStep(execId string, pExecId string, sExecId string, value any) (bool, error) {
	ex, err := execution.GetExecution(execId)
	if err != nil {
		return false, perr.NotFoundWithMessage(fmt.Sprintf("execution %s not found", execId))
	}

	pipelineExecution := ex.PipelineExecutions[pExecId]
	if pipelineExecution == nil {
		return false, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution %s not found", pExecId))
	}

	stepExecution := pipelineExecution.StepExecutions[sExecId]
	if stepExecution == nil {
		return false, perr.NotFoundWithMessage(fmt.Sprintf("step execution %s not found", sExecId))
	}

	if stepExecution.Status == "finished" || pipelineExecution.IsFinished() || pipelineExecution.IsFinishing() {
		// step already processed
		return false, nil
	}

	evt := &event.Event{ExecutionID: execId, CreatedAt: time.Now()}
	stepFinishedEvent, err := event.NewStepFinished()
	if err != nil {
		return false, perr.InternalWithMessage("unable to create step finished event: " + err.Error())
	}

	out := modconfig.Output{
		Data: map[string]any{
			"value": value,
		},
		Status: "finished",
	}

	stepFinishedEvent.Event = evt
	stepFinishedEvent.PipelineExecutionID = pExecId
	stepFinishedEvent.StepExecutionID = stepExecution.ID
	stepFinishedEvent.StepForEach = stepExecution.StepForEach
	stepFinishedEvent.StepLoop = stepExecution.StepLoop
	stepFinishedEvent.StepRetry = stepExecution.StepRetry
	stepFinishedEvent.StepOutput = map[string]any{}
	stepFinishedEvent.Output = &out
	err = api.EsService.Raise(stepFinishedEvent)
	if err != nil {
		return false, perr.InternalWithMessage(fmt.Sprintf("error raising step finished event: %s", err.Error()))
	}
	return true, nil
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

func parseSlackResponse(bodyBytes []byte) (slackResponse, error) {
	var response slackResponse
	var values []string
	var encodedPayload string
	decodedBody, err := url.QueryUnescape(string(bodyBytes))
	if err != nil {
		return response, err
	}
	decodedBody = decodedBody[8:] // strip non-json prefix

	var jsonBody map[string]any
	err = json.Unmarshal([]byte(decodedBody), &jsonBody)
	if err != nil {
		return response, err
	}

	// determine if finished
	isFinished := false
	if actions, ok := jsonBody["actions"].([]any); ok {
		for _, a := range actions {
			action := a.(map[string]any)
			if strings.HasPrefix(action["action_id"].(string), "finished") {
				isFinished = true
				break
			}
		}
	}
	response.isFinished = isFinished
	if !isFinished {
		return response, nil
	}

	// parse state / action -> values
	if stateValues, ok := jsonBody["state"].(map[string]any)["values"].(map[string]any); ok && len(stateValues) > 0 {
		for key, value := range stateValues {
			v := value.(map[string]any)
			o := v[utils.SortedMapKeys(v)[0]].(map[string]any)
			switch o["type"].(string) {
			case "static_select":
				encodedPayload = key
				values = append(values, o["selected_option"].(map[string]any)["value"].(string))
				break
			case "multi_static_select":
				encodedPayload = key
				selectedOptions := o["selected_options"].([]any)
				for _, selectedOption := range selectedOptions {
					values = append(values, selectedOption.(map[string]any)["value"].(string))
				}
				break
			case "plain_text_input":
				encodedPayload = key
				values = append(values, o["value"].(string))
				break
			default:
				// ignore
			}
		}
	} else { // button response doesn't have state
		action := jsonBody["actions"].([]any)[0].(map[string]any)
		actionType := action["type"].(string)
		if actionType != "button" {
			return response, fmt.Errorf("error parsing response")
		}
		values = append(values, action["value"].(string))
		firstBlock := jsonBody["message"].(map[string]any)["blocks"].([]any)[0].(map[string]any)
		encodedPayload = firstBlock["block_id"].(string)
	}

	// parse ids - encoded payload should be block_id of first block in message
	payload, err := decodePayload(encodedPayload)
	if err != nil {
		return response, fmt.Errorf("error parsing execution id payload: %s", err.Error())
	}
	response.ExecutionID = payload.ExecutionID
	response.PipelineExecutionID = payload.PipelineExecutionID
	response.StepExecutionID = payload.StepExecutionID

	// parse user
	if user, ok := jsonBody["user"].(map[string]any); ok {
		response.User = user["username"].(string)
	}

	// response url
	response.ResponseUrl = jsonBody["response_url"].(string)

	// parse prompt
	firstBlock := jsonBody["message"].(map[string]any)["blocks"].([]any)[0].(map[string]any)
	if labelBlock, ok := firstBlock["label"].(map[string]any); ok {
		response.Prompt = labelBlock["text"].(string)
	} else if textBlock, ok := firstBlock["text"].(map[string]any); ok {
		response.Prompt = textBlock["text"].(string)
	}

	// value can be nil, string or []string
	switch len(values) {
	case 0:
		response.Value = nil
	case 1:
		response.Value = values[0]
	default:
		response.Value = values
	}

	return response, nil
}

func updateSlackMessage(responseUrl string, newMessage string) error {
	msg := slackUpdate{
		Text:            newMessage,
		ReplaceOriginal: true,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	//nolint: gosec // variable url is by design
	_, err = http.Post(responseUrl, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}

	return nil
}

func parseWebformResponse(bodyBytes []byte) (webformResponse, error) {
	var response webformResponse
	err := json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func webformUpdateFromResponse(response webformResponse, status string) webformUpdate {
	return webformUpdate{
		ExecutionID:         response.ExecutionID,
		PipelineExecutionID: response.PipelineExecutionID,
		StepExecutionID:     response.StepExecutionID,
		Status:              status,
	}
}
