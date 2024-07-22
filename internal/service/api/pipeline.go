package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	putils "github.com/turbot/pipe-fittings/utils"
)

func (api *APIService) PipelineRegisterAPI(router *gin.RouterGroup) {
	router.GET("/pipeline", api.listPipelines)
	router.GET("/pipeline/:pipeline_name", api.getPipeline)
	router.POST("/pipeline/:pipeline_name/command", api.cmdPipeline)
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
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /pipeline [get]
func (api *APIService) listPipelines(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slog.Info("received list pipelines request", "next_token", nextToken, "limit", limit)

	result, err := ListPipelines(api.EsService.RootMod.Name())
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func ListPipelines(rootMod string) (*types.ListPipelineResponse, error) {
	pipelines, err := db.ListAllPipelines()
	if err != nil {
		return nil, err
	}

	// Convert the list of pipelines to FpPipeline type
	var listPipelineResponseItems []types.FpPipeline

	for _, pipeline := range pipelines {
		item, err := types.FpPipelineFromModPipeline(pipeline, rootMod)
		if err != nil {
			return nil, err
		}
		listPipelineResponseItems = append(listPipelineResponseItems, *item)
	}

	sort.Slice(listPipelineResponseItems, func(i, j int) bool {
		return listPipelineResponseItems[i].Name < listPipelineResponseItems[j].Name
	})

	// TODO: paging, filter, sorting
	result := &types.ListPipelineResponse{
		Items: listPipelineResponseItems,
	}
	return result, nil
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
// @Success 200 {object} types.FpPipeline
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /pipeline/{pipeline_name} [get]
func (api *APIService) getPipeline(c *gin.Context) {

	var uri types.PipelineRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	getPipelineresponse, err := GetPipeline(uri.PipelineName, api.EsService.RootMod.Name())
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, getPipelineresponse)
}

func GetPipeline(pipelineName string, rootMod string) (*types.FpPipeline, error) {
	pipelineFullName := ConstructPipelineFullyQualifiedName(pipelineName)

	pipelineCached, found := cache.GetCache().Get(pipelineFullName)
	if !found {
		return nil, perr.NotFoundWithMessage("pipeline not found")
	}

	pipeline, ok := pipelineCached.(*modconfig.Pipeline)
	if !ok {
		return nil, perr.NotFoundWithMessage("pipeline not found")
	}

	return types.FpPipelineFromModPipeline(pipeline, rootMod)
}

// @Summary Execute a pipeline command
// @Description Execute a pipeline command
// @ID   pipeline_command
// @Tags Pipeline
// @Accept json
// @Produce json
// / ...
// @Param pipeline_name path string true "The name of the pipeline" format(^[a-z_]{0,32}$)
// @Param request body types.CmdPipeline true "Pipeline command."
// ...
// @Success 200 {object} types.PipelineExecutionResponse
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /pipeline/{pipeline_name}/command [post]
func (api *APIService) cmdPipeline(c *gin.Context) {
	var uri types.PipelineRequestURI
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

	pipelineName := ConstructPipelineFullyQualifiedName(uri.PipelineName)

	executionMode := input.GetExecutionMode()
	waitRetry := input.GetWaitRetry()

	response, pipelineCmd, err := ExecutePipeline(input, "", pipelineName, api.EsService)
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

func ExecutePipeline(input types.CmdPipeline, executionId, pipelineName string, esService *es.ESService) (types.PipelineExecutionResponse, *event.PipelineQueue, error) {
	pipelineDefn, err := db.GetPipeline(pipelineName)
	if err != nil {
		return nil, nil, err
	}

	// Execute the command
	if input.Command != "run" {
		return nil, nil, perr.BadRequestWithMessage("invalid command")
	}

	if len(input.Args) > 0 && len(input.ArgsString) > 0 {
		return nil, nil, perr.BadRequestWithMessage("args and args_string are mutually exclusive")
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewEventForExecutionID(executionId),
		PipelineExecutionID: util.NewPipelineExecutionId(),
		Name:                pipelineDefn.Name(),
	}

	if len(input.Args) > 0 || len(input.ArgsString) == 0 {
		errs := pipelineDefn.ValidatePipelineParam(input.Args)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			return nil, nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
		}
		pipelineCmd.Args = input.Args

	} else if len(input.ArgsString) > 0 {
		args, errs := pipelineDefn.CoercePipelineParams(input.ArgsString)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			return nil, nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
		}
		pipelineCmd.Args = args
	}

	if err := esService.Send(pipelineCmd); err != nil {
		return nil, nil, err
	}

	response := types.PipelineExecutionResponse{
		"flowpipe": map[string]interface{}{
			"execution_id":          pipelineCmd.Event.ExecutionID,
			"pipeline_execution_id": pipelineCmd.PipelineExecutionID,
			"pipeline":              pipelineCmd.Name,
		},
	}
	return response, pipelineCmd, nil
}

func ConstructPipelineFullyQualifiedName(pipelineName string) string {
	return ConstructFullyQualifiedName("pipeline", 1, pipelineName)
}

func ConstructFullyQualifiedName(resourceType string, length int, resourceName string) string {
	// If we run the API server with a mod foo, in order run the pipeline, the API needs the fully-qualified name of the pipeline.
	// For example: foo.pipeline.bar
	// However, since foo is the top level mod, we should be able to just run the pipeline bar
	splitResourceName := strings.Split(resourceName, ".")
	// If the pipeline name provided is not fully qualified
	if len(splitResourceName) == length {
		// Get the root mod name from the cache
		if rootModNameCached, found := cache.GetCache().Get("#rootmod.name"); found {
			if rootModName, ok := rootModNameCached.(string); ok {
				// Prepend the root mod name to the pipeline name to get the fully qualified name
				resourceName = fmt.Sprintf("%s.%s.%s", rootModName, resourceType, resourceName)
			}
		}
	}
	return resourceName
}
