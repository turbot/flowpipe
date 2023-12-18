package scheduler

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type SchedulerService struct {
	ctx           context.Context
	Triggers      map[string]*modconfig.Trigger
	esService     *es.ESService
	cronScheduler *gocron.Scheduler
}

func NewSchedulerService(ctx context.Context, esService *es.ESService, triggers map[string]*modconfig.Trigger) *SchedulerService {
	return &SchedulerService{
		ctx:       ctx,
		esService: esService,
		Triggers:  triggers,
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

func (s *SchedulerService) RescheduleTriggers() error {
	if s.cronScheduler == nil {
		return nil
	}

	validJobsNames := []string{}

	for _, t := range s.Triggers {
		var scheduleString string
		switch config := t.Config.(type) {
		case *modconfig.TriggerSchedule:
			scheduleString = config.Schedule
		case *modconfig.TriggerInterval:
			scheduleString = config.Schedule
		}

		if scheduleString == "" {
			continue
		}

		validJobsNames = append(validJobsNames, "id:"+t.FullName)

		// Find the job in the scheduler
		jobs, err := s.cronScheduler.FindJobsByTag("id:" + t.FullName)
		if err != nil && err == gocron.ErrJobNotFoundWithTag {
			err := s.scheduleTrigger(t)
			if err != nil {
				return err
			}
			continue
		} else if err != nil {
			return err
		}

		if len(jobs) > 1 {
			return perr.ConflictWithMessage("multiple jobs found for trigger: " + t.FullName)
		}

		if len(jobs) == 0 {
			err := s.scheduleTrigger(t)
			if err != nil {
				return err
			}
			continue
		}

		job := jobs[0]
		jobTags := job.Tags()

		if jobTags[1] != "schedule:"+scheduleString {
			slog.Info("Rescheduling trigger", "name", t.Name(), "schedule", scheduleString)
			s.cronScheduler.RemoveByReference(job)
			err := s.scheduleTrigger(t)
			if err != nil {
				return err
			}
			continue
		}

		if jobTags[2] != "pipeline:"+t.Pipeline.AsValueMap()[schema.LabelName].AsString() {
			slog.Info("Rescheduling trigger", "name", t.Name(), "schedule", scheduleString)
			s.cronScheduler.RemoveByReference(job)
			err := s.scheduleTrigger(t)
			if err != nil {
				return err
			}
			continue
		}
	}

	// now loop through all the jobs in the scheduler and remove any that are not in the valid list
	allJobs := s.cronScheduler.Jobs()
	for _, job := range allJobs {
		jobTags := job.Tags()
		if len(jobTags) != 3 || !strings.HasPrefix(jobTags[0], "id:") {
			continue
		}

		if !slices.Contains[[]string, string](validJobsNames, jobTags[0]) {
			slog.Info("Removing trigger", "name", jobTags[0])
			s.cronScheduler.RemoveByReference(job)
		}
	}

	return nil
}

func (s *SchedulerService) scheduleTrigger(t *modconfig.Trigger) error {
	pipelineValueMap := t.Pipeline.AsValueMap()
	if pipelineValueMap == nil {
		return perr.BadRequestWithMessage("pipeline not found for trigger " + t.Name())
	}

	if pipelineValueMap[schema.LabelName] == cty.NilVal {
		return perr.BadRequestWithMessage("pipeline name not found for trigger " + t.Name())
	}

	pipelineName := t.Pipeline.AsValueMap()[schema.LabelName].AsString()
	switch config := t.Config.(type) {
	case *modconfig.TriggerSchedule:

		tags := []string{
			"id:" + t.FullName,
			"schedule:" + config.Schedule,
			"pipeline:" + pipelineName,
		}

		slog.Info("Scheduling trigger", "name", t.Name(), "schedule", config.Schedule, "tags", tags)

		triggerRunner := trigger.NewTriggerRunner(s.ctx, s.esService.CommandBus, s.esService.RootMod, t)
		_, err := s.cronScheduler.Cron(config.Schedule).Tag(tags...).Do(triggerRunner.Run)
		if err != nil {
			return err
		}

	case *modconfig.TriggerInterval:
		tags := []string{
			"id:" + t.FullName,
			"schedule:" + config.Schedule,
			"pipeline:" + pipelineName,
		}

		slog.Info("Scheduling trigger", "name", t.Name(), "interval", config.Schedule)

		triggerRunner := trigger.NewTriggerRunner(s.ctx, s.esService.CommandBus, s.esService.RootMod, t)

		var err error
		switch strings.ToLower(config.Schedule) {
		case "hourly":
			ts := randomizeTimestamp(0.0, 0.1, time.Now().UTC(), 1*time.Hour)
			_, err = s.cronScheduler.Every(1).Hour().StartAt(ts).Tag(tags...).Do(triggerRunner.Run)
		case "daily":
			ts := randomizeTimestamp(0.1, 0.5, time.Now().UTC(), 1*time.Hour)
			_, err = s.cronScheduler.Every(1).Day().StartAt(ts).Tag(tags...).Do(triggerRunner.Run)
		case "weekly":
			ts := randomizeTimestamp(0.2, 1.0, time.Now().UTC(), 1*time.Hour)
			_, err = s.cronScheduler.Every(1).Week().StartAt(ts).Tag(tags...).Do(triggerRunner.Run)
		case "monthly":
			ts := randomizeTimestamp(0.2, 1.0, time.Now().UTC(), 1*time.Hour)
			_, err = s.cronScheduler.Every(1).Month().StartAt(ts).Tag(tags...).Do(triggerRunner.Run)
		default:
			return perr.BadRequestWithMessage("invalid interval schedule: " + config.Schedule)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedulerService) Start() error {

	if len(s.Triggers) == 0 {
		return nil
	}

	s.cronScheduler = gocron.NewScheduler(time.UTC)

	for _, t := range s.Triggers {
		err := s.scheduleTrigger(t)
		if err != nil {
			return err
		}
	}

	s.cronScheduler.StartAsync()
	return nil
}
