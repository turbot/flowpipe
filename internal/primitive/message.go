package primitive

import (
	"context"
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
	return mp.Input.validateInputNotifier(input)
}

func (mp *Message) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	err := mp.ValidateInput(ctx, input)
	if err != nil {
		return nil, err
	}

	var text string

	if b, ok := input[schema.AttributeTypeText].(string); ok {
		text = b
	}

	return mp.Input.execute(ctx, input, &MessageStepMessageCreator{
		Text: text,
	})
}

type MessageStepMessageCreator struct {
	Text string
}

func (icm *MessageStepMessageCreator) SlackMessage(ip *InputIntegrationSlack, options []InputIntegrationResponseOption) (slack.Blocks, error) {
	var blocks slack.Blocks

	promptBlock := slack.NewTextBlockObject(slack.PlainTextType, icm.Text, false, false)

	header := slack.NewSectionBlock(promptBlock, nil, nil)
	blocks.BlockSet = append(blocks.BlockSet, header)

	return blocks, nil
}

func (icm *MessageStepMessageCreator) EmailMessage(iim *InputIntegrationEmail, _ []InputIntegrationResponseOption) (string, error) {

	header := make(map[string]string)
	header["From"] = iim.From
	header["To"] = strings.Join(iim.To, ", ")

	header["Content-Type"] = "text/html; charset=\"UTF-8\";"
	header["MIME-version"] = "1.0;"

	subject := iim.Subject

	if subject == "" {
		subject = icm.Text
		if len(subject) > 50 {
			subject = subject[:50] + "..."
		}
	}

	header["Subject"] = subject

	var message string
	for key, value := range header {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	data := struct {
		Prompt string
	}{
		Prompt: icm.Text,
	}

	templateMessage, err := parseEmailInputTemplate("message-basic.html", data)
	if err != nil {
		return "", err
	}
	message += templateMessage

	return message, nil

}
