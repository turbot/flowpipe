package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/es/execution"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/service/api/common"
	"github.com/turbot/flowpipe/types"
)

func (api *APIService) ProcessRegisterAPI(router *gin.RouterGroup) {
	router.GET("/process", api.listProcesss)
	router.GET("/process/:process_id", api.getProcess)
}

// @Summary List processs
// @Description Lists processs
// @ID   process_list
// @Tags Process
// @Accept json
// @Produce json
// / ...
// @Param limit query int false "The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25." default(25) minimum(1) maximum(100)
// @Param next_token query string false "When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data."
// ...
// @Success 200 {object} types.ListProcessResponse
// @Failure 400 {object} fperr.ErrorModel
// @Failure 401 {object} fperr.ErrorModel
// @Failure 403 {object} fperr.ErrorModel
// @Failure 429 {object} fperr.ErrorModel
// @Failure 500 {object} fperr.ErrorModel
// @Router /process [get]
func (api *APIService) listProcesss(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(api.ctx).Info("received list process request", "next_token", nextToken, "limit", limit)

	result := types.ListProcessResponse{
		Items: []types.Process{},
	}

	result.Items = append(result.Items, types.Process{ID: "123"}, types.Process{ID: "456"})

	c.JSON(http.StatusOK, result)
}

// @Summary Get process
// @Description Get process
// @ID   process_get
// @Tags Process
// @Accept json
// @Produce json
// / ...
// @Param process_id path string true "The name of the process" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.Process
// @Failure 400 {object} fperr.ErrorModel
// @Failure 401 {object} fperr.ErrorModel
// @Failure 403 {object} fperr.ErrorModel
// @Failure 404 {object} fperr.ErrorModel
// @Failure 429 {object} fperr.ErrorModel
// @Failure 500 {object} fperr.ErrorModel
// @Router /process/{process_id} [get]
func (api *APIService) getProcess(c *gin.Context) {

	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	logEntries, err := execution.LoadEventLogEntries(uri.ProcessId)
	if err != nil {
		common.AbortWithError(c, err)
	}

	result := types.Process{ID: uri.ProcessId, EventLogEntry: logEntries}

	c.JSON(http.StatusOK, result)
}
