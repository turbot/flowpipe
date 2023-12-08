package api

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) ProcessRegisterAPI(router *gin.RouterGroup) {
	router.GET("/process", api.listProcess)
	router.GET("/process/:process_id", api.getProcess)
	router.GET("/process/:process_id/output", api.getProcessOutput)
	router.POST("/process/:process_id/cmd", api.cmdProcess)
	router.GET("/process/:process_id/log/process.json", api.listProcessEventLog)
	router.GET("/process/:process_id/log/process.jsonl", api.listProcessEventLogJSONLine)
	router.GET("/process/:process_id/log/process.sps", api.listProcessSps)
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
	// Read the log directory to list out all the process that have been executed
	eventStoreDir := filepaths.EventStoreDir()
	processLogFiles, err := os.ReadDir(eventStoreDir)
	if err != nil {
		return nil, err
	}

	// Extract the execution IDs from the log file names
	var executionIDs []string
	for _, f := range processLogFiles {
		execID := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		executionIDs = append(executionIDs, execID)
	}

	// Get the log entries using the execution ID and extract the pipeline name
	var processList []types.Process
	for _, execID := range executionIDs {

		evt := &event.Event{
			ExecutionID: execID,
		}

		ex, err := execution.NewExecution(context.Background(), execution.WithEvent(evt))
		if err != nil {
			// Skip if the execution is for a different mod
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
	var process types.Process
	evt := &event.Event{
		ExecutionID: executionId,
	}

	// WithEvent loads the process
	ex, err := execution.NewExecution(context.Background(), execution.WithEvent(evt))
	if err != nil {
		return nil, err
	}

	// get outer pipeline (not child)
	var outerPipeline execution.PipelineExecution
	for _, pipeline := range ex.PipelineExecutions {
		if pipeline.ParentExecutionID == "" && pipeline.ParentStepExecutionID == "" {
			outerPipeline = *pipeline
			continue
		}
	}

	process = types.Process{
		ID:        ex.ID,
		Pipeline:  outerPipeline.Name,
		Status:    outerPipeline.Status,
		CreatedAt: outerPipeline.StartTime,
	}

	return &process, nil
}

// TODO: temp API for All Hands demo
// @Summary Get process output
// @Description Get process output
// @ID   process_get_output
// @Tags Process
// @Accept json
// @Produce json
// / ...
// @Param process_id path string true "The name of the process" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.ProcessOutputData
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/{process_id}/output [get]
func (api *APIService) getProcessOutput(c *gin.Context) {

	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// evt := &event.Event{
	// 	ExecutionID: uri.ProcessId,
	// }

	// // Open the JSON file
	// file, err := os.Open(outputPath)
	// if err != nil {
	// 	common.AbortWithError(c, err)
	// 	return
	// }
	// defer file.Close()

	// // Decode JSON data
	// var output map[string]interface{}
	// decoder := json.NewDecoder(file)
	// err = decoder.Decode(&output)
	// if err != nil {
	// 	common.AbortWithError(c, err)
	// 	return
	// }

	// pipelineOutput := types.ProcessOutputData{
	// 	ID:     evt.ExecutionID,
	// 	Output: output,
	// }

	c.JSON(http.StatusOK, "")
}

func (api *APIService) cmdProcess(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Validate input data
	var input types.CmdProcess
	if err := c.ShouldBindJSON(&input); err != nil {
		common.AbortWithError(c, err)
		return
	}

	if input.Command != "cancel" && input.Command != "pause" && input.Command != "resume" {
		common.AbortWithError(c, perr.BadRequestWithMessage("invalid command"))
		return
	}

	if input.Command == "cancel" {
		// Raise the event.PipelineCancel event .. but will actually handled by command.PipelineCancel command handler
		// the command to event binding is in the NewCommand() function
		pipelineEvent := event.PipelineCancel{
			Event:               event.NewEventForExecutionID(uri.ProcessId),
			PipelineExecutionID: input.PipelineExecutionID,
			ExecutionID:         uri.ProcessId,
			Reason:              "because I said so",
		}

		if err := api.EsService.Send(pipelineEvent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if input.Command == "pause" {
		// Raise the event.PipelinePause event .. but will actually handled by command.PipelinePause command handler
		// the command to event binding is in the NewCommand() function
		pipelineEvent := &event.PipelinePause{
			Event:               event.NewEventForExecutionID(uri.ProcessId),
			PipelineExecutionID: input.PipelineExecutionID,
			ExecutionID:         uri.ProcessId,
			Reason:              "just because",
		}

		if err := api.EsService.Send(pipelineEvent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if input.Command == "resume" {
		pipelineEvent := &event.PipelineResume{
			Event:               event.NewEventForExecutionID(uri.ProcessId),
			PipelineExecutionID: input.PipelineExecutionID,
			ExecutionID:         uri.ProcessId,
			Reason:              "just because",
		}

		if err := api.EsService.Send(pipelineEvent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

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

	logEntries, err := execution.LoadEventStoreEntries(uri.ProcessId)
	if err != nil {
		common.AbortWithError(c, err)
	}

	var items []types.ProcessEventLog
	for _, item := range logEntries {
		items = append(items, types.ProcessEventLog{
			EventType: item.EventType,
			Timestamp: item.Timestamp,
			Payload:   string(item.Payload),
		})
	}

	result := types.ListProcessLogJSONResponse{
		Items: items,
	}

	c.JSON(http.StatusOK, result)
}

func (api *APIService) listProcessEventLogJSONLine(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	logEntries, err := execution.LoadEventStoreEntries(uri.ProcessId)
	if err != nil {
		common.AbortWithError(c, err)
	}
	result := types.ListProcessLogResponse{
		Items: logEntries,
	}
	c.JSON(http.StatusOK, result)
}

// @Summary Get process snapshot
// @Description Get process snapshot
// @ID   process_get_snapshot
// @Tags Process
// @Produce json
// / ...
// @Param process_id path string true "The id of the process" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} execution.Snapshot
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/:process_id/log/process.sps [get]
func (api *APIService) listProcessSps(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	snapshotPath := path.Join(filepaths.EventStoreDir(), uri.ProcessId+".sps")

	jsonBytes, err := os.ReadFile(snapshotPath)
	if err != nil {
		slog.Error("error reading sps file", "error", err, "file_path", snapshotPath)
		common.AbortWithError(c, perr.InternalWithMessage("internal error"))
		return
	}

	// Set the appropriate headers
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=process.sps")

	// Return the JSON content
	c.Data(http.StatusOK, "application/json", jsonBytes)
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

	evt := &event.Event{
		ExecutionID: uri.ProcessId,
	}

	ex, err := execution.NewExecution(c, execution.WithEvent(evt))
	if err != nil {
		common.AbortWithError(c, err)
		return

	}

	err = ex.LoadProcess(evt)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, ex)
}
