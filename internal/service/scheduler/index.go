package scheduler

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	estrigger "github.com/turbot/flowpipe/internal/es/trigger"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
)

type Scheduler struct {
	ctx       context.Context
	triggers  map[string]types.ITrigger
	esService *es.ESService
}

func NewSchedulerService(ctx context.Context, esService *es.ESService, triggers map[string]types.ITrigger) *Scheduler {
	return &Scheduler{
		ctx:       ctx,
		esService: esService,
		triggers:  triggers,
	}
}

func (s *Scheduler) Start() error {

	logger := fplog.Logger(s.ctx)

	if len(s.triggers) == 0 {
		return nil
	}

	cronScheduler := gocron.NewScheduler(time.UTC)

	for _, t := range s.triggers {
		switch t := t.(type) {
		case *types.TriggerSchedule:
			logger.Info("Scheduling trigger", "name", t.Name, "schedule", t.Schedule)

			triggerRunner := estrigger.NewTriggerRunner(s.ctx, s.esService, t)
			_, err := cronScheduler.Cron(t.Schedule).Do(triggerRunner.Run)
			if err != nil {
				return err
			}
		}
	}

	cronScheduler.StartAsync()
	return nil
}
