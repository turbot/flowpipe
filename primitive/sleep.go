package primitive

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Sleep struct{}

func (e *Sleep) ValidateInput(ctx context.Context, input Input) error {

	if input["duration"] == nil {
		return errors.New("Sleep input must define a duration")
	}

	durationString := input["duration"].(string)
	_, err := time.ParseDuration(durationString)
	if err != nil {
		return fmt.Errorf("Invalid sleep duration: %s", durationString)
	}

	return nil
}

func (e *Sleep) Run(ctx context.Context, input Input) (Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	durationString := input["duration"].(string)
	// Already validated
	duration, _ := time.ParseDuration(durationString)

	fmt.Println("Sleeping for ", duration, "...")
	time.Sleep(duration)

	return Output{}, nil
}
