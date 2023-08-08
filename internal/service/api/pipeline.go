package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
)

func (api *APIService) PipelineRegisterAPI(router *gin.RouterGroup) {
	router.GET("/pipeline", api.listPipelines)
	router.GET("/pipeline/:pipeline_name", api.getPipeline)

	router.POST("/pipeline/:pipeline_name/cmd", api.cmdPipeline)
}

// @Summary List pipelines
// @Description Lists pipelines
// @ID   pipeline_list
// @Tags Pipeline
// @Accept json
// @Produce json
// / ...
// @Param limit query int false "The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25." default(25) minimum(1) maximum(100)
// @Param next_token query string false "When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data."
// ...
// @Success 200 {object} types.ListPipelineResponse
// @Failure 400 {object} fperr.ErrorModel
// @Failure 401 {object} fperr.ErrorModel
// @Failure 403 {object} fperr.ErrorModel
// @Failure 429 {object} fperr.ErrorModel
// @Failure 500 {object} fperr.ErrorModel
// @Router /pipeline [get]
func (api *APIService) listPipelines(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(api.ctx).Info("received list pipelines request", "next_token", nextToken, "limit", limit)

	pipelines, err := db.ListAllPipelines()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// TODO: paging, filter
	result := types.ListPipelineResponse{
		Items: pipelines,
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Get pipeline
// @Description Get pipeline
// @ID   pipeline_get
// @Tags Pipeline
// @Accept json
// @Produce json
// / ...
// @Param pipeline_name path string true "The name of the pipeline" format(^[a-z_]{0,32}$)
// ...
// @Success 200 {object} types.Pipeline
// @Failure 400 {object} fperr.ErrorModel
// @Failure 401 {object} fperr.ErrorModel
// @Failure 403 {object} fperr.ErrorModel
// @Failure 404 {object} fperr.ErrorModel
// @Failure 429 {object} fperr.ErrorModel
// @Failure 500 {object} fperr.ErrorModel
// @Router /pipeline/{pipeline_name} [get]
func (api *APIService) getPipeline(c *gin.Context) {

	var uri types.PipelineRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	pipeline, found := cache.GetCache().Get(uri.PipelineName)
	if !found {
		common.AbortWithError(c, fperr.NotFoundWithMessage("pipeline not found"))
		return
	}

	c.JSON(http.StatusOK, pipeline)
}

// @Summary Execute a pipeline command
// @DescriptionExecute a pipeline command
// @ID   pipeline_cmd
// @Tags Pipeline
// @Accept json
// @Produce json
// / ...
// @Param pipeline_name path string true "The name of the pipeline" format(^[a-z_]{0,32}$)
// @Param request body types.CmdPipeline true "Pipeline command."
// ...
// @Success 200 {object} types.RunPipelineResponse
// @Failure 400 {object} fperr.ErrorModel
// @Failure 401 {object} fperr.ErrorModel
// @Failure 403 {object} fperr.ErrorModel
// @Failure 404 {object} fperr.ErrorModel
// @Failure 429 {object} fperr.ErrorModel
// @Failure 500 {object} fperr.ErrorModel
// @Router /pipeline/{pipeline_name}/cmd [post]
func (api *APIService) cmdPipeline(c *gin.Context) {

	var uri types.PipelineRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	pipeline, err := db.GetPipeline(uri.PipelineName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Validate input data
	var input types.CmdPipeline
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Execute the command
	if input.Command != "run" {
		common.AbortWithError(c, fperr.BadRequestWithMessage("invalid command"))
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(c),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipeline.Name,
	}

	if err := api.esService.Send(pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	response := types.RunPipelineResponse{
		ExecutionID:           pipelineCmd.Event.ExecutionID,
		PipelineExecutionID:   pipelineCmd.PipelineExecutionID,
		ParentStepExecutionID: pipelineCmd.ParentStepExecutionID,
	}
	c.JSON(http.StatusOK, response)
}
