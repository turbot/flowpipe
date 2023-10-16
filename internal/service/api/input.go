package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

func (api *APIService) InputRegisterAPI(router *gin.RouterGroup) {
	router.POST("/input/:input/:hash", api.runInput)
}

func (api *APIService) runInput(c *gin.Context) {
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

	ex, err := execution.NewExecution(api.ctx)
	if err != nil {
		logger.Error("error creating execution", "error", err)
		common.AbortWithError(c, err)
		return
	}

	event := &event.Event{
		ExecutionID: inputQuery.ExecutionID,
	}

	err = ex.LoadProcess(event)
	if err != nil {
		logger.Error("error loading process", "error", err)
		common.AbortWithError(c, err)
		return
	}

	response := types.PipelineExecutionResponse{
		"flowpipe": map[string]interface{}{
			"execution_id":          "xyz",
			"pipeline_execution_id": "xyz",
		},
	}
	c.JSON(http.StatusOK, response)
}
