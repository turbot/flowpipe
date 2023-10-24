package api

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func (api *APIService) WebhookRegisterAPI(router *gin.RouterGroup) {
	router.POST("/hook/:trigger/:hash", api.runWebhook)
}

func (api *APIService) runWebhook(c *gin.Context) {
	logger := fplog.Logger(api.ctx)

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

	executionMode := "asynchronous"
	if webhookQuery.ExecutionMode != nil {
		executionMode = *webhookQuery.ExecutionMode
	}

	webhookTriggerName := webhookUri.Trigger
	webhookTriggerHash := webhookUri.Hash

	// Get the trigger from the cache
	triggerCached, found := cache.GetCache().Get(webhookTriggerName)
	if !found {
		common.AbortWithError(c, perr.NotFoundWithMessage("trigger not found"))
		return
	}

	// check if the t is a webhook t
	t, ok := triggerCached.(*modconfig.Trigger)
	if !ok {
		common.AbortWithError(c, perr.NotFoundWithMessage("object is not a trigger"))
		return
	}

	_, ok = t.Config.(*modconfig.TriggerHttp)
	if !ok {
		common.AbortWithError(c, perr.NotFoundWithMessage("object is not a webhook trigger"))
		return
	}

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	mod := api.EsService.RootMod
	modFullName := t.GetMetadata().ModFullName

	if modFullName != mod.FullName {
		logger.Error("Trigger can only be run from root mod", "trigger", t.Name(), "mod", modFullName, "root_mod", mod.FullName)
		return
	}

	hashString := util.CalculateHash(webhookTriggerName, salt.(string))

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
		Functions: funcs.ContextFunctions(viper.GetString("work.dir")),
	}

	pipelineArgs, diags := t.GetArgs(evalContext)
	if diags.HasErrors() {
		common.AbortWithError(c, error_helpers.HclDiagsToError("trigger", diags))

	}

	pipeline := t.GetPipeline()
	pipelineName := pipeline.AsValueMap()["name"].AsString()

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(c),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineName,
	}

	pipelineCmd.Args = pipelineArgs

	if err := api.EsService.Send(pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	if executionMode == "synchronous" {
		api.waitForPipeline(c, pipelineCmd)
		return
	}

	c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
	c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
	c.String(http.StatusOK, "")
}

func (api *APIService) waitForPipeline(c *gin.Context, pipelineCmd *event.PipelineQueue) {
	logger := fplog.Logger(api.ctx)

	ex, err := execution.NewExecution(api.ctx)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("error creating execution object"))
		return
	}

	waitRetry := 60 // TODO: Make configurable potentially via CLI arg
	waitTime := 1 * time.Second
	expectedState := "finished"

	var pex *execution.PipelineExecution

	// Wait for the pipeline to complete, but not forever
	for i := 0; i < waitRetry; i++ {
		time.Sleep(waitTime)

		err = ex.LoadProcess(pipelineCmd.Event)
		if err != nil {
			logger.Warn("error loading process", "error", err)
			continue
		}

		pex = ex.PipelineExecutions[pipelineCmd.PipelineExecutionID]
		if pex == nil {
			logger.Warn("Pipeline execution not found", "pipeline_execution_id", pipelineCmd.PipelineExecutionID)
			continue
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

	response["flowpipe"] = map[string]interface{}{
		"execution_id":          pipelineCmd.Event.ExecutionID,
		"pipeline_execution_id": pipelineCmd.PipelineExecutionID,
		"status":                pex.Status,
	}

	c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
	c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
	c.Header("flowpipe-status", pex.Status)

	if pex.Status == expectedState {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(209, response)
	}
}
