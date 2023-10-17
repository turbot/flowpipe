package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) InputRegisterAPI(router *gin.RouterGroup) {
	router.POST("/input/:input/:hash", api.runInputPost)
	router.GET("/input/:input/:hash", api.runInputGet)
}

func (api *APIService) runPipeline(c *gin.Context, executionID, pipelineExecutionID, stepExecutionID string) {
	logger := fplog.Logger(api.ctx)

	ex, err := execution.NewExecution(api.ctx)
	if err != nil {
		logger.Error("error creating execution", "error", err)
		common.AbortWithError(c, err)
		return
	}

	evt := &event.Event{
		ExecutionID: executionID,
	}

	err = ex.LoadProcess(evt)
	if err != nil {
		logger.Error("error loading process", "error", err)
		common.AbortWithError(c, err)
		return
	}

	input := primitive.Input{}
	var stepOutput *modconfig.Output

	if c.Request.Body != nil {
		var err error
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}

		stepOutput, err = input.ProcessOutput(api.ctx, bodyBytes)
		if err != nil {
			logger.Error("error processing output", "error", err)
			common.AbortWithError(c, err)
			return
		}

		logger.Debug("stepOutput", "stepOutput", stepOutput)
	}

	// Find the step start for the step execution id
	pipelineExecution := ex.PipelineExecutions[pipelineExecutionID]
	if pipelineExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage("pipeline execution "+pipelineExecutionID+" not found"))
		return
	}

	stepExecution := pipelineExecution.StepExecutions[stepExecutionID]
	if stepExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage("step execution "+stepExecutionID+" not found"))
		return
	}

	pipelineStepFinishedEvent, err := event.NewPipelineStepFinished()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	pipelineStepFinishedEvent.Event = evt
	pipelineStepFinishedEvent.PipelineExecutionID = pipelineExecutionID
	pipelineStepFinishedEvent.StepExecutionID = stepExecutionID
	pipelineStepFinishedEvent.StepForEach = stepExecution.StepForEach
	pipelineStepFinishedEvent.StepOutput = map[string]interface{}{}
	pipelineStepFinishedEvent.Output = stepOutput

	err = api.EsService.Raise(pipelineStepFinishedEvent)
	if err != nil {
		common.AbortWithError(c, err)
	}

	c.JSON(http.StatusOK, "{}")
}

func (api *APIService) runInputGet(c *gin.Context) {
	logger := fplog.Logger(api.ctx)

	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}

	logger.Info("executionMode", "executionMode", executionMode)

	inputName := inputUri.Input
	inputHash := inputUri.Hash

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		logger.Error("salt not found")
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	hashString := util.CalculateHash(inputName, salt.(string))

	if hashString != inputHash {
		logger.Warn("invalid hash, but we're ignoring it for now ... ", "hash", inputHash, "expected", hashString)
		// common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash for "+inputName))
		// return
	}

	api.runPipeline(c, inputQuery.ExecutionID, inputQuery.PipelineExecutionID, inputQuery.StepExecutionID)
}

func (api *APIService) runInputPost(c *gin.Context) {
	logger := fplog.Logger(api.ctx)

	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}

	logger.Info("executionMode", "executionMode", executionMode)

	inputName := inputUri.Input
	inputHash := inputUri.Hash

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		logger.Error("salt not found")
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	hashString := util.CalculateHash(inputName, salt.(string))

	if hashString != inputHash {
		logger.Warn("invalid hash, but we're ignoring it for now ... ", "hash", inputHash, "expected", hashString)
		// common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash for "+inputName))
		// return
	}

	api.runPipeline(c, inputQuery.ExecutionID, inputQuery.PipelineExecutionID, inputQuery.StepExecutionID)
}
