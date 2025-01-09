package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fperr"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/cache"
	"github.com/turbot/pipe-fittings/perr"
	putils "github.com/turbot/pipe-fittings/utils"
)

func (api *APIService) TriggerRegisterAPI(router *gin.RouterGroup) {
	router.GET("/trigger", api.listTriggers)
	router.GET("/trigger/:trigger_name", api.getTrigger)
	router.POST("/trigger/:trigger_name/command", api.cmdTrigger)
	// router.GET("/trigger/:trigger_name/key", api.listTriggerKeys)
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

	result, err := ListTriggers(api.EsService.RootMod.Name())
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func ListTriggers(rootMod string) (*types.ListTriggerResponse, error) {
	triggers, err := db.ListAllTriggers()
	if err != nil {
		return nil, err
	}

	// Convert the list of triggers to FpTrigger type
	var fpTriggers []types.FpTrigger

	for _, trigger := range triggers {
		fpTrigger, err := types.FpTriggerFromModTrigger(trigger, rootMod)
		if err != nil {
			return nil, err
		}

		fpTriggers = append(fpTriggers, *fpTrigger)
	}

	// Sort the triggers by type, name
	sort.Slice(fpTriggers, func(i, j int) bool {
		if fpTriggers[i].Mod != fpTriggers[j].Mod {
			return fpTriggers[i].Mod < fpTriggers[j].Mod
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

	fpTrigger, err := GetTrigger(triggerName, api.EsService.RootMod.Name())
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, fpTrigger)
}

func GetTrigger(triggerName string, rootMod string) (*types.FpTrigger, error) {
	// If we run the API server with a mod foo, in order get the trigger, the API needs the fully-qualified name of the trigger.
	// For example: foo.trigger.trigger_type.bar
	// However, since foo is the top level mod, we should be able to just get the trigger bar
	splitTriggerName := strings.Split(triggerName, ".")
	// If the trigger name provided is not fully qualified
	var rootModName string
	if len(splitTriggerName) < 4 {
		// Get the root mod name from the cache
		if rootModNameCached, found := cache.GetCache().Get("#rootmod.name"); found {
			var ok bool
			if rootModName, ok = rootModNameCached.(string); ok {
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

	trigger, ok := triggerCached.(*resources.Trigger)
	if !ok {
		return nil, perr.NotFoundWithMessage("trigger not found")
	}

	fpTrigger, err := types.FpTriggerFromModTrigger(*trigger, rootModName)
	if err != nil {
		return nil, err
	}
	return fpTrigger, nil
}

func ConstructTriggerFullyQualifiedName(triggerName string) string {
	return ConstructFullyQualifiedName("trigger", 2, triggerName)
}

func ExecuteTrigger(ctx context.Context, input types.CmdTrigger, executionId, triggerName string, esService *es.ESService) (string, error) {
	_, err := db.GetTrigger(triggerName)
	if err != nil {
		if perr.IsNotFound(err) {
			newErr := perr.NotFoundWithMessage("unable to find trigger " + triggerName)
			newErr.Type = fperr.ErrorCodeResourceNotFound
			fperr.FailOnError(newErr, nil, "")
		}

		return "", err
	}

	executionCmd := event.NewExecutionQueueForTrigger(executionId, triggerName)
	err = esService.CommandBus.Send(ctx, executionCmd)
	if err != nil {
		return "", err
	}

	return executionCmd.Event.ExecutionID, nil
}

// @Summary Execute a trigger command
// @Description Execute a trigger command
// @ID   trigger_command
// @Tags Trigger
// @Accept json
// @Produce json
// / ...
// @Param trigger_name path string true "The name of the trigger" format(^[a-z_]{0,32}$)
// @Param request body types.CmdTrigger true "Trigger command."
// ...
// @Success 200 {object} types.TriggerExecutionResponse
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

	var input types.CmdTrigger
	if err := c.ShouldBindJSON(&input); err != nil {
		slog.Error("error binding input", "error", err)
		common.AbortWithError(c, perr.BadRequestWithMessage(err.Error()))
		return
	}

	triggerName := ConstructTriggerFullyQualifiedName(uri.TriggerName)

	executionMode := input.GetExecutionMode()
	waitRetry := input.GetWaitRetry()

	executionId, err := ExecuteTrigger(c, input, input.ExecutionID, triggerName, api.EsService)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	trg, err := db.GetTrigger(triggerName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	if executionMode == localconstants.ExecutionModeSynchronous {
		triggerExecutionResponse, err := WaitForTrigger(triggerName, executionId, waitRetry)
		if err != nil {
			slog.Error("error waiting for trigger", "error", err)
			common.AbortWithError(c, err)
			return
		}

		api.processTriggerExecutionResult(c, triggerExecutionResponse, event.PipelineQueue{}, err)
		return
	}

	triggerExecutionResponse := types.TriggerExecutionResponse{
		Flowpipe: types.FlowpipeTriggerResponseMetadata{
			ProcessID: executionId,
			Name:      trg.FullName,
			Type:      trg.Config.GetType(),
		},
	}

	if api.ModMetadata.IsStale {
		triggerExecutionResponse.Flowpipe.IsStale = &api.ModMetadata.IsStale
		triggerExecutionResponse.Flowpipe.LastLoaded = &api.ModMetadata.LastLoaded
		c.Header("flowpipe-mod-is-stale", "true")
		c.Header("flowpipe-mod-last-loaded", api.ModMetadata.LastLoaded.Format(putils.RFC3339WithMS))
	}

	c.JSON(http.StatusOK, triggerExecutionResponse)
}

func (api *APIService) processTriggerExecutionResult(c *gin.Context, triggerExecutionResponse types.TriggerExecutionResponse, pipelineCmd event.PipelineQueue, err error) {

	if err != nil {
		if errorModel, ok := err.(perr.ErrorModel); ok {
			pipelineExecutionResponse := types.PipelineExecutionResponse{}

			pipelineExecutionResponse.Flowpipe.ExecutionID = pipelineCmd.Event.ExecutionID
			pipelineExecutionResponse.Flowpipe.PipelineExecutionID = pipelineCmd.PipelineExecutionID
			pipelineExecutionResponse.Flowpipe.Pipeline = pipelineCmd.Name
			pipelineExecutionResponse.Flowpipe.Status = "failed"

			pipelineExecutionResponse.Errors = []resources.StepError{
				{
					PipelineExecutionID: pipelineCmd.PipelineExecutionID,
					Pipeline:            pipelineCmd.Name,
					Error:               errorModel,
				},
			}

			captureGroup := "unknown"
			// correlate which capture group the error is coming from
			for _, res := range triggerExecutionResponse.Results {
				if pipelineExecutionRes, ok := res.(types.PipelineExecutionResponse); ok {
					if pipelineExecutionRes.Flowpipe.PipelineExecutionID == pipelineCmd.PipelineExecutionID {
						captureGroup = pipelineExecutionRes.Flowpipe.Type
					}
				}
			}

			triggerExecutionResponse.Results[captureGroup] = pipelineExecutionResponse

			c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
			c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
			c.Header("flowpipe-status", "failed")

			c.JSON(500, triggerExecutionResponse)
			return
		} else {
			common.AbortWithError(c, err)
			return
		}
	}

	allFinished := true
	for _, res := range triggerExecutionResponse.Results {
		if pipelineExecutionStatus, ok := res.(types.PipelineExecutionResponse); ok {
			if pipelineExecutionStatus.Flowpipe.Status != "finished" {
				allFinished = false
				break
			}
		}
	}

	if allFinished {
		c.JSON(http.StatusOK, triggerExecutionResponse)
	} else {
		c.JSON(209, triggerExecutionResponse)
	}
}

// func (api *APIService) listTriggerKeys(c *gin.Context) {
// 	// Get paging parameters
// 	nextToken, limit, err := common.ListPagingRequest(c)
// 	if err != nil {
// 		common.AbortWithError(c, err)
// 		return
// 	}

// 	slog.Info("received list trigger request", "next_token", nextToken, "limit", limit)

// 	result, err := ListTriggers()
// 	if err != nil {
// 		common.AbortWithError(c, err)
// 		return
// 	}

// 	c.JSON(http.StatusOK, result)
// }
