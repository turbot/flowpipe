package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
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
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /trigger [get]
func (api *APIService) listTriggers(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slog.Info("received list trigger request", "next_token", nextToken, "limit", limit)

	result, err := ListTriggers()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func ListTriggers() (*types.ListTriggerResponse, error) {
	triggers, err := db.ListAllTriggers()
	if err != nil {
		return nil, err
	}

	// Convert the list of triggers to FpTrigger type
	var fpTriggers []types.FpTrigger

	for _, trigger := range triggers {
		pipelineInfo := trigger.Pipeline.AsValueMap()
		pipelineName := pipelineInfo["name"].AsString()

		fpTrigger := types.FpTrigger{
			Name:          trigger.FullName,
			Type:          modconfig.GetTriggerTypeFromTriggerConfig(trigger.Config),
			Description:   trigger.Description,
			Pipeline:      pipelineName,
			Title:         trigger.Title,
			Tags:          trigger.Tags,
			Documentation: trigger.Documentation,
		}

		if tc, ok := trigger.Config.(*modconfig.TriggerHttp); ok {
			fpTrigger.Url = &tc.Url
		} else if tc, ok := trigger.Config.(*modconfig.TriggerSchedule); ok {
			fpTrigger.Schedule = &tc.Schedule
		}
		fpTriggers = append(fpTriggers, fpTrigger)
	}

	// Sort the triggers by pipeline, type, name
	sort.Slice(fpTriggers, func(i, j int) bool {
		if fpTriggers[i].Pipeline != fpTriggers[j].Pipeline {
			return fpTriggers[i].Pipeline < fpTriggers[j].Pipeline
		}
		if fpTriggers[i].Type != fpTriggers[j].Type {
			return fpTriggers[i].Type < fpTriggers[j].Type
		}
		return fpTriggers[i].Name < fpTriggers[j].Name
	})

	result := &types.ListTriggerResponse{
		Items: fpTriggers,
	}
	return result, nil
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
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /trigger/{trigger_name} [get]
func (api *APIService) getTrigger(c *gin.Context) {

	var uri types.TriggerRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}
	triggerName := uri.TriggerName

	fpTrigger, err := GetTrigger(triggerName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, fpTrigger)
}

func GetTrigger(triggerName string) (*types.FpTrigger, error) {
	// If we run the API server with a mod foo, in order get the trigger, the API needs the fully-qualified name of the trigger.
	// For example: foo.trigger.trigger_type.bar
	// However, since foo is the top level mod, we should be able to just get the trigger bar
	splitTriggerName := strings.Split(triggerName, ".")
	// If the trigger name provided is not fully qualified
	if len(splitTriggerName) < 4 {
		// Get the root mod name from the cache
		if rootModNameCached, found := cache.GetCache().Get("#rootmod.name"); found {
			if rootModName, ok := rootModNameCached.(string); ok {
				// Prepend the root mod name to the trigger name to get the fully qualified name
				// For example: foo.trigger.trigger_type.bar
				triggerName = fmt.Sprintf("%s.trigger.%s", rootModName, triggerName)
			}
		}
	}

	triggerCached, found := cache.GetCache().Get(triggerName)
	if !found {
		return nil, perr.NotFoundWithMessage("trigger not found")
	}

	trigger, ok := triggerCached.(*modconfig.Trigger)
	if !ok {
		return nil, perr.NotFoundWithMessage("trigger not found")
	}

	// Get the pipeline name from the trigger
	pipelineInfo := trigger.GetPipeline().AsValueMap()
	pipelineName := pipelineInfo["name"].AsString()

	fpTrigger := &types.FpTrigger{
		Name:            trigger.FullName,
		Type:            modconfig.GetTriggerTypeFromTriggerConfig(trigger.Config),
		Description:     trigger.Description,
		Pipeline:        pipelineName,
		Title:           trigger.Title,
		Tags:            trigger.Tags,
		Documentation:   trigger.Documentation,
		FileName:        trigger.FileName,
		StartLineNumber: trigger.StartLineNumber,
		EndLineNumber:   trigger.EndLineNumber,
	}

	if tc, ok := trigger.Config.(*modconfig.TriggerHttp); ok {
		fpTrigger.Url = &tc.Url
	} else if tc, ok := trigger.Config.(*modconfig.TriggerSchedule); ok {
		fpTrigger.Schedule = &tc.Schedule
	}

	return fpTrigger, nil
}
