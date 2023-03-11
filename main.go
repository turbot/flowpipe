package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turbot/steampipe-pipelines/config"
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

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	ctx = config.Set(ctx, cfg)

	logger := watermillzap.NewLogger(fplog.Logger(ctx))

	cqrsMarshaler := cqrs.JSONMarshaler{}

	goChannelConfig := gochannel.Config{
		// TODO - I really don't understand this and I'm not sure it's necessary.
		//OutputChannelBuffer: 10000,
		//Persistent:          true,
	}
	commandsPubSub := gochannel.NewGoChannel(goChannelConfig, logger)
	eventsPubSub := gochannel.NewGoChannel(goChannelConfig, logger)

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

	// Dump the state of the event sourcing log with every event
	//router.AddMiddleware(DumpState(ctx))

	// Throttle, if required
	//router.AddMiddleware(middleware.NewThrottle(4, time.Second).Middleware)

	// Retry, if required
	/*
		retry := middleware.Retry{
			MaxRetries: 3,
		}
		router.AddMiddleware(retry.Middleware)
	*/

	// cqrs.Facade is facade for Command and Event buses and processors.
	// You can use facade, or create buses and processors manually (you can inspire with cqrs.NewFacade)
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			// we are using queue RabbitMQ config, so we need to have topic per command type
			return commandName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.LoadHandler{EventBus: eb},
				command.PipelineFinishHandler{EventBus: eb},
				command.PipelineLoadHandler{EventBus: eb},
				command.PipelinePlanHandler{EventBus: eb},
				command.PipelineQueueHandler{EventBus: eb},
				command.PipelineStartHandler{EventBus: eb},
				command.PipelineStepFinishHandler{EventBus: eb},
				command.PipelineStepStartHandler{EventBus: eb},
				command.QueueHandler{EventBus: eb},
				command.StartHandler{EventBus: eb},
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
				handler.Failed{CommandBus: cb},
				handler.Loaded{CommandBus: cb},
				handler.PipelineFailed{CommandBus: cb},
				handler.PipelineFinished{CommandBus: cb},
				handler.PipelineLoaded{CommandBus: cb},
				handler.PipelinePlanned{CommandBus: cb},
				handler.PipelineQueued{CommandBus: cb},
				handler.PipelineStarted{CommandBus: cb},
				handler.PipelineStepFinished{CommandBus: cb},
				handler.PipelineStepStarted{CommandBus: cb},
				handler.Queued{CommandBus: cb},
				handler.Started{CommandBus: cb},
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

	runID := utils.NewProcessID()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
		/*
			// Graceful exit through stop, but this makes the event log
			// seem like the process is complete.
			cmd := &event.Stop{
				RunID:     runID,
				SpanID:    runID,
				CreatedAt: time.Now().UTC(),
			}
			if err := cqrsFacade.CommandBus().Send(ctx, cmd); err != nil {
				panic(err)
			}
		*/
	}()

	// publish commands every second to simulate incoming traffic
	go publishCommands(ctx, runID, cqrsFacade.CommandBus())

	// processors are based on router, so they will work when router will start
	if err := router.Run(ctx); err != nil {
		panic(err)
	}
}

func publishCommands(ctx context.Context, sessionID string, commandBus *cqrs.CommandBus) {

	// Initialize the mod
	cmd := &event.Queue{
		Event:     event.NewExecutionEvent(ctx),
		Workspace: "e-gineer/scratch",
	}

	if err := commandBus.Send(ctx, cmd); err != nil {
		panic(err)
	}

	// Manually trigger some pipelines for testing
	// TODO - these should be triggered instead (e.g. cron, webhook, etc)
	for _, s := range []string{"call_pipelines_in_for_loop"} {
		time.Sleep(0 * time.Second)
		fmt.Println()
		pipelineCmd := &event.PipelineQueue{
			Event: event.NewChildEvent(cmd.Event),
			Name:  s,
			//Input:        e.Input,
		}
		if err := commandBus.Send(ctx, pipelineCmd); err != nil {
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

			var pe event.PayloadWithEvent
			err := json.Unmarshal(msg.Payload, &pe)
			if err != nil {
				panic("TODO - invalid log payload, log me?")
			}

			executionID := pe.Event.ExecutionID
			if executionID == "" {
				m := fmt.Sprintf("SHOULD NOT HAPPEN - No execution_id found in payload: %s", msg.Payload)
				return nil, errors.New(m)
			}

			var payload map[string]interface{}
			err = json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				panic("TODO - invalid log payload, log me?")
			}

			payloadWithoutEvent := make(map[string]interface{})
			for key, value := range payload {
				if key == "event" {
					continue
				}
				payloadWithoutEvent[key] = value
			}
			fmt.Printf("%s %-30s %s\n", pe.Event.CreatedAt.Format("15:04:05.000"), msg.Metadata["name"], payloadWithoutEvent)

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

/*
func DumpState(ctx context.Context) message.HandlerMiddleware {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			execution.Dump(ctx)
			return h(msg)
		}
	}
}
*/
