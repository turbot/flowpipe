package primitive

import (
	"context"

	"github.com/turbot/pipe-fittings/modconfig"
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
	return nil, nil
}
