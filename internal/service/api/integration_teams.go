package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
)

type teamsResponse struct {
	Value               string `json:"value"`
	ExecutionID         string `json:"execution_id"`
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	Prompt              string `json:"prompt"`
}

func (api *APIService) teamsPostHandler(c *gin.Context) {
	var uri types.InputIDHash
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// support for omitting the type in the url
	if !strings.HasPrefix(uri.ID, "teams.") {
		uri.ID = fmt.Sprintf("teams.%s", uri.ID)
	}

	// verify hash
	salt, err := util.GetGlobalSalt()
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}
	hashString, err := util.CalculateHash(uri.ID, salt)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("error calculating hash"))
		return
	}
	if hashString != uri.Hash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash"))
		return
	}

	var resp teamsResponse
	err = c.BindJSON(&resp)
	if err != nil {
		common.AbortWithError(c, perr.BadRequestWithMessage("invalid payload received"))
		return
	}

	var value any
	if strings.Contains(resp.Value, "; ") { // MSTeams puts multiselect into single string with '; ' separator
		value = strings.Split(resp.Value, "; ")

	} else {
		value = resp.Value
	}

	eventPublished, stepExec, err := api.finishInputStep(resp.ExecutionID, resp.PipelineExecutionID, resp.StepExecutionID, value)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	var text string
	if !eventPublished { // only event this is false & we don't have error is that we've already processed the step

		if stepExec.EndTime.After(time.Now().AddDate(-10, 0, 0)) {
			text = fmt.Sprintf("Response was previously received at: %s", stepExec.EndTime.Format(time.RFC1123))
		} else {
			text = "Response was previously received"
		}
	} else {
		text = fmt.Sprintf("Response received: %s", resp.Value)
	}

	c.Header("CARD-UPDATE-IN-BODY", "true")
	c.JSON(http.StatusOK, gin.H{
		"@type":    "MessageCard",
		"@context": "http://schema.org/extensions",
		"summary":  "Received Response",
		"title":    resp.Prompt,
		"text":     text,
	})
}
