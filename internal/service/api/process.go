package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
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
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
// @Router /process [get]
func (api *APIService) listProcess(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(api.ctx).Info("received list process request", "next_token", nextToken, "limit", limit)

	result := types.ListProcessResponse{
		Items: []types.Process{},
	}

	result.Items = append(result.Items, types.Process{ID: "123"}, types.Process{ID: "456"})

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
// @Success 200 {object} execution.Execution
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 404 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
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

	c.JSON(http.StatusOK, ex)
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
// @Success 200 {object} modconfig.OutputData
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 404 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
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

	filePath := path.Join(viper.GetString("output.dir"), evt.ExecutionID+"_output.json")

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

	pipelineOutput := output["output"]

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
		common.AbortWithError(c, pcerr.BadRequestWithMessage("invalid command"))
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

		if err := api.esService.Send(pipelineEvent); err != nil {
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

		if err := api.esService.Send(pipelineEvent); err != nil {
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

		if err := api.esService.Send(pipelineEvent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

}

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

	result := types.ListProcessLogResponse{
		Items: logEntries,
	}

	c.JSON(http.StatusOK, result)
}

func (api *APIService) listProcessSps(c *gin.Context) {
	var uri types.ProcessRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	filePath := path.Join(viper.GetString("log.dir"), uri.ProcessId+".sps")

	jsonBytes, err := os.ReadFile(filePath)
	if err != nil {
		fplog.Logger(api.ctx).Error("error reading sps file", "error", err, "file_path", filePath)
		common.AbortWithError(c, pcerr.InternalWithMessage("internal error"))
		return
	}

	// Set the appropriate headers
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=process.sps")

	// Return the JSON content
	c.Data(http.StatusOK, "application/json", jsonBytes)
}
