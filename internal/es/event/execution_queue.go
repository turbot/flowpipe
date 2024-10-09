package event

import "github.com/turbot/flowpipe/internal/util"

type ExecutionQueue struct {
	Event *Event `json:"event"`
	Type  string `json:"type"`

	// TODO: make this an interface and implement JSON serialization
	PipelineQueue *PipelineQueue `json:"pipeline_queue"`
	TriggerQueue  *TriggerQueue  `json:"trigger_queue"`
}

func (e *ExecutionQueue) GetEvent() *Event {
	return e.Event
}

func (e *ExecutionQueue) HandlerName() string {
	return CommandExecutionQueue
}

func NewExecutionQueueForPipeline(executionId, pipelineName string) *ExecutionQueue {
	pipelineCmd := &PipelineQueue{
		Name:                pipelineName,
		PipelineExecutionID: util.NewPipelineExecutionId(),
	}

	if executionId == "" {
		executionId = util.NewExecutionId()
	}

	executionCmd := &ExecutionQueue{
		Event:         NewEventForExecutionID(executionId),
		PipelineQueue: pipelineCmd,
	}

	pipelineCmd.Event = executionCmd.Event

	return executionCmd
}

func NewExecutionQueueForTrigger(executionId, triggerName string) *ExecutionQueue {
	triggerCmd := &TriggerQueue{
		Name:               triggerName,
		TriggerExecutionID: util.NewTriggerExecutionId(),
	}

	if executionId == "" {
		executionId = util.NewExecutionId()
	}

	executionCmd := &ExecutionQueue{
		Event:        NewEventForExecutionID(executionId),
		TriggerQueue: triggerCmd,
	}
	triggerCmd.Event = executionCmd.Event

	return executionCmd
}
