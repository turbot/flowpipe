package scheduler

import (
	"context"
	"log/slog"

	"github.com/turbot/flowpipe/internal/trigger"
)

type TriggerScheduleRunner struct {
	TriggerRunner trigger.TriggerRunner
}

func (s *TriggerScheduleRunner) Run() {
	_, err := s.TriggerRunner.ExecuteTriggerWithArgs(context.Background(), nil, nil)
	if err != nil {
		slog.Error("Error executing trigger", "error", err)
	}
}
