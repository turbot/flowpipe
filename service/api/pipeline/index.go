package pipeline

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/api/common"
)

func RegisterAPI(router *gin.RouterGroup) {
	router.GET("/pipeline", listPipelines)
}

func listPipelines(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(c).Info("received list pipelines request", "next_token", nextToken, "limit", limit)

	result := ""
	c.JSON(http.StatusOK, result)
}
