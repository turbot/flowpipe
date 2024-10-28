package api

import (
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig/flowpipe"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
	pfconstants "github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/funcs"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	putils "github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
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

	pipeline, ok := pipelineCached.(*flowpipe.Pipeline)
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

	pipelineExecutionResponse, pipelineCmd, err := ExecutePipeline(input, input.ExecutionID, pipelineName, api.EsService)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	if executionMode == localconstants.ExecutionModeSynchronous {
		waitPipelineExecutionResponse, err := api.waitForPipeline(*pipelineCmd, waitRetry)
		api.processSinglePipelineResult(c, &waitPipelineExecutionResponse, pipelineCmd, err)
		return
	}

	if api.ModMetadata.IsStale {
		pipelineExecutionResponse.Flowpipe.IsStale = &api.ModMetadata.IsStale
		pipelineExecutionResponse.Flowpipe.LastLoaded = &api.ModMetadata.LastLoaded
		c.Header("flowpipe-mod-is-stale", "true")
		c.Header("flowpipe-mod-last-loaded", api.ModMetadata.LastLoaded.Format(putils.RFC3339WithMS))
	}

	c.JSON(http.StatusOK, pipelineExecutionResponse)
}

func (api *APIService) processSinglePipelineResult(c *gin.Context, pipelineExecutionResponse *types.PipelineExecutionResponse, pipelineCmd *event.PipelineQueue, err error) {
	expectedState := "finished"

	if err != nil {
		if errorModel, ok := err.(perr.ErrorModel); ok {
			pipelineExecutionResponse := types.PipelineExecutionResponse{}

			pipelineExecutionResponse.Flowpipe.ExecutionID = pipelineCmd.Event.ExecutionID
			pipelineExecutionResponse.Flowpipe.PipelineExecutionID = pipelineCmd.PipelineExecutionID
			pipelineExecutionResponse.Flowpipe.Pipeline = pipelineCmd.Name
			pipelineExecutionResponse.Flowpipe.Status = "failed"

			pipelineExecutionResponse.Errors = []flowpipe.StepError{
				{
					PipelineExecutionID: pipelineCmd.PipelineExecutionID,
					Pipeline:            pipelineCmd.Name,
					Error:               errorModel,
				},
			}

			c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
			c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
			c.Header("flowpipe-status", "failed")

			c.JSON(500, pipelineExecutionResponse)
			return
		} else {
			common.AbortWithError(c, err)
			return
		}
	}

	c.Header("flowpipe-execution-id", pipelineCmd.Event.ExecutionID)
	c.Header("flowpipe-pipeline-execution-id", pipelineCmd.PipelineExecutionID)
	c.Header("flowpipe-status", pipelineExecutionResponse.Flowpipe.Status)

	if api.ModMetadata.IsStale {
		pipelineExecutionResponse.Flowpipe.IsStale = &api.ModMetadata.IsStale
		pipelineExecutionResponse.Flowpipe.LastLoaded = &api.ModMetadata.LastLoaded
		c.Header("flowpipe-mod-is-stale", "true")
		c.Header("flowpipe-mod-last-loaded", api.ModMetadata.LastLoaded.Format(putils.RFC3339WithMS))
	}

	if pipelineExecutionResponse.Flowpipe.Status == expectedState {
		c.JSON(http.StatusOK, pipelineExecutionResponse)
	} else {
		c.JSON(209, pipelineExecutionResponse)
	}
}

func buildTempEvalContextForApi() (*hcl.EvalContext, error) {
	executionVariables := make(map[string]cty.Value)

	evalContext := &hcl.EvalContext{
		Variables: executionVariables,
		Functions: funcs.ContextFunctions(viper.GetString(pfconstants.ArgModLocation)),
	}

	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return nil, err
	}
	// Why do we add notifier earlier? Because of the param validation before
	notifierMap, err := parse.BuildNotifierMapForEvalContext(fpConfig.Notifiers)
	if err != nil {
		return nil, err
	}

	evalContext.Variables[schema.BlockTypeNotifier] = cty.ObjectVal(notifierMap)

	// **temporarily** add add connections to eval context .. we need to remove them later and only add connections
	// that are used by the pipelines. The connections are special because they may need to be resolved before
	// we use them i.e. temp AWS creds.

	connMap := parse.BuildTemporaryConnectionMapForEvalContext(fpConfig.PipelingConnections)
	evalContext.Variables[schema.BlockTypeConnection] = cty.ObjectVal(connMap)

	return evalContext, nil
}
func ExecutePipeline(input types.CmdPipeline, executionId, pipelineName string, esService *es.ESService) (types.PipelineExecutionResponse, *event.PipelineQueue, error) {
	pipelineDefn, err := db.GetPipeline(pipelineName)
	response := types.PipelineExecutionResponse{}
	if err != nil {
		return response, nil, err
	}

	// Execute the command
	validCommands := map[string]struct{}{
		"run": {},
	}

	if _, ok := validCommands[input.Command]; !ok {
		return response, nil, perr.BadRequestWithMessage("invalid command")
	}

	if len(input.Args) > 0 && len(input.ArgsString) > 0 {
		return response, nil, perr.BadRequestWithMessage("args and args_string are mutually exclusive")
	}

	if input.Command == "run" {
		executionCmd := event.NewExecutionQueueForPipeline(executionId, pipelineDefn.Name())

		evalContext, err := buildTempEvalContextForApi()
		if err != nil {
			return response, nil, err
		}

		if len(input.Args) > 0 || len(input.ArgsString) == 0 {
			errs := parse.ValidateParams(pipelineDefn, input.Args, evalContext)
			if len(errs) > 0 {
				errStrs := error_helpers.MergeErrors(errs)
				return response, nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
			}
			executionCmd.PipelineQueue.Args = input.Args

		} else if len(input.ArgsString) > 0 {
			args, errs := parse.CoerceParams(pipelineDefn, input.ArgsString, evalContext)
			if len(errs) > 0 {
				errStrs := error_helpers.MergeErrors(errs)
				return response, nil, perr.BadRequestWithMessage(strings.Join(errStrs, "; "))
			}
			executionCmd.PipelineQueue.Args = args
		}

		if err := esService.Send(executionCmd); err != nil {
			return response, nil, err
		}

		response.Flowpipe = types.FlowpipeResponseMetadata{
			ExecutionID:         executionCmd.Event.ExecutionID,
			PipelineExecutionID: executionCmd.PipelineQueue.PipelineExecutionID,
			Pipeline:            executionCmd.PipelineQueue.Name,
		}

		// This effectively returns the root pipeline queue command
		return response, executionCmd.PipelineQueue, nil
	}

	return response, nil, perr.BadRequestWithMessage("invalid command")
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
