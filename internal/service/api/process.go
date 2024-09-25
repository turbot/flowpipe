package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/metrics"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) ProcessRegisterAPI(router *gin.RouterGroup) {
	router.GET("/process", api.listProcess)
	router.GET("/process/:process_id", api.getProcess)
	router.POST("/process/:process_id/command", api.cmdProcess)
	router.GET("/process/:process_id/log/process.json", api.listProcessEventLog)
	router.GET("/process/:process_id/execution", api.getProcessExecution)
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
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process [get]
func (api *APIService) listProcess(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slog.Info("received list process request", "next_token", nextToken, "limit", limit)

	result, err := ListProcesses()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func ListProcesses() (*types.ListProcessResponse, error) {
	var processList []types.Process

	allExec := metrics.RunMetricInstance.RunningExecutions()
	for _, exMetric := range allExec {
		slog.Debug("allExec", "ex", exMetric)
		ex, err := execution.GetExecution(exMetric.ExecutionID)
		if err != nil {
			slog.Error("Error loading execution", "error", err)
			return nil, err
		}

		processList = append(processList, types.Process{
			ID:        ex.ID,
			Pipeline:  exMetric.Pipeline,
			CreatedAt: exMetric.StartTimestamp,
			Status:    "started", // We assume started as the finished pipeline shouldn't be in the Metrics instance``
		})

	}

	// Extract the execution IDs from the log file names
	executionIDs, err := store.ListExecutionIDs()
	if err != nil {
		slog.Error("Error listing execution IDs", "error", err)
		return nil, perr.InternalWithMessage("Error listing execution IDs")
	}

	// Get the log entries using the execution ID and extract the pipeline name

	for _, execID := range executionIDs {

		evt := &event.Event{
			ExecutionID: execID,
		}

		// TODO .. we need to skip if execution is for a different mod, but how do we know?
		ex, err := execution.NewExecution(context.Background(), execution.WithEvent(evt))
		if err != nil {
			continue
		}

		// get outer pipeline (not child)
		var outerPipeline execution.PipelineExecution
		for _, pipeline := range ex.PipelineExecutions {
			if pipeline.ParentExecutionID == "" && pipeline.ParentStepExecutionID == "" {
				outerPipeline = *pipeline
				continue
			}
		}

		processList = append(processList, types.Process{
			ID:        ex.ID,
			Pipeline:  outerPipeline.Name,
			Status:    outerPipeline.Status,
			CreatedAt: outerPipeline.StartTime,
		})
	}

	return &types.ListProcessResponse{
		Items: processList,
	}, nil
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
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/{process_id} [get]
func (api *APIService) getProcess(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	process, err := GetProcess(uri.ProcessId)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, process)
}

func GetProcess(executionId string) (*types.Process, error) {

	// check in memory first
	ex, err := execution.GetExecution(executionId)
	if err != nil && !perr.IsNotFound(err) {
		return nil, err
	}

	// get outer pipeline (not child)
	var outerPipeline *execution.PipelineExecution

	if ex != nil {
		for _, pex := range ex.PipelineExecutions {
			if pex.ParentExecutionID == "" && pex.ParentStepExecutionID == "" {
				outerPipeline = pex
				break
			}
		}

		if outerPipeline == nil {
			return nil, perr.NotFoundWithMessage("No pipeline found for process " + executionId)
		}

		process := types.Process{
			ID:        ex.ID,
			Pipeline:  outerPipeline.Name,
			CreatedAt: outerPipeline.StartTime,
			Status:    outerPipeline.Status,
		}

		return &process, nil
	}

	// Read the execution from file system
	evt := &event.Event{
		ExecutionID: executionId,
	}

	// WithEvent loads the process
	exFile, err := execution.NewExecution(context.Background(), execution.WithEvent(evt))
	if err != nil {
		return nil, err
	}

	for _, pex := range exFile.PipelineExecutions {
		if pex.ParentExecutionID == "" && pex.ParentStepExecutionID == "" {
			outerPipeline = pex
			break
		}
	}

	if outerPipeline == nil {
		return nil, perr.NotFoundWithMessage("No pipeline found for process " + executionId)
	}

	process := types.Process{
		ID:        exFile.ID,
		Pipeline:  outerPipeline.Name,
		Status:    outerPipeline.Status,
		CreatedAt: outerPipeline.StartTime,
	}

	return &process, nil
}

// @Summary Get process log
// @Description Get process log
// @ID   process_get_log
// @Tags Process
// @Produce json
// / ...
// @Param process_id path string true "The id of the process" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.ListProcessLogJSONResponse
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/{process_id}/log/process.json [get]
func (api *APIService) listProcessEventLog(c *gin.Context) {

	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Get the one in memory first
	ex, err := execution.GetExecution(uri.ProcessId)
	if err == nil && ex != nil {
		result := types.ListProcessLogJSONResponse{
			Items: ex.Events,
		}

		c.JSON(http.StatusOK, result)
		return
	}

	evt := &event.Event{
		ExecutionID: uri.ProcessId,
	}

	exFile, err := execution.NewExecution(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	logEntries, err := exFile.LoadProcessDB(evt)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	result := types.ListProcessLogJSONResponse{
		Items: logEntries,
	}

	c.JSON(http.StatusOK, result)
}

// @Summary Get process execution
// @Description Get process execution
// @ID   process_get_execution
// @Tags Process
// @Accept json
// @Produce json
// / ...
// @Param process_id path string true "The name of the process" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} execution.Execution
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/{process_id}/execution [get]
func (api *APIService) getProcessExecution(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// check in memory first
	ex, err := execution.GetExecution(uri.ProcessId)
	if err != nil && !perr.IsNotFound(err) {
		slog.Error("Error loading execution", "error", err)
		common.AbortWithError(c, perr.InternalWithMessage("Error loading execution"))
		return
	}

	if ex != nil {
		c.JSON(http.StatusOK, ex.Execution)
		return
	}

	evt := &event.Event{
		ExecutionID: uri.ProcessId,
	}

	exFile, err := execution.NewExecution(c, execution.WithEvent(evt))
	if err != nil {
		common.AbortWithError(c, err)
		return

	}

	c.JSON(http.StatusOK, exFile)
}

// @Summary Command process
// @Description Command process
// @ID   process_command
// @Tags Process
// @Accept json
// @Produce json
// / ...
// @Param process_id path string true "The name of the process" format(^[a-z]{0,32}$)
// @Param command body types.CmdProcess true "The command to execute"
// ...
// @Success 200 {object} types.Process
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/{process_id}/command [post]
func (api *APIService) cmdProcess(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	var input types.CmdProcess
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortWithError(c, err)
		return
	}

	if input.Command != "resume" {
		common.AbortWithError(c, perr.BadRequestWithMessage("Invalid command"))
	}

	executionId := uri.ProcessId

	evt := &event.Event{
		ExecutionID: executionId,
	}
	ex, err := execution.LoadExecutionFromProcessDB(evt)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	if ex == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage("execution not found"))
		return
	}

	// Check if execution can be resumed, pipeline must be in a "paused" state
	if !ex.IsPaused() {
		common.AbortWithError(c, perr.BadRequestWithMessage("execution is not paused: "+executionId))
	}

	slog.Info("Resuming execution", "execution_id", executionId)
	// Effectively forever
	ok := cache.GetCache().SetWithTTL(executionId, ex, 10*365*24*time.Hour)
	if !ok {
		slog.Error("Error setting execution in cache", "execution_id", executionId)
		common.AbortWithError(c, perr.InternalWithMessage("Error setting execution in cache"))
		return
	}

	// We have to do this after the execution is stored in the cache, there are code downstream
	// that loads the execution from the cache
	for _, pex := range ex.PipelineExecutions {
		if !pex.IsPaused() {
			common.AbortWithError(c, perr.BadRequestWithMessage("pipeline is not paused: "+pex.ID))
		}
		// Resume the execution
		e := event.NewPipelineResume(ex.ID, pex.ID)
		err := api.EsService.Send(e)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
	}
}
