package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/handler"
	"github.com/turbot/steampipe-pipelines/fplog"
	"github.com/turbot/steampipe-pipelines/utils"
	"go.uber.org/zap"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	"github.com/garsue/watermillzap"
)

func main() {

	ctx := context.Background()
	ctx = utils.ContextWithSession(ctx)
	ctx = fplog.ContextWithLogger(ctx)

	logger := watermillzap.NewLogger(fplog.Logger(ctx))

	cqrsMarshaler := cqrs.JSONMarshaler{}

	commandsPubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
	eventsPubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

	// CQRS is built on messages router. Detailed documentation: https://watermill.io/docs/messages-router/
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	// Simple middleware which will recover panics from event or command handlers.
	// More about router middlewares you can find in the documentation:
	// https://watermill.io/docs/messages-router/#middleware
	//
	// List of available middlewares you can find in message/router/middleware.
	//router.AddMiddleware(middleware.RandomFail(0.5))
	//router.AddMiddleware(middleware.Recoverer)

	// Log to file for creation of state
	router.AddMiddleware(LogEventMiddlewareWithContext(ctx))

	// cqrs.Facade is facade for Command and Event buses and processors.
	// You can use facade, or create buses and processors manually (you can inspire with cqrs.NewFacade)
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			// we are using queue RabbitMQ config, so we need to have topic per command type
			return commandName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.QueueHandler{EventBus: eb},
				command.LoadHandler{EventBus: eb},
				command.StartHandler{EventBus: eb},
				command.PlanHandler{EventBus: eb},
				command.PipelineQueueHandler{EventBus: eb},
				command.PipelineLoadHandler{EventBus: eb},
				command.PipelineStartHandler{EventBus: eb},
				command.PipelinePlanHandler{EventBus: eb},
				command.PipelineFinishHandler{EventBus: eb},
				command.PipelineStepExecuteHandler{EventBus: eb},
				command.StopHandler{EventBus: eb},
			}
		},
		CommandsPublisher: commandsPubSub,
		CommandsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return commandsPubSub, nil
		},
		GenerateEventsTopic: func(eventName string) string {
			return eventName
		},
		EventHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				handler.Queued{CommandBus: cb},
				handler.Loaded{CommandBus: cb},
				handler.Started{CommandBus: cb},
				handler.Planned{CommandBus: cb},
				handler.PipelineQueued{CommandBus: cb},
				handler.PipelineLoaded{CommandBus: cb},
				handler.PipelineStarted{CommandBus: cb},
				handler.PipelinePlanned{CommandBus: cb},
				handler.PipelineFinished{CommandBus: cb},
				handler.PipelineStepExecuted{CommandBus: cb},
				handler.Failed{CommandBus: cb},
				handler.Stopped{CommandBus: cb},
			}
		},
		EventsPublisher: eventsPubSub,
		EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return eventsPubSub, nil
		},
		/*
			EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
				config := amqp.NewDurablePubSubConfig(
					amqpAddress,
					amqp.GenerateQueueNameTopicNameWithSuffix(handlerName),
				)
				return amqp.NewSubscriber(config, logger)
			},
		*/
		Router:                router,
		CommandEventMarshaler: cqrsMarshaler,
		Logger:                logger,
	})
	if err != nil {
		panic(err)
	}

	runID := xid.New().String()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd := &event.Stop{
			RunID:     runID,
			SpanID:    runID,
			CreatedAt: time.Now().UTC(),
		}
		if err := cqrsFacade.CommandBus().Send(ctx, cmd); err != nil {
			panic(err)
		}
	}()

	// publish commands every second to simulate incoming traffic
	go publishCommands(ctx, runID, cqrsFacade.CommandBus())

	// processors are based on router, so they will work when router will start
	if err := router.Run(ctx); err != nil {
		panic(err)
	}
}

func publishCommands(ctx context.Context, modSpanID string, commandBus *cqrs.CommandBus) {

	// Initialize the mod
	cmd := &event.Queue{
		Workspace: "e-gineer/scratch",
		SpanID:    modSpanID,
		CreatedAt: time.Now().UTC(),
	}
	if err := commandBus.Send(ctx, cmd); err != nil {
		panic(err)
	}

	// Manually trigger some pipelines for testing
	// TODO - these should be triggered instead (e.g. cron, webhook, etc)
	for _, i := range []int{1, 2, 3} {
		time.Sleep(3 * time.Second)
		fmt.Println()
		spanID := xid.New().String()
		cmd := &event.PipelineQueue{
			RunID:     modSpanID,
			SpanID:    spanID,
			CreatedAt: time.Now().UTC(),
			//StackID:      e.StackID,
			Name: fmt.Sprintf("my_pipeline_%d", i%2),
			//Input:        e.Input,
		}
		if err := commandBus.Send(ctx, cmd); err != nil {
			panic(err)
		}
	}
}

type PipelinePayload struct {
	RunID  string `json:"run_id"`
	SpanID string `json:"span_id"`
}

func LogEventMiddlewareWithContext(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {

			var payload map[string]interface{}
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				panic("TODO - invalid log payload, log me?")
			}

			stackID := ""
			if payload["run_id"] != nil {
				if runID, ok := payload["run_id"].(string); ok {
					stackID = runID
				}
			}
			if stackID == "" {
				if payload["span_id"] != nil {
					if spanID, ok := payload["span_id"].(string); ok {
						stackID = spanID
					}
				}
			}
			if stackID == "" {
				panic(fmt.Sprintf("TODO - No run_id or span_id found in payload: %s\n", msg.Payload))
			}

			logger := fplog.Logger(ctx)
			defer logger.Sync()
			logger.Info("es",
				// Structured context as strongly typed Field values.
				zap.String("event_type", msg.Metadata["name"]),
				// zap adds ts field automatically, so don't need zap.Time("created_at", time.Now()),
				zap.Any("payload", payload),
			)

			return h(msg)

		}
	}
}
