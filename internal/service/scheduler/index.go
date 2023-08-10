package scheduler

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/turbot/flowpipe/fperr"
	estrigger "github.com/turbot/flowpipe/internal/es/trigger"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/types"
)

type Scheduler struct {
	ctx           context.Context
	triggers      map[string]types.ITrigger
	esService     *es.ESService
	cronScheduler *gocron.Scheduler
}

func NewSchedulerService(ctx context.Context, esService *es.ESService, triggers map[string]types.ITrigger) *Scheduler {
	return &Scheduler{
		ctx:       ctx,
		esService: esService,
		triggers:  triggers,
	}
}

func randomizeTimestamp(baseTime time.Time, interval time.Duration) time.Time {
	// Calculate the range for randomization (0-10% of the interval)
	rangeStart := int64(interval.Seconds() * 0.0)
	rangeEnd := int64(interval.Seconds() * 0.1)

	// Generate a random offset within the range
	randomOffset := time.Duration(rand.Int63n(rangeEnd-rangeStart) + rangeStart)

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
		case *types.TriggerSchedule:
			logger.Info("Scheduling trigger", "name", t.Name, "schedule", t.Schedule)

			triggerRunner := estrigger.NewTriggerRunner(s.ctx, s.esService, t)
			_, err := s.cronScheduler.Cron(t.Schedule).Do(triggerRunner.Run)
			if err != nil {
				return err
			}
		case *types.TriggerInterval:
			logger.Info("Scheduling trigger", "name", t.Name, "interval", t.Schedule)

			triggerRunner := estrigger.NewTriggerRunner(s.ctx, s.esService, t)

			var err error
			switch strings.ToLower(t.Schedule) {
			case "hourly":
				ts := randomizeTimestamp(time.Now().UTC(), 1*time.Hour)
				_, err = s.cronScheduler.Every(1).Hour().StartAt(ts).Do(triggerRunner.Run)
			case "daily":
				ts := randomizeTimestamp(time.Now().UTC(), 24*time.Hour)
				_, err = s.cronScheduler.Every(1).Day().StartAt(ts).Do(triggerRunner.Run)
			case "weekly":
				ts := randomizeTimestamp(time.Now().UTC(), 7*24*time.Hour)
				_, err = s.cronScheduler.Every(1).Week().StartAt(ts).Do(triggerRunner.Run)
			case "monthly":
				ts := randomizeTimestamp(time.Now().UTC(), 30*24*time.Hour)
				_, err = s.cronScheduler.Every(1).Month().StartAt(ts).Do(triggerRunner.Run)
			default:
				return fperr.BadRequestWithMessage("invalid interval schedule: " + t.Schedule)
			}

			if err != nil {
				return err
			}
		}
	}

	s.cronScheduler.StartAsync()
	return nil
}
