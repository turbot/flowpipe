package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/xid"
	"github.com/turbot/steampipe-pipelines/es/command"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/handler"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

func main() {
	logger := watermill.NewStdLogger(false, false)
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
	router.AddMiddleware(LogEventMiddleware)

	// cqrs.Facade is facade for Command and Event buses and processors.
	// You can use facade, or create buses and processors manually (you can inspire with cqrs.NewFacade)
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			// we are using queue RabbitMQ config, so we need to have topic per command type
			return commandName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.PipelineRunQueueHandler{EventBus: eb},
				command.PipelineRunLoadHandler{EventBus: eb},
				command.PipelineRunStartHandler{EventBus: eb},
				command.PipelineRunStepExecuteHandler{EventBus: eb},
				command.PipelineRunStepPrimitiveExecuteHandler{EventBus: eb},
				command.PipelineRunStepHTTPRequestExecuteHandler{EventBus: eb},
				command.PipelineRunFinishHandler{EventBus: eb},
				command.PipelineRunFailHandler{EventBus: eb},

				command.QueueHandler{EventBus: eb},
				command.LoadHandler{EventBus: eb},
				command.StartHandler{EventBus: eb},
				command.PlanHandler{EventBus: eb},
				command.PipelineStartHandler{EventBus: eb},
				command.PipelinePlanHandler{EventBus: eb},
				command.PipelineFinishHandler{EventBus: eb},
				command.ExecuteHandler{EventBus: eb},
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
				handler.PipelineRunQueued{CommandBus: cb},
				handler.PipelineRunLoaded{CommandBus: cb},
				handler.PipelineRunStarted{CommandBus: cb},
				handler.PipelineRunStepExecuted{CommandBus: cb},
				handler.PipelineRunStepPrimitivePlanned{CommandBus: cb},
				handler.PipelineRunStepHTTPRequestPlanned{CommandBus: cb},
				handler.PipelineRunStepFailed{CommandBus: cb},
				handler.PipelineRunFinished{CommandBus: cb},
				handler.PipelineRunFailed{CommandBus: cb},

				handler.Queued{CommandBus: cb},
				handler.Loaded{CommandBus: cb},
				handler.Started{CommandBus: cb},
				handler.Planned{CommandBus: cb},
				handler.PipelineStarted{CommandBus: cb},
				handler.PipelinePlanned{CommandBus: cb},
				handler.PipelineFinished{CommandBus: cb},
				handler.Executed{CommandBus: cb},
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
		cmd := &command.Stop{
			RunID: runID,
		}
		if err := cqrsFacade.CommandBus().Send(context.Background(), cmd); err != nil {
			panic(err)
		}
	}()

	// publish commands every second to simulate incoming traffic
	go publishCommands(runID, cqrsFacade.CommandBus())

	// processors are based on router, so they will work when router will start
	if err := router.Run(context.Background()); err != nil {
		panic(err)
	}
}

func publishCommands(runID string, commandBus *cqrs.CommandBus) {
	cmd := &event.Queue{
		IdentityID:   "e-gineer",
		WorkspaceID:  "scratch",
		PipelineName: fmt.Sprintf("my_pipeline_%d", 0),
		RunID:        runID,
	}
	if err := commandBus.Send(context.Background(), cmd); err != nil {
		panic(err)
	}
}

type PipelinePayload struct {
	RunID string `json:"run_id"`
}

func LogEventMiddleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {

		// Get the run ID from the payload
		var pp PipelinePayload
		err := json.Unmarshal(msg.Payload, &pp)
		if err != nil {
			log.Println(err)
		}

		// event.log
		f, err := os.OpenFile(fmt.Sprintf("logs/%s.jsonl", pp.RunID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		startOfLine := []byte(fmt.Sprintf(`{"event_type":"%s","timestamp":"%s","payload":`, msg.Metadata["name"], time.Now().Format(time.RFC3339)))
		endOfLine := []byte("}\n")
		logJson := append(startOfLine, msg.Payload...)
		logJson = append(logJson, endOfLine...)
		if _, err := f.Write(logJson); err != nil {
			fmt.Println("error", err)
		}

		// stdout
		//fmt.Printf("[event  ] %s: %s\n", msg.Metadata["name"], string(msg.Payload))

		return h(msg)
	}
}
