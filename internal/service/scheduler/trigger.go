package scheduler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/trigger"
)

type TriggerScheduleRunner struct {
	TriggerRunner trigger.TriggerRunner
	CommandBus    handler.FpCommandBus
}

func (s *TriggerScheduleRunner) Run() {
	triggerName := s.TriggerRunner.GetTrigger().Name()

	executionCmd := event.NewExecutionQueueForTrigger("", triggerName)

	// Send the trigger command
	err := s.CommandBus.Send(context.TODO(), executionCmd)
	if err != nil {
		slog.Error("Error sending trigger command", "trigger", triggerName, "error", err)
	}
}
