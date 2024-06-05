package event

import (
	"time"

	"github.com/turbot/flowpipe/internal/util"
)

// All events have a shared structure to track execution context.
type Event struct {
	// Every execution has a unique ID. This is used right through from initial
	// triggering down through all pipelines, steps and nested pipelines.
	ExecutionID string `json:"execution_id"`
	// Time when the command was created.
	CreatedAt time.Time `json:"created_at"`
}

type PayloadWithEvent struct {
	Event *Event `json:"event"`
}

func NewExecutionEvent() *Event {
	return NewEventForExecutionID(util.NewExecutionId())
}

func NewEventForExecutionID(executionID string) *Event {
	if executionID == "" {
		executionID = util.NewExecutionId()
	}

	return &Event{
		ExecutionID: executionID,
		CreatedAt:   time.Now().UTC(),
	}
}

func NewChildEvent(parent *Event) *Event {
	return &Event{
		ExecutionID: parent.ExecutionID,
		CreatedAt:   time.Now().UTC(),
	}
}

func NewFlowEvent(before *Event) *Event {
	return &Event{
		ExecutionID: before.ExecutionID,
		CreatedAt:   time.Now().UTC(),
	}
}

func NewParentEvent(child *Event) *Event {
	return &Event{
		ExecutionID: child.ExecutionID,
		CreatedAt:   time.Now().UTC(),
	}
}

type CommandEvent interface {
	GetEvent() *Event
	HandlerName() string
}
