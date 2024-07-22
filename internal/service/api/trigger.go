package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/schema"
	putils "github.com/turbot/pipe-fittings/utils"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) TriggerRegisterAPI(router *gin.RouterGroup) {
	router.GET("/trigger", api.listTriggers)
	router.GET("/trigger/:trigger_name", api.getTrigger)
	router.POST("/trigger/:trigger_name/command", api.cmdTrigger)
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
		fpTrigger := getFpTriggerFromTrigger(trigger)
		fpTriggers = append(fpTriggers, fpTrigger)
	}

	// Sort the triggers by type, name
	sort.Slice(fpTriggers, func(i, j int) bool {
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

	fpTrigger := getFpTriggerFromTrigger(*trigger)
	return &fpTrigger, nil
}

func getFpTriggerFromTrigger(t modconfig.Trigger) types.FpTrigger {
	tt := modconfig.GetTriggerTypeFromTriggerConfig(t.Config)

	fpTrigger := types.FpTrigger{
		Name:            t.FullName,
		Type:            tt,
		Description:     t.Description,
		Title:           t.Title,
		Tags:            t.Tags,
		Documentation:   t.Documentation,
		FileName:        t.FileName,
		StartLineNumber: t.StartLineNumber,
		EndLineNumber:   t.EndLineNumber,
		Enabled:         helpers.IsNil(t.Enabled) || *t.Enabled,
	}

	switch tt {
	case schema.TriggerTypeHttp:
		cfg := t.Config.(*modconfig.TriggerHttp)
		fpTrigger.Url = &cfg.Url
		for _, method := range cfg.Methods {
			pipelineInfo := method.Pipeline.AsValueMap()
			pipelineName := pipelineInfo["name"].AsString()
			fpTrigger.Pipelines = append(fpTrigger.Pipelines, types.FpTriggerPipeline{
				CaptureGroup: method.Type,
				Pipeline:     pipelineName,
			})
		}
	case schema.TriggerTypeQuery:
		cfg := t.Config.(*modconfig.TriggerQuery)
		fpTrigger.Schedule = &cfg.Schedule
		fpTrigger.Query = &cfg.Sql
		for _, capture := range cfg.Captures {
			pipelineInfo := capture.Pipeline.AsValueMap()
			pipelineName := pipelineInfo["name"].AsString()
			fpTrigger.Pipelines = append(fpTrigger.Pipelines, types.FpTriggerPipeline{
				CaptureGroup: capture.Type,
				Pipeline:     pipelineName,
			})
		}
	case schema.TriggerTypeSchedule:
		cfg := t.Config.(*modconfig.TriggerSchedule)
		fpTrigger.Schedule = &cfg.Schedule
		pipelineInfo := t.GetPipeline().AsValueMap()
		pipelineName := pipelineInfo["name"].AsString()
		fpTrigger.Pipelines = append(fpTrigger.Pipelines, types.FpTriggerPipeline{
			CaptureGroup: "default",
			Pipeline:     pipelineName,
		})
	}

	return fpTrigger
}

func ConstructTriggerFullyQualifiedName(triggerName string) string {
	return ConstructFullyQualifiedName("trigger", 2, triggerName)
}

func ExecuteTrigger(ctx context.Context, input types.CmdPipeline, executionId, triggerName string, esService *es.ESService) (types.PipelineExecutionResponse, *event.PipelineQueue, error) {
	modTrigger, err := db.GetTrigger(triggerName)
	if err != nil {
		return nil, nil, err
	}

	triggerRunner := trigger.NewTriggerRunner(ctx, esService.CommandBus, esService.RootMod, modTrigger)

	resp, evt, err := triggerRunner.ExecuteTrigger()
	if err != nil {
		return nil, nil, err
	}

	return resp, evt, nil
}

// @Summary Execute a trigger command
// @Description Execute a trigger command
// @ID   trigger_command
// @Tags Trigger
// @Accept json
// @Produce json
// / ...
// @Param trigger_name path string true "The name of the trigger" format(^[a-z_]{0,32}$)
// @Param request body types.CmdPipeline true "Trigger command."
// ...
// @Success 200 {object} types.PipelineExecutionResponse
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /trigger/{trigger_name}/command [post]
func (api *APIService) cmdTrigger(c *gin.Context) {
	var uri types.TriggerRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	var input types.CmdPipeline
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Error("error binding input", "error", err)
		common.AbortWithError(c, perr.BadRequestWithMessage(err.Error()))
		return
	}

	triggerName := ConstructTriggerFullyQualifiedName(uri.TriggerName)

	executionMode := input.GetExecutionMode()
	waitRetry := input.GetWaitRetry()

	response, pipelineCmd, err := ExecuteTrigger(c, input, "", triggerName, api.EsService)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	if executionMode == localconstants.ExecutionModeSynchronous {
		api.waitForPipeline(c, pipelineCmd, waitRetry)
		return
	}

	if api.ModMetadata.IsStale {
		response["flowpipe"].(map[string]interface{})["is_stale"] = api.ModMetadata.IsStale
		response["flowpipe"].(map[string]interface{})["last_loaded"] = api.ModMetadata.LastLoaded
		c.Header("flowpipe-mod-is-stale", "true")
		c.Header("flowpipe-mod-last-loaded", api.ModMetadata.LastLoaded.Format(putils.RFC3339WithMS))
	}

	c.JSON(http.StatusOK, response)
}
