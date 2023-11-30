package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type InputType string

const (
	InputTypeSlack InputType = "slack"
	InputTypeEmail InputType = "email"
)

type Input struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

type InputIntegration interface {
	PostMessage(modconfig.Input) error
	ReceiveMessage() (*modconfig.Output, error)
}

type InputIntegrationSlack struct {
	InputIntegrationBase
}

type InputIntegrationBase struct {
	ExecutionID         string
	PipelineExecutionID string
	StepExecutionID     string
}

func (ip *InputIntegrationSlack) PostMessage(input modconfig.Input) error {
	// Set the slack user token provided in the input config
	userToken := input[schema.AttributeTypeToken].(string)
	api := slack.New(userToken)

	// Set the slack user token provided in the input config
	channelID := input[schema.AttributeTypeChannel].(string)

	// Encode the callback_id to pass to the interactive element
	payload := map[string]interface{}{
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
		"execution_id":          ip.ExecutionID,
	}
	unmarshaledAdditionalInfo, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	encodedText := base64.StdEncoding.EncodeToString(unmarshaledAdditionalInfo)

	// Check if the user has already made a selection
	userHasMadeSelection := false

	var attachment slack.Attachment

	slackType := input[schema.AttributeTypeSlackType].(string)
	prompt := input[schema.AttributeTypePrompt].(string)

	// Get the options
	var options []string
	if _, ok := input[schema.AttributeTypeOptions].([]string); ok {
		options = input[schema.AttributeTypeOptions].([]string)
	}
	if _, ok := input[schema.AttributeTypeOptions].([]interface{}); ok {
		for _, v := range input[schema.AttributeTypeOptions].([]interface{}) {
			options = append(options, v.(string))
		}
	}

	// Check for the prompt
	if slackType == "button" {
		attachment = slack.Attachment{
			Text:       prompt,
			Color:      "#3AA3E3",
			CallbackID: encodedText,
		}

		var actions []slack.AttachmentAction
		for _, opt := range options {
			actions = append(actions, slack.AttachmentAction{
				Name:  opt,
				Text:  opt,
				Type:  "button",
				Value: opt,
			})
		}

		if len(actions) > 0 {
			attachment.Actions = actions
		}
	} else if slackType == "select" {
		attachment = slack.Attachment{
			Text:       prompt,
			Color:      "#3AA3E3",
			CallbackID: encodedText,
		}

		var actions []slack.AttachmentAction
		var actionOptions []slack.AttachmentActionOption
		for _, opt := range options {
			actionOptions = append(actionOptions, slack.AttachmentActionOption{
				Text:  opt,
				Value: opt,
			})
		}

		if len(actionOptions) > 0 {
			actions = []slack.AttachmentAction{
				{
					Name:    "Choose an option",
					Text:    "Choose an option",
					Type:    "select",
					Options: actionOptions,
				},
			}
			attachment.Actions = actions
		}
	}

	// Remove the interactive element if the user has made a selection
	if userHasMadeSelection {
		attachment.Actions = nil
	}

	_, _, err = api.PostMessage(channelID,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionAsUser(true))

	if err != nil {
		return err
	}
	return nil
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

func (*InputIntegrationSlack) ReceiveMessage(ctx context.Context, requestBody []byte) (*modconfig.Output, error) {
	var bodyJSON map[string]interface{}

	err := json.Unmarshal(requestBody, &bodyJSON)
	if err != nil {
		return nil, err
	}

	// Decode the callback_id to extract the execution_id, pipeline_execution_id and step_execution_id
	rawDecodedText, err := base64.StdEncoding.DecodeString(bodyJSON["callback_id"].(string))
	if err != nil {
		return nil, err
	}

	var decodedText JSONPayload
	err = json.Unmarshal(rawDecodedText, &decodedText)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"pipeline_execution_id": decodedText.PipelineExecutionID,
		"step_execution_id":     decodedText.StepExecutionID,
		"execution_id":          decodedText.ExecutionID,
	}

	// If the input type isa button, then the value is a string
	// if the input is multi-select box, then the value is a list
	var value interface{}
	if bodyJSON["actions"] != nil {
		for _, action := range bodyJSON["actions"].([]interface{}) {
			if action.(map[string]interface{})["type"] == "button" {
				value = action.(map[string]interface{})["value"]
			}
		}
	}
	data["value"] = value

	o := modconfig.Output{
		Data: data,
	}

	return &o, nil
}

type InputIntegrationEmail struct {
	InputIntegrationBase
}

func (inputEmail *InputIntegrationEmail) ValidateInputIntegrationEmail(ctx context.Context, i modconfig.Input) error {

	// Validate sender's information
	if i[schema.AttributeTypeFrom] == nil {
		return perr.BadRequestWithMessage("Email input must define from")
	}
	if i[schema.AttributeTypeSmtpPassword] == nil {
		return perr.BadRequestWithMessage("Email input must define sender_credential")
	}
	if i[schema.AttributeTypeHost] == nil {
		return perr.BadRequestWithMessage("Email input must define a SMTP host")
	}
	if i[schema.AttributeTypePort] == nil {
		return perr.BadRequestWithMessage("Email input must define a port")
	}

	// Validate the port input
	if i[schema.AttributeTypePort] != nil {
		var port int64
		switch data := i[schema.AttributeTypePort].(type) {
		case float64:
			port = int64(data)
		case int64:
			port = data
		default:
			return perr.BadRequestWithMessage("port must be a number")
		}

		portInString := strconv.FormatInt(port, 10)
		match, err := regexp.MatchString("^((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4}))$", portInString)
		if err != nil {
			return perr.BadRequestWithMessage("error while validating the port")
		}

		if !match {
			return perr.BadRequestWithMessage(fmt.Sprintf("%s is not a valid port", portInString))
		}
	}

	if i[schema.AttributeTypeSenderName] != nil {
		if _, ok := i[schema.AttributeTypeSenderName].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'sender_name' must be a string")
		}
	}

	// Validate the recipients
	if i[schema.AttributeTypeTo] == nil {
		return perr.BadRequestWithMessage("Email input must define to")
	}
	if _, ok := i[schema.AttributeTypeTo].([]string); !ok {
		// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
		// Check if the input is an interface slice, and the elements are strings
		if data, ok := i[schema.AttributeTypeTo].([]interface{}); ok {
			for _, v := range data {
				if _, ok := v.(string); !ok {
					return perr.BadRequestWithMessage("Email attribute 'to' must have elements of type string")
				}
			}
			return nil
		}

		return perr.BadRequestWithMessage("Email attribute 'to' must be an array")
	}

	var recipients []string
	if _, ok := i[schema.AttributeTypeTo].([]string); ok {
		recipients = i[schema.AttributeTypeTo].([]string)
	}

	if _, ok := i[schema.AttributeTypeTo].([]interface{}); ok {
		for _, v := range i[schema.AttributeTypeTo].([]interface{}) {
			recipients = append(recipients, v.(string))
		}
	}

	if len(recipients) == 0 {
		return perr.BadRequestWithMessage("Recipients must not be empty")
	}

	// Validate the Cc recipients
	if i[schema.AttributeTypeCc] != nil {
		if _, ok := i[schema.AttributeTypeCc].([]string); !ok {
			// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
			// Check if the input is an interface slice, and the elements are strings
			if data, ok := i[schema.AttributeTypeCc].([]interface{}); ok {
				for _, v := range data {
					if _, ok := v.(string); !ok {
						return perr.BadRequestWithMessage("Email attribute 'cc' must have elements of type string")
					}
				}
				return nil
			}

			return perr.BadRequestWithMessage("Email attribute 'cc' must be an array")
		}
	}

	// Validate the Bcc recipients
	if i[schema.AttributeTypeBcc] != nil {
		if _, ok := i[schema.AttributeTypeBcc].([]string); !ok {
			// The given input is a string slice, but the step input stores it as an interface slice during the JSON unmarshalling?
			// Check if the input is an interface slice, and the elements are strings
			if data, ok := i[schema.AttributeTypeBcc].([]interface{}); ok {
				for _, v := range data {
					if _, ok := v.(string); !ok {
						return perr.BadRequestWithMessage("Email attribute 'bcc' must have elements of type string")
					}
				}
				return nil
			}

			return perr.BadRequestWithMessage("Email attribute 'bcc' must be an array")
		}
	}

	// Validate the email body
	if i[schema.AttributeTypeBody] != nil {
		if _, ok := i[schema.AttributeTypeBody].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'body' must be a string")
		}
	}

	if i[schema.AttributeTypeContentType] != nil {
		if _, ok := i[schema.AttributeTypeContentType].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'content_type' must be a string")
		}
	}

	// validate the subject
	if i[schema.AttributeTypeSubject] != nil {
		if _, ok := i[schema.AttributeTypeSubject].(string); !ok {
			return perr.BadRequestWithMessage("Email attribute 'subject' must be a string")
		}
	}

	return nil
}

func (i *InputIntegrationEmail) PostMessage(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {

	input["executionID"] = i.ExecutionID
	input["pipelineExecutionID"] = i.PipelineExecutionID
	input["stepExecutionID"] = i.StepExecutionID

	// Validate the inputs
	// if err := i.ValidateInputIntegrationEmail(ctx, input); err != nil {
	// 	return nil, err
	// }

	return util.RunSendEmail(ctx, input)
}

func (*InputIntegrationEmail) ReceiveMessage(c *gin.Context) (*modconfig.Output, error) {
	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return nil, err
	}

	output := make(map[string]interface{})
	output["value"] = inputQuery.Value
	return &modconfig.Output{
		Data: output,
	}, nil
}

func (ip *Input) ValidateInput(ctx context.Context, i modconfig.Input) error {

	if i[schema.AttributeTypeType] == nil {
		return perr.BadRequestWithMessage("Slack input must define a type")
	}

	if _, ok := i[schema.AttributeTypeType].(string); !ok {
		return perr.BadRequestWithMessage("Slack input type must be a string")
	}
	inputType := i[schema.AttributeTypeType].(string)

	switch inputType {
	case string(InputTypeSlack):
		// Validate token
		if i[schema.AttributeTypeToken] == nil {
			return perr.BadRequestWithMessage("Slack input must define a token")
		}
		if _, ok := i[schema.AttributeTypeToken].(string); !ok {
			return perr.BadRequestWithMessage("Slack input token must be a string")
		}

		// Validate channel
		if i[schema.AttributeTypeChannel] == nil {
			return perr.BadRequestWithMessage("Slack input must define a channel")
		}
		if _, ok := i[schema.AttributeTypeChannel].(string); !ok {
			return perr.BadRequestWithMessage("Slack input channel must be a string")
		}

		// Validate the prompt
		if i[schema.AttributeTypePrompt] == nil {
			return perr.BadRequestWithMessage("Slack input must define a prompt")
		}
		if _, ok := i[schema.AttributeTypePrompt].(string); !ok {
			return perr.BadRequestWithMessage("Slack input prompt must be a string")
		}

		// Validate the slack type
		if i[schema.AttributeTypeSlackType] == nil {
			return perr.BadRequestWithMessage("Slack input must define a slack type")
		}
		if _, ok := i[schema.AttributeTypeSlackType].(string); !ok {
			return perr.BadRequestWithMessage("Slack input slack type must be a string")
		}

		// Validate the options
		var options []string
		if i[schema.AttributeTypeOptions] == nil {
			return perr.BadRequestWithMessage("Slack input options must define options")
		}
		if _, ok := i[schema.AttributeTypeOptions].([]string); ok {
			options = i[schema.AttributeTypeOptions].([]string)
		}
		if _, ok := i[schema.AttributeTypeOptions].([]interface{}); ok {
			for _, v := range i[schema.AttributeTypeOptions].([]interface{}) {
				options = append(options, v.(string))
			}
		}
		if len(options) == 0 {
			return perr.BadRequestWithMessage("Slack input options must have at least one option")
		}
	case string(InputTypeEmail):
	}

	return nil
}

func (ip *Input) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := ip.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// This is where the actual work is done to setup the approval stuff in slack
	inputType := input["type"].(string)

	switch inputType {
	case string(InputTypeSlack):
		slack := InputIntegrationSlack{
			InputIntegrationBase: InputIntegrationBase{
				ExecutionID:         ip.ExecutionID,
				PipelineExecutionID: ip.PipelineExecutionID,
				StepExecutionID:     ip.StepExecutionID,
			},
		}
		err := slack.PostMessage(input)
		return &modconfig.Output{}, err

	case string(InputTypeEmail):
		email := InputIntegrationEmail{
			InputIntegrationBase: InputIntegrationBase{
				ExecutionID:         ip.ExecutionID,
				PipelineExecutionID: ip.PipelineExecutionID,
				StepExecutionID:     ip.StepExecutionID,
			},
		}
		output, err := email.PostMessage(ctx, input)
		return output, err
	}
	return nil, nil
}

func (ip *Input) ProcessOutput(c *gin.Context, inputType InputType, requestBody []byte) (*modconfig.Output, error) {

	// TODO: error handling

	var output *modconfig.Output
	var err error

	switch inputType {
	case InputTypeSlack:
		slack := InputIntegrationSlack{}
		output, err = slack.ReceiveMessage(c, requestBody)
		if err != nil {
			return nil, err
		}
	case InputTypeEmail:
		email := InputIntegrationEmail{}
		output, err = email.ReceiveMessage(c)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}
