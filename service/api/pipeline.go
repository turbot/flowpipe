package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/api/common"
)

func (api *APIService) PipelineRegisterAPI(router *gin.RouterGroup) {
	router.GET("/pipeline", api.listPipelines)
}

func (api *APIService) listPipelines(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(api.ctx).Info("received list pipelines request", "next_token", nextToken, "limit", limit)

	result := "pipelines ..."
	c.JSON(http.StatusOK, result)
}
