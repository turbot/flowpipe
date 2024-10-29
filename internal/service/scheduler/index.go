package scheduler

import (
	"context"
	"github.com/turbot/flowpipe/internal/resources"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/turbot/flowpipe/internal/schedule"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/trigger"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

type SchedulerService struct {
	ctx           context.Context
	Triggers      map[string]*resources.Trigger
	esService     *es.ESService
	cronScheduler *gocron.Scheduler
}

func NewSchedulerService(ctx context.Context, esService *es.ESService, triggers map[string]*resources.Trigger) *SchedulerService {
	return &SchedulerService{
		ctx:       ctx,
		esService: esService,
		Triggers:  triggers,
	}
}

func (s *SchedulerService) RescheduleTriggers() error {
	if s.cronScheduler == nil {
		return nil
	}

	validJobsNames := []string{}

	for _, t := range s.Triggers {
		var scheduleString string
		switch config := t.Config.(type) {
		case *resources.TriggerSchedule:
			scheduleString = config.Schedule
		case *resources.TriggerQuery:
			scheduleString = config.Schedule
			if scheduleString == "" {
				scheduleString = "hourly"
			}
		case *resources.TriggerHttp:
			continue
		}

		if t.Enabled != nil && !*t.Enabled {
			// if trigger is disabled, skip the scheduling logic, do not add to the validJobNames list
			// it will be removed below
			continue
		}

		validJobsNames = append(validJobsNames, "id:"+t.FullName)

		// Find the job in the scheduler
		jobs, err := s.cronScheduler.FindJobsByTag("id:" + t.FullName)
		if err != nil && err == gocron.ErrJobNotFoundWithTag {
			// Job not found in the scheduler, schedule it
			err := s.scheduleTrigger(t)
			if err != nil {
				return err
			}
			continue
		} else if err != nil {
			// Real error, return the error
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

		// Detect changes, only changes in the schedule should result in a re-schedule. Changes in the trigger config itself,
		// i.e. pipeline changes don't need a re-schedule. We trigger config is not stored in the scheduler, when mod is updated
		// the cache is updated and the definition is retrieved again when we run the trigger.
		if jobTags[1] != "schedule:"+scheduleString {
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
		if !strings.HasPrefix(jobTags[0], "id:") {
			continue
		}

		if !slices.Contains[[]string, string](validJobsNames, jobTags[0]) {
			slog.Info("Removing trigger", "name", jobTags[0])
			s.cronScheduler.RemoveByReference(job)
		}
	}

	return nil
}

func (s *SchedulerService) scheduleTrigger(t *resources.Trigger) error {

	scheduleString := ""

	switch config := t.Config.(type) {
	case *resources.TriggerSchedule:
		scheduleString = config.Schedule
	case *resources.TriggerQuery:
		scheduleString = config.Schedule
		if scheduleString == "" {
			scheduleString = "hourly"
		}
	default:
		// can't schedule HTTP Trigger
		return nil
	}

	if t.Enabled != nil && !*t.Enabled {
		slog.Debug("Trigger is disabled", "name", t.Name())
		return nil
	}

	tags := []string{
		"id:" + t.FullName,
		"schedule:" + scheduleString,
	}

	pipelineName := ""
	if t.Pipeline != cty.NilVal {
		pipelineValueMap := t.Pipeline.AsValueMap()
		if pipelineValueMap == nil {
			return perr.BadRequestWithMessage("pipeline not found for trigger " + t.Name())
		}

		if pipelineValueMap[schema.LabelName] == cty.NilVal {
			return perr.BadRequestWithMessage("pipeline name not found for trigger " + t.Name())
		}

		pipelineName = t.Pipeline.AsValueMap()[schema.LabelName].AsString()
	}

	if pipelineName != "" {
		tags = append(tags, "pipeline:"+pipelineName)
	}

	triggerRunner := trigger.NewTriggerRunner(t, "", "")

	scheduledTriggerRunner := TriggerScheduleRunner{
		TriggerRunner: triggerRunner,
		CommandBus:    s.esService.CommandBus,
	}

	_, err := s.cronScheduler.Cron(scheduleString).Tag(tags...).Do(scheduledTriggerRunner.Run)
	if err != nil {
		cronExpression, err := schedule.IntervalToCronExpression(t.FullName, scheduleString)
		if err != nil {
			return err
		}

		slog.Info("Scheduling trigger", "name", t.Name(), "schedule", scheduleString, "tags", tags, "cronExpression", cronExpression)
		_, err = s.cronScheduler.Cron(cronExpression).Tag(tags...).Do(scheduledTriggerRunner.Run)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchedulerService) Start() error {
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

func (s SchedulerService) ScheduleCoreServices() error {

	currentTime := time.Now()

	// Add 1 minute to the current time
	oneMinuteLater := currentTime.Add(time.Minute * 2)

	minute := oneMinuteLater.Minute()
	scheduleString := strconv.Itoa(minute) + " * * * *"

	tags := []string{
		"core-services",
		"flowpipe-db-cleanup",
	}

	slog.Info("Scheduling flowpipe db cleanup", "schedule", scheduleString, "tags", tags)

	_, err := s.cronScheduler.Cron(scheduleString).Tag(tags...).Do(store.CleanupRunner)
	if err != nil {
		slog.Error("Error scheduling flowpipe db cleanup", "error", err)
		return perr.InternalWithMessage("error scheduling flowpipe db cleanup")
	}

	return nil
}
