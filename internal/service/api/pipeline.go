package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/utils"
)

func (api *APIService) PipelineRegisterAPI(router *gin.RouterGroup) {
	router.GET("/pipeline", api.listPipelines)
	router.GET("/pipeline/:pipeline_name", api.getPipeline)

	router.POST("/pipeline/:pipeline_name/cmd", api.cmdPipeline)
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

	fplog.Logger(api.ctx).Info("received list pipelines request", "next_token", nextToken, "limit", limit)

	result, err := ListPipelines()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func ListPipelines() (*types.ListPipelineResponse, error) {
	pipelines, err := db.ListAllPipelines()
	if err != nil {
		return nil, err
	}

	// Convert the list of pipelines to FpPipeline type
	var listPipelineResponseItems []types.ListPipelineResponseItem

	for _, pipeline := range pipelines {
		listPipelineResponseItems = append(listPipelineResponseItems, types.ListPipelineResponseItem{
			Name:          pipeline.Name(),
			Description:   pipeline.Description,
			Mod:           pipeline.GetMod().ShortName,
			Title:         pipeline.Title,
			Tags:          pipeline.Tags,
			Documentation: pipeline.Documentation,
		})
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
// @Success 200 {object} types.GetPipelineResponse
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
	getPipelineresponse, err := GetPipeline(uri.PipelineName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, getPipelineresponse)
}

func GetPipeline(pipelineName string) (*types.GetPipelineResponse, error) {
	pipelineFullName := constructPipelineFullyQualifiedName(pipelineName)

	pipelineCached, found := cache.GetCache().Get(pipelineFullName)
	if !found {
		return nil, perr.NotFoundWithMessage("pipeline not found")
	}

	pipeline, ok := pipelineCached.(*modconfig.Pipeline)
	if !ok {
		return nil, perr.NotFoundWithMessage("pipeline not found")
	}

	resp := &types.GetPipelineResponse{
		Name:          pipeline.Name(),
		Description:   pipeline.Description,
		Mod:           pipeline.GetMod().FullName,
		Title:         pipeline.Title,
		Tags:          pipeline.Tags,
		Documentation: pipeline.Documentation,
		Steps:         pipeline.Steps,
		OutputConfig:  pipeline.OutputConfig,
	}

	var pipelineParams []types.FpPipelineParam
	for _, param := range pipeline.Params {

		paramDefault := map[string]interface{}{}
		if !param.Default.IsNull() {
			paramDefaultGoVal, err := hclhelpers.CtyToGo(param.Default)
			if err != nil {
				return nil, perr.BadRequestWithMessage("unable to convert param default to go value: " + param.Name)
			}
			paramDefault[param.Name] = paramDefaultGoVal
		}

		pipelineParams = append(pipelineParams, types.FpPipelineParam{
			Name:        param.Name,
			Description: utils.ToStringPointer(param.Description),
			Optional:    &param.Optional,
			Type:        param.Type.FriendlyName(),
			Default:     paramDefault,
		})

		resp.Params = pipelineParams
	}
	return resp, nil
}

// @Summary Execute a pipeline command
// @DescriptionExecute a pipeline command
// @ID   pipeline_cmd
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
// @Router /pipeline/{pipeline_name}/cmd [post]
func (api *APIService) cmdPipeline(c *gin.Context) {

	var uri types.PipelineRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}
	pipelineName := constructPipelineFullyQualifiedName(uri.PipelineName)

	pipelineDefn, err := db.GetPipeline(pipelineName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Validate input data
	var input types.CmdPipeline
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if input.ExecutionMode != nil {
		executionMode = *input.ExecutionMode
	}
	waitRetry := 60
	if input.WaitRetry != nil {
		waitRetry = *input.WaitRetry
	}

	// Execute the command
	if input.Command != "run" {
		common.AbortWithError(c, perr.BadRequestWithMessage("invalid command"))
		return
	}

	if len(input.Args) > 0 && len(input.ArgsString) > 0 {
		common.AbortWithError(c, perr.BadRequestWithMessage("args and args_string are mutually exclusive"))
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(c),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipelineDefn.Name(),
	}

	if len(input.Args) > 0 || len(input.ArgsString) == 0 {
		errs := pipelineDefn.ValidatePipelineParam(input.Args)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			common.AbortWithError(c, perr.BadRequestWithMessage(strings.Join(errStrs, "; ")))
			return
		}
		pipelineCmd.Args = input.Args

	} else if len(input.ArgsString) > 0 {
		args, errs := pipelineDefn.CoercePipelineParams(input.ArgsString)
		if len(errs) > 0 {
			errStrs := error_helpers.MergeErrors(errs)
			common.AbortWithError(c, perr.BadRequestWithMessage(strings.Join(errStrs, "; ")))
			return
		}
		pipelineCmd.Args = args
	}

	if err := api.EsService.Send(pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	if executionMode == "synchronous" {
		api.waitForPipeline(c, pipelineCmd, waitRetry)
		return
	}

	response := types.PipelineExecutionResponse{
		"flowpipe": map[string]interface{}{
			"execution_id":          pipelineCmd.Event.ExecutionID,
			"pipeline_execution_id": pipelineCmd.PipelineExecutionID,
		},
	}

	if api.ModMetadata.IsStale {
		response["flowpipe"].(map[string]interface{})["is_stale"] = api.ModMetadata.IsStale
		response["flowpipe"].(map[string]interface{})["last_loaded"] = api.ModMetadata.LastLoaded
		c.Header("flowpipe-mod-is-stale", "true")
		c.Header("flowpipe-mod-last-loaded", api.ModMetadata.LastLoaded.Format(time.RFC3339))
	}

	c.JSON(http.StatusOK, response)
}

func constructPipelineFullyQualifiedName(pipelineName string) string {
	// If we run the API server with a mod foo, in order run the pipeline, the API needs the fully-qualified name of the pipeline.
	// For example: foo.pipeline.bar
	// However, since foo is the top level mod, we should be able to just run the pipeline bar
	splitPipelineName := strings.Split(pipelineName, ".")
	// If the pipeline name provided is not fully qualified
	if len(splitPipelineName) == 1 {
		// Get the root mod name from the cache
		if rootModNameCached, found := cache.GetCache().Get("#rootmod.name"); found {
			if rootModName, ok := rootModNameCached.(string); ok {
				// Prepend the root mod name to the pipeline name to get the fully qualified name
				pipelineName = fmt.Sprintf("%s.pipeline.%s", rootModName, pipelineName)
			}
		}
	}
	return pipelineName
}
