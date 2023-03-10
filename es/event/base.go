package event

import (
	"context"
	"strings"
	"time"

	"github.com/turbot/steampipe-pipelines/utils"
)

// All events have a shared structure to track execution context.
type Event struct {
	// Every execution has a unique ID. This is used right through from initial
	// triggering down through all pipelines, steps and nested pipelines.
	ExecutionID string `json:"execution_id"`
	// A given execution may initiate many different pipelines and steps. Each
	// pipeline and step is given a unique StackID. The StackID is nested according
	// to the hierarchy of calls:
	//   Pipeline A (StackID: db23)
	//   	Step 1 (StackID: db23.1a2b)
	//   	Step 2 (StackID: db23.b4cd)
	//   	  Pipeline B (StackID: db23.b4cd.e6ab)
	//   	  	Step 1 (StackID: db23.b4cd.e6ab.ac23)
	//   	Step 3 (StackID: db23.98b3)
	StackIDs []string `json:"stack_ids"`
	// Time when the command was created.
	CreatedAt time.Time `json:"created_at"`
}

type PayloadWithEvent struct {
	Event *Event `json:"event"`
}

func NewExecutionEvent(ctx context.Context) *Event {
	return &Event{
		ExecutionID: utils.Session(ctx),
		StackIDs:    []string{},
		CreatedAt:   time.Now().UTC(),
	}
}

func NewChildEvent(parent *Event) *Event {
	return &Event{
		ExecutionID: parent.ExecutionID,
		StackIDs:    append(parent.StackIDs, utils.NewUniqueID()),
		CreatedAt:   time.Now().UTC(),
	}
}

func NewFlowEvent(before *Event) *Event {
	return &Event{
		ExecutionID: before.ExecutionID,
		StackIDs:    before.StackIDs,
		CreatedAt:   time.Now().UTC(),
	}
}

func NewParentEvent(child *Event) *Event {
	return &Event{
		ExecutionID: child.ExecutionID,
		StackIDs:    child.StackIDs[:len(child.StackIDs)-1],
		CreatedAt:   time.Now().UTC(),
	}
}

func (e *Event) StackID() string {
	return strings.Join(e.StackIDs, ".")
}

func (e *Event) LastStackID() string {
	if len(e.StackIDs) == 0 {
		return ""
	}
	if len(e.StackIDs) == 1 {
		return e.StackIDs[0]
	}
	return e.StackIDs[len(e.StackIDs)-2]
}

func (e *Event) ShortExecutionID() string {
	return e.ExecutionID[len(e.ExecutionID)-4:]
}

func (e *Event) ShortExecutionStackID() string {
	ssID := e.ShortStackID()
	if ssID == "" {
		return e.ShortExecutionID()
	}
	return e.ShortExecutionID() + "." + ssID
}

func (e *Event) ShortStackID() string {
	shortIDs := make([]string, len(e.StackIDs))
	for i, id := range e.StackIDs {
		shortIDs[i] = id[len(id)-4:]
	}
	return strings.Join(shortIDs, ".")
}
