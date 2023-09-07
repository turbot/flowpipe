package scheduler

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
)

type Scheduler struct {
	ctx           context.Context
	triggers      map[string]*modconfig.Trigger
	esService     *es.ESService
	cronScheduler *gocron.Scheduler
}

func NewSchedulerService(ctx context.Context, esService *es.ESService, triggers map[string]*modconfig.Trigger) *Scheduler {
	return &Scheduler{
		ctx:       ctx,
		esService: esService,
		triggers:  triggers,
	}
}

func randomizeTimestamp(start, end float64, baseTime time.Time, interval time.Duration) time.Time {
	rangeStart := int64(interval.Seconds() * start)
	rangeEnd := int64(interval.Seconds() * end)

	// Generate a random offset within the range
	n, _ := rand.Int(rand.Reader, big.NewInt(rangeEnd-rangeStart))

	randomOffset := time.Duration((n.Int64() + rangeStart) * int64(time.Second))

	// Create the randomized timestamp
	randomTimestamp := baseTime.Add(randomOffset)

	return randomTimestamp
}

func (s *Scheduler) Start() error {

	logger := fplog.Logger(s.ctx)

	if len(s.triggers) == 0 {
		return nil
	}

	s.cronScheduler = gocron.NewScheduler(time.UTC)

	for _, t := range s.triggers {
		switch config := t.Config.(type) {
		case *modconfig.TriggerSchedule:
			logger.Info("Scheduling trigger", "name", t.Name(), "schedule", config.Schedule)

			triggerRunner := trigger.NewTriggerRunner(s.ctx, s.esService, t)

			_, err := s.cronScheduler.Cron(config.Schedule).Do(triggerRunner.Run)
			if err != nil {
				return err
			}
		case *modconfig.TriggerInterval:
			logger.Info("Scheduling trigger", "name", t.Name(), "interval", config.Schedule)

			triggerRunner := trigger.NewTriggerRunner(s.ctx, s.esService, t)

			var err error
			switch strings.ToLower(config.Schedule) {
			case "hourly":
				ts := randomizeTimestamp(0.0, 0.1, time.Now().UTC(), 1*time.Hour)
				_, err = s.cronScheduler.Every(1).Hour().StartAt(ts).Do(triggerRunner.Run)
			case "daily":
				ts := randomizeTimestamp(0.1, 0.5, time.Now().UTC(), 1*time.Hour)
				_, err = s.cronScheduler.Every(1).Day().StartAt(ts).Do(triggerRunner.Run)
			case "weekly":
				ts := randomizeTimestamp(0.2, 1.0, time.Now().UTC(), 1*time.Hour)
				_, err = s.cronScheduler.Every(1).Week().StartAt(ts).Do(triggerRunner.Run)
			case "monthly":
				ts := randomizeTimestamp(0.2, 1.0, time.Now().UTC(), 1*time.Hour)
				_, err = s.cronScheduler.Every(1).Month().StartAt(ts).Do(triggerRunner.Run)
			default:
				return perr.BadRequestWithMessage("invalid interval schedule: " + config.Schedule)
			}

			if err != nil {
				return err
			}
		}
	}

	s.cronScheduler.StartAsync()
	return nil
}
