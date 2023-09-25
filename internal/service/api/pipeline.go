package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
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

	pipelines, err := db.ListAllPipelines()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Convert the list of pipelines to FpPipeline type
	var fpPipelines []types.FpPipeline

	for _, pipeline := range pipelines {
		fpPipelines = append(fpPipelines, types.FpPipeline{
			Name:        pipeline.Name(),
			Description: pipeline.Description,
			Mod:         pipeline.GetMod().ShortName,
		})
	}

	sort.Slice(fpPipelines, func(i, j int) bool {
		return fpPipelines[i].Name < fpPipelines[j].Name
	})

	// TODO: paging, filter, sorting
	result := types.ListPipelineResponse{
		Items: fpPipelines,
	}

	c.JSON(http.StatusOK, result)
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
// @Success 200 {object} modconfig.Pipeline
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
	pipelineName := constructPipelineFullyQualifiedName(uri.PipelineName)

	pipelineCached, found := cache.GetCache().Get(pipelineName)
	if !found {
		common.AbortWithError(c, perr.NotFoundWithMessage("pipeline not found"))
		return
	}

	pipeline, ok := pipelineCached.(*modconfig.Pipeline)
	if !ok {
		return
	}

	fpPipeline := types.FpPipeline{
		Name:        pipeline.Name(),
		Description: pipeline.Description,
		Mod:         pipeline.GetMod().FullName,
	}

	c.JSON(http.StatusOK, fpPipeline)
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
// @Success 200 {object} types.RunPipelineResponse
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

	pipeline, err := db.GetPipeline(pipelineName)
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

	// Execute the command
	if input.Command != "run" {
		common.AbortWithError(c, perr.BadRequestWithMessage("invalid command"))
		return
	}

	pipelineCmd := &event.PipelineQueue{
		Event:               event.NewExecutionEvent(c),
		PipelineExecutionID: util.NewPipelineExecutionID(),
		Name:                pipeline.Name(),
	}

	if input.Args != nil {
		pipelineCmd.Args = input.Args
	}

	if err := api.EsService.Send(pipelineCmd); err != nil {
		common.AbortWithError(c, err)
		return
	}

	response := types.RunPipelineResponse{
		ExecutionID:           pipelineCmd.Event.ExecutionID,
		PipelineExecutionID:   pipelineCmd.PipelineExecutionID,
		ParentStepExecutionID: pipelineCmd.ParentStepExecutionID,
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
