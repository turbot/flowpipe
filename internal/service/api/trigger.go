package api

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
)

func (api *APIService) TriggerRegisterAPI(router *gin.RouterGroup) {
	router.GET("/trigger", api.listTriggers)
	router.GET("/trigger/:trigger_name", api.getTrigger)
}

// @Summary List triggers
// @Description Lists triggers
// @ID   trigger_list
// @Tags Trigger
// @Accept json
// @Produce json
// / ...
// @Param limit query int false "The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25." default(25) minimum(1) maximum(100)
// @Param next_token query string false "When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data."
// ...
// @Success 200 {object} types.ListTriggerResponse
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
// @Router /trigger [get]
func (api *APIService) listTriggers(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(api.ctx).Info("received list trigger request", "next_token", nextToken, "limit", limit)

	triggers, err := db.ListAllTriggers()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Convert the list of triggers to FpTrigger type
	var fpTriggers []types.FpTrigger

	for _, trigger := range triggers {
		pipelineInfo := trigger.Pipeline.AsValueMap()
		pipelineName := pipelineInfo["name"].AsString()

		fpTriggers = append(fpTriggers, types.FpTrigger{
			Name:        trigger.Name,
			Type:        trigger.Type,
			Description: trigger.Description,
			Args:        trigger.Args,
			Pipeline:    pipelineName,
		})
	}

	sort.Slice(fpTriggers, func(i, j int) bool {
		return fpTriggers[i].Name < fpTriggers[j].Name
	})

	result := types.ListTriggerResponse{
		Items: fpTriggers,
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Get trigger
// @Description Get trigger
// @ID   trigger_get
// @Tags Trigger
// @Accept json
// @Produce json
// / ...
// @Param trigger_name path string true "The name of the trigger" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.FpTrigger
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 404 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
// @Router /trigger/{trigger_name} [get]
func (api *APIService) getTrigger(c *gin.Context) {

	var uri types.TriggerRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	triggerCached, found := cache.GetCache().Get(uri.TriggerName)
	if !found {
		common.AbortWithError(c, pcerr.NotFoundWithMessage("trigger not found"))
		return
	}

	trigger, ok := triggerCached.(modconfig.ITrigger)
	if !ok {
		return
	}

	// Get the pipeline name from the trigger
	pipelineInfo := trigger.GetPipeline().AsValueMap()
	pipelineName := pipelineInfo["name"].AsString()

	fpTrigger := types.FpTrigger{
		Name:        trigger.GetName(),
		Type:        trigger.GetType(),
		Description: trigger.GetDescription(),
		Args:        trigger.GetArgs(),
		Pipeline:    pipelineName,
	}

	c.JSON(http.StatusOK, fpTrigger)
}
