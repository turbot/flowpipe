package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/cache"
	"github.com/turbot/flowpipe/es/event"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/api/common"
	"github.com/turbot/flowpipe/types"
)

func (api *APIService) PipelineRegisterAPI(router *gin.RouterGroup) {
	router.GET("/pipeline", api.listPipelines)
	router.GET("/pipeline/:pipeline_name", api.getPipeline)

	router.POST("/pipeline/:pipeline_name", api.runPipeline)
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

	result := types.ListPipelineResponse{
		Items: []types.Pipeline{},
	}

	result.Items = append(result.Items, types.Pipeline{Type: "pipeline_sleep", Name: "Foo"}, types.Pipeline{Type: "pipeline_hello", Name: "Bar"})

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

func (api *APIService) runPipeline(c *gin.Context) {
	// Initialize the mod
	cmd := &event.Queue{
		Event:     event.NewExecutionEvent(c),
		Workspace: "e-gineer/scratch",
	}

	if err := api.esService.CommandBus.Send(api.esService.Ctx, cmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event: event.NewExecutionEvent(c),
		Name:  "series_of_for_loop_steps",
	}

	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("XXXX")
	if err := api.esService.CommandBus.Send(api.esService.Ctx, pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}
}
