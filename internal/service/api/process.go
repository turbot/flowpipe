package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

func (api *APIService) ProcessRegisterAPI(router *gin.RouterGroup) {
	router.GET("/process", api.listProcess)
	router.GET("/process/:process_id", api.getProcess)
	router.GET("/process/:process_id/output", api.getProcessOutput)
	router.POST("/process/:process_id/cmd", api.cmdProcess)
	router.GET("/process/:process_id/log/process.jsonl", api.listProcessEventLog)
	router.GET("/process/:process_id/log/process.sps", api.listProcessSps)

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

	fplog.Logger(api.ctx).Info("received list process request", "next_token", nextToken, "limit", limit)

	// Read the log directory to list out all the process that have been executed
	logDir := viper.GetString(constants.ArgLogDir)
	processLogFiles, err := os.ReadDir(logDir)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	// Extract the execution IDs from the log file names
	executionIDs := []string{}
	for _, f := range processLogFiles {
		execID := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		executionIDs = append(executionIDs, execID)
	}

	// Get the log entries using the execution ID and extract the pipeline name
	processList := []types.Process{}
	for _, execID := range executionIDs {

		logEntries, err := execution.LoadEventLogEntries(execID)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}

		for _, e := range logEntries {
			if e.EventType == "command.pipeline_queue" {
				var payload *types.ProcessPayload
				err := json.Unmarshal(e.Payload, &payload)
				if err != nil {
					common.AbortWithError(c, err)
					return
				}

				evt := &event.Event{
					ExecutionID: execID,
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

				pex := ex.PipelineExecutions[payload.PipelineExecutionID]
				if pex == nil {
					common.AbortWithError(c, err)
					return
				}

				processList = append(processList, types.Process{
					ID:       execID,
					Pipeline: payload.PipelineName,
					Status:   pex.Status,
				})

				break
			}
		}
	}

	result := types.ListProcessResponse{
		Items: processList,
	}

	c.JSON(http.StatusOK, result)
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

	process := types.Process{
		ID: ex.ID,
	}

	c.JSON(http.StatusOK, process)
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

	evt := &event.Event{
		ExecutionID: uri.ProcessId,
	}

	filePath := path.Join(viper.GetString(constants.ArgOutputDir), evt.ExecutionID+"_output.json")

	// Open the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	defer file.Close()

	// Decode JSON data
	var output map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&output)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	pipelineOutput := types.ProcessOutputData{
		ID:     evt.ExecutionID,
		Output: output,
	}

	c.JSON(http.StatusOK, pipelineOutput)
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

// @Summary Get process logs
// @Description Get process logs
// @ID   process_get_log
// @Tags Process
// @Produce json
// / ...
// @Param process_id path string true "The id of the process" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.ProcessEventLog
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /process/:process_id/log/process.jsonl [get]
func (api *APIService) listProcessEventLog(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	logEntries, err := execution.LoadEventLogEntries(uri.ProcessId)
	if err != nil {
		common.AbortWithError(c, err)
	}

	var items []types.ProcessEventLog
	for _, item := range logEntries {
		items = append(items, types.ProcessEventLog{
			EventType: item.EventType,
			Timestamp: item.Timestamp,
			Payload:   item.Payload,
		})
	}

	result := types.ListProcessLogResponse{
		Items: items,
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

	filePath := path.Join(viper.GetString(constants.ArgLogDir), uri.ProcessId+".sps")

	jsonBytes, err := os.ReadFile(filePath)
	if err != nil {
		fplog.Logger(api.ctx).Error("error reading sps file", "error", err, "file_path", filePath)
		common.AbortWithError(c, perr.InternalWithMessage("internal error"))
		return
	}

	// Set the appropriate headers
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=process.sps")

	// Return the JSON content
	c.Data(http.StatusOK, "application/json", jsonBytes)
}
