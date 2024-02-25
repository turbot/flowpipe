package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type slackResponse struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
	User                string
	Value               any
	ResponseUrl         string
	Prompt              string
	isFinished          bool
}

func (s slackResponse) ValueAsString() string {
	switch t := s.Value.(type) {
	case string:
		return t
	case []string:
		return strings.Join(s.Value.([]string), ", ")
	default:
		return fmt.Sprintf("%v", s.Value)
	}
}

type slackUpdate struct {
	Text            string `json:"text"`
	ReplaceOriginal bool   `json:"replace_original"`
}

func (api *APIService) slackPostHandler(c *gin.Context) {
	var uri types.InputIDHash
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
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

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	resp, err := parseSlackResponse(bodyBytes)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	if !resp.isFinished { // acknowledge received event but take no further action
		c.Status(200)
		return
	}

	eventPublished, err := api.finishInputStep(resp.ExecutionID, resp.PipelineExecutionID, resp.StepExecutionID, resp.Value)
	if err != nil {
		common.AbortWithError(c, err)
		return
	} else if !eventPublished { // only event this is false & we don't have error is that we've already processed the step
		common.AbortWithError(c, perr.ConflictWithMessage("already processed"))
		replyMsg := fmt.Sprintf("%s\n<@%s> this was already responded to previously", resp.Prompt, resp.User)
		_ = updateSlackMessage(resp.ResponseUrl, replyMsg)
		return
	} else {
		c.Status(200)
		replyMsg := fmt.Sprintf("%s\n<@%s> responded: %s", resp.Prompt, resp.User, resp.ValueAsString())
		_ = updateSlackMessage(resp.ResponseUrl, replyMsg)
		return
	}
}

func parseSlackResponse(bodyBytes []byte) (slackResponse, error) {
	var response slackResponse
	var values []string
	var encodedPayload string
	var in slack.InteractionCallback

	decodedBody, err := url.QueryUnescape(string(bodyBytes))
	if err != nil {
		return response, err
	}
	decodedBody = decodedBody[8:] // strip non-json prefix
	err = json.Unmarshal([]byte(decodedBody), &in)
	if err != nil {
		return response, err
	}

	response.isFinished = false
	for _, action := range in.ActionCallback.BlockActions {
		if strings.HasPrefix(action.ActionID, "finished") {
			response.isFinished = true
		}
	}
	if !response.isFinished {
		return response, nil
	}

	response.User = in.User.Name
	response.ResponseUrl = in.ResponseURL

	firstBlock := in.Message.Blocks.BlockSet[0]
	switch firstBlock.BlockType() {
	case slack.MBTSection:
		fb := firstBlock.(*slack.SectionBlock)
		response.Prompt = fb.Text.Text
		encodedPayload = fb.BlockID
		values = append(values, in.ActionCallback.BlockActions[0].Value)
	case slack.MBTInput:
		fb := firstBlock.(*slack.InputBlock)
		response.Prompt = fb.Label.Text
		encodedPayload = fb.BlockID
		for _, vs := range in.BlockActionState.Values {
			for _, v := range vs {
				switch v.Type {
				case "static_select":
					values = append(values, v.SelectedOption.Value)
				case "multi_static_select":
					for _, selected := range v.SelectedOptions {
						values = append(values, selected.Value)
					}
				case "plain_text_input":
					values = append(values, v.Value)
				default:
					// ignore
				}
			}
		}
	default:
		return response, perr.BadRequestWithMessage("unexpected payload received from slack")
	}

	payload, err := decodePayload(encodedPayload)
	if err != nil {
		return response, fmt.Errorf("error parsing execution id payload: %s", err.Error())
	}
	response.ExecutionID = payload.ExecutionID
	response.PipelineExecutionID = payload.PipelineExecutionID
	response.StepExecutionID = payload.StepExecutionID

	// value can be nil, string or []string
	switch len(values) {
	case 0:
		response.Value = nil
	case 1:
		response.Value = values[0]
	default:
		response.Value = values
	}

	return response, nil
}

func updateSlackMessage(responseUrl string, newMessage string) error {
	msg := slackUpdate{
		Text:            newMessage,
		ReplaceOriginal: true,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	//nolint: gosec // variable url is by design
	_, err = http.Post(responseUrl, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}

	return nil
}

func decodePayload(input string) (JSONPayload, error) {
	b64decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return JSONPayload{}, err
	}
	var out JSONPayload
	err = json.Unmarshal(b64decoded, &out)
	if err != nil {
		return JSONPayload{}, err
	}

	return out, nil
}
