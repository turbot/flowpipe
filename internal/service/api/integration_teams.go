package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/types"
	"github.com/turbot/pipe-fittings/perr"
)

type msTeamsResponse struct {
	Value               string `json:"value"`
	ExecutionID         string `json:"execution_id"`
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	StepExecutionToken  string `json:"step_execution_token"`
	Prompt              string `json:"prompt"`
}

func (api *APIService) msTeamsPostHandler(c *gin.Context) {
	var e perr.ErrorModel
	var uri types.InputIDHash
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// support for omitting the type in the url
	if !strings.HasPrefix(uri.ID, "msteams.") {
		uri.ID = fmt.Sprintf("msteams.%s", uri.ID)
	}

	// verify hash
	hashString, err := util.CalculateHashFromGlobalSalt(uri.ID)
	if err != nil {
		errors.As(err, &e)
		api.msTeamsPostHandlerFail(c, e, true, "[Internal] unable to calculate hash: "+e.Error(), nil)
		return
	}
	if hashString != uri.Hash {
		e := perr.UnauthorizedWithMessage("invalid hash")
		api.msTeamsPostHandlerFail(c, e, true, "[Unauthorized] invalid hash", nil)
		return
	}

	var resp msTeamsResponse
	err = c.BindJSON(&resp)
	if err != nil {
		msg := "[BadRequest] invalid payload received, unable to parse body content"
		api.msTeamsPostHandlerFail(c, perr.BadRequestWithMessage(msg), false, msg, nil)
		return
	}

	hSid, err := util.CalculateHashFromGlobalSalt(resp.StepExecutionID)
	if err != nil || resp.StepExecutionToken == "" || hSid != resp.StepExecutionToken {
		msg := "[Unauthorized] invalid step_execution_token"
		api.msTeamsPostHandlerFail(c, perr.UnauthorizedWithMessage(msg), false, msg, nil)
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
		errors.As(err, &e)
		switch e.Status {
		case http.StatusNotFound: // exec/pexec/sexec not found, replace card cannot succeed
			api.msTeamsPostHandlerFail(c, e, true, "Pipeline instance not found on the server", &resp.Prompt)
			return
		case http.StatusBadRequest: // submitted value invalid, can retry
			api.msTeamsPostHandlerFail(c, e, false, fmt.Sprintf("Error validating submitted response: %s - please amend the response to a valid option and try again", e.Detail), nil)
			return
		case http.StatusInternalServerError: // error submitting event, can retry
			api.msTeamsPostHandlerFail(c, e, false, fmt.Sprintf("Error encountered when responding: %s - please try again", e.Detail), nil)
			return
		}
	}

	var text string
	if !eventPublished { // only event this is false & we don't have error is that we've already processed the step

		if stepExec.EndTime.After(time.Now().AddDate(-10, 0, 0)) {
			text = fmt.Sprintf("Response was previously received at: %s", stepExec.EndTime.Format(time.RFC1123))
		} else {
			text = "Response was previously received"
		}
	} else {
		values, err := parseLabelsFromValues(stepExec.Input, value)
		if err != nil {
			values = fmt.Sprintf("%v", value)
		}
		text = fmt.Sprintf("Response received: %s", values)
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

func (api *APIService) msTeamsPostHandlerFail(c *gin.Context, err perr.ErrorModel, replaceCard bool, msg string, cardTitle *string) {
	// log error
	var requestURL *url.URL
	if c.Request != nil {
		requestURL = c.Request.URL
	}
	slog.Error("Error "+err.Instance,
		"error", err,
		"errorID", err.Instance,
		"requestURL", requestURL)

	if !replaceCard {
		c.Header("CARD-ACTION-STATUS", msg)
		c.AbortWithStatusJSON(http.StatusOK, err)
		return
	}

	title := "Error"
	if cardTitle != nil {
		title = *cardTitle
	}
	c.Header("CARD-UPDATE-IN-BODY", "true")
	c.JSON(http.StatusOK, gin.H{
		"@type":    "MessageCard",
		"@context": "http://schema.org/extensions",
		"summary":  "Error encountered",
		"title":    title,
		"text":     msg,
	})
}
