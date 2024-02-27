package primitive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

type Message struct {
	Input
}

func NewMessagePrimitive(executionId, pipelineExecutionId, stepExecutionId, pipelineName, stepName string) *Message {
	return &Message{
		Input: *NewInputPrimitive(executionId, pipelineExecutionId, stepExecutionId, pipelineName, stepName),
	}
}

func (mp *Message) ValidateInput(ctx context.Context, input modconfig.Input) error {
	return nil
}

func (mp *Message) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	err := mp.ValidateInput(ctx, input)
	if err != nil {
		return nil, err
	}

	var body string

	if b, ok := input[schema.AttributeTypeBody].(string); ok {
		body = b
	}

	return mp.Input.execute(ctx, input, &MessageStepMessageCreator{
		Body: body,
	})
}

type MessageStepMessageCreator struct {
	Body string
}

func (icm *MessageStepMessageCreator) SlackMessage(ip *InputIntegrationSlack, options []InputIntegrationResponseOption) (slack.Blocks, error) {
	var blocks slack.Blocks

	// payload for callback
	payload := map[string]any{
		"execution_id":          ip.ExecutionID,
		"pipeline_execution_id": ip.PipelineExecutionID,
		"step_execution_id":     ip.StepExecutionID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return blocks, err
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonPayload)

	boldPromptBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("*%s*", icm.Body), false, false)

	header := slack.NewSectionBlock(boldPromptBlock, nil, nil, slack.SectionBlockOptionBlockID(encodedPayload))
	blocks.BlockSet = append(blocks.BlockSet, header)

	return blocks, nil
}

func (icm *MessageStepMessageCreator) EmailMessage(iim *InputIntegrationEmail) (string, error) {

	header := make(map[string]string)
	header["From"] = iim.From
	header["To"] = strings.Join(iim.To, ", ")
	header["Subject"] = iim.Subject
	header["Content-Type"] = "text/html; charset=\"UTF-8\";"
	header["MIME-version"] = "1.0;"

	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	data := struct {
		Prompt string
	}{
		Prompt: icm.Body,
	}

	templateMessage, err := parseEmailInputTemplate("message-basic.html", data)
	if err != nil {
		return "", err
	}
	message += templateMessage

	return message, nil

}
