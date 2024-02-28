package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	router.POST("/hook/:hook/:hash", api.runTriggerHook)
	router.GET("/hook/:hook/:hash", api.runTriggerHook)
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
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
	triggerFullName := fmt.Sprintf("%s.trigger.http.%s", api.EsService.RootMod.ModName, webhookTriggerName)
	triggerCached, found := cache.GetCache().Get(triggerFullName)
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

	hashString, err := util.CalculateHash(webhookTriggerName, salt)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("error validating hash"))
		return
	}

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
