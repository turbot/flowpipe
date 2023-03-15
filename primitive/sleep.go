package primitive

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type Sleep struct{}

func (e *Sleep) ValidateInput(ctx context.Context, input pipeline.StepInput) error {

	if input["duration"] == nil {
		return errors.New("Sleep input must define a duration")
	}

	durationString := input["duration"].(string)
	_, err := time.ParseDuration(durationString)
	if err != nil {
		return fmt.Errorf("invalid sleep duration: %s", durationString)
	}

	return nil
}

func (e *Sleep) Run(ctx context.Context, input pipeline.StepInput) (*pipeline.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	durationString := input["duration"].(string)
	// Already validated
	duration, _ := time.ParseDuration(durationString)

	fmt.Println("Sleeping for ", duration, "...")
	start := time.Now().UTC()
	time.Sleep(duration)
	finish := time.Now().UTC()

	return &pipeline.Output{"started_at": start, "finished_at": finish}, nil
}
