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
	"github.com/turbot/flowpipe/pipeparser/pcerr"
	"github.com/turbot/flowpipe/pipeparser/pipeline"
)

type Scheduler struct {
	ctx           context.Context
	triggers      map[string]pipeline.ITrigger
	esService     *es.ESService
	cronScheduler *gocron.Scheduler
}

func NewSchedulerService(ctx context.Context, esService *es.ESService, triggers map[string]pipeline.ITrigger) *Scheduler {
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
		switch t := t.(type) {
		case *pipeline.TriggerSchedule:
			logger.Info("Scheduling trigger", "name", t.Name, "schedule", t.Schedule)

			triggerRunner := trigger.NewTriggerRunner(s.ctx, s.esService, t)

			_, err := s.cronScheduler.Cron(t.Schedule).Do(triggerRunner.Run)
			if err != nil {
				return err
			}
		case *pipeline.TriggerInterval:
			logger.Info("Scheduling trigger", "name", t.Name, "interval", t.Schedule)

			triggerRunner := trigger.NewTriggerRunner(s.ctx, s.esService, t)

			var err error
			switch strings.ToLower(t.Schedule) {
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
				return pcerr.BadRequestWithMessage("invalid interval schedule: " + t.Schedule)
			}

			if err != nil {
				return err
			}
		}
	}

	s.cronScheduler.StartAsync()
	return nil
}
