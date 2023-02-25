package command

import (
	"context"
	"fmt"
	"time"

	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/state"
)

type PipelinePlanHandler CommandHandler

func (h PipelinePlanHandler) HandlerName() string {
	return "command.pipeline_plan"
}

func (h PipelinePlanHandler) NewCommand() interface{} {
	return &event.PipelinePlan{}
}

func (h PipelinePlanHandler) Handle(ctx context.Context, c interface{}) error {

	cmd := c.(*event.PipelinePlan)

	fmt.Printf("[%-20s] %v\n", h.HandlerName(), cmd)

	s, err := state.NewState(ctx, cmd.RunID)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	defn, err := PipelineDefinition(s.PipelineName)
	if err != nil {
		// TODO - should this return a failed event? how are errors caught here?
		return err
	}

	e := event.PipelinePlanned{
		RunID:           cmd.RunID,
		SpanID:          cmd.SpanID,
		CreatedAt:       time.Now().UTC(),
		NextStepIndexes: []int{},
	}

	for i, _ := range defn.Steps {
		if s.PipelineStepStatus[i] == "" {
			e.NextStepIndexes = append(e.NextStepIndexes, i)
		}
	}

	return h.EventBus.Publish(ctx, &e)
}
