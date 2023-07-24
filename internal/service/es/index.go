package es

import (
	"context"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/garsue/watermillzap"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/command"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/pipeline"

	"github.com/turbot/flowpipe/internal/service/es/middleware"
	"github.com/turbot/flowpipe/internal/util"
)

type ESService struct {
	ctx        context.Context
	runID      string
	commandBus *cqrs.CommandBus

	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	StoppedAt *time.Time `json:"stopped_at,omitempty"`
}

func NewESService(ctx context.Context) (*ESService, error) {
	// Defaults
	es := &ESService{
		ctx:    ctx,
		Status: "initialized",
	}
	return es, nil
}

func (es *ESService) Send(cmd interface{}) error {
	err := es.commandBus.Send(es.ctx, cmd)
	return err
}

func (es *ESService) Start() error {
	// Convenience
	logger := fplog.Logger(es.ctx)

	logger.Debug("ES starting")
	defer logger.Debug("ES started")

	pipelineDir := viper.GetString("pipeline.dir")

	logger.Debug("Pipeline dir", "dir", pipelineDir)

	// _, err := pipeline.LoadPipelines(es.ctx, pipelineDir)
	// if err != nil {
	// 	return err
	// }

	outputDir := viper.GetString("output.dir")
	logger.Debug("Output dir", "dir", pipelineDir)

	logDir := viper.GetString("log.dir")
	logger.Debug("Log dir", "dir", pipelineDir)

	// Check if the provided output dir exists, if not create it
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			return err
		}
	}

	// Check if the provided execution log dir exists, if not create it
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.Mkdir(logDir, 0755)
		if err != nil {
			return err
		}
	}

	pipelines, err := pipeline.LoadPipelines(es.ctx, pipelineDir)
	if err != nil {
		return err
	}

	inMemoryCache := cache.GetCache()
	var pipelineNames []string

	for pipelineName := range pipelines {
		pipelineNames = append(pipelineNames, pipelineName)

		// TODO: how do we want to do this?
		inMemoryCache.SetWithTTL(pipelineName, pipelines[pipelineName], 24*7*52*99*time.Hour)
	}

	inMemoryCache.SetWithTTL("#pipeline.names", pipelineNames, 24*7*52*99*time.Hour)

	cqrsMarshaler := cqrs.JSONMarshaler{}

	goChannelConfig := gochannel.Config{
		//TODO - I really don't understand this and I'm not sure it's necessary.
		// OutputChannelBuffer: 10000,
		// Persistent:          true,
	}
	wLogger := watermillzap.NewLogger(logger.Zap)

	commandsPubSub := gochannel.NewGoChannel(goChannelConfig, wLogger)
	eventsPubSub := gochannel.NewGoChannel(goChannelConfig, wLogger)

	// CQRS is built on messages router. Detailed documentation: https://watermill.io/docs/messages-router/
	router, err := message.NewRouter(message.RouterConfig{}, wLogger)
	if err != nil {
		return err
	}

	// Simple middleware which will recover panics from event or command handlers.
	// More about router middlewares you can find in the documentation:
	// https://watermill.io/docs/messages-router/#middleware
	//
	// List of available middlewares you can find in message/router/middleware.

	// Recoverer handles panics from handlers.
	router.AddMiddleware(middleware.PanicRecovererMiddleware(es.ctx))

	// Log to file for creation of state
	// ! Ensure that the log event middleware is the first middleware to be added in the router
	// ! so the log entry is written ASAP
	router.AddMiddleware(middleware.LogEventMiddlewareWithContext(es.ctx))

	retryer := middleware.Retry{
		MaxRetries: 0,
	}

	router.AddMiddleware(retryer.Middleware)

	plannerControl := middleware.NewPlannerControl(es.ctx)
	router.AddMiddleware(plannerControl.Middleware)

	// Delay PipelineStepStart command (if required)
	// router.AddMiddleware(middleware.PipelineStepStartCommandDelayMiddlewareWithContext(es.ctx))

	// cqrs.Facade is facade for Command and Event buses and processors.
	// You can use facade, or create buses and processors manually (you can inspire with cqrs.NewFacade)
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			// we are using queue RabbitMQ config, so we need to have topic per command type
			return commandName
		},

		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.PipelineCancelHandler{EventBus: eb},
				command.PipelineFailHandler{EventBus: eb},
				command.PipelineFinishHandler{EventBus: eb},
				command.PipelineLoadHandler{EventBus: eb},
				command.PipelinePauseHandler{EventBus: eb},
				command.PipelinePlanHandler{EventBus: eb},
				command.PipelineQueueHandler{EventBus: eb},
				command.PipelineResumeHandler{EventBus: eb},
				command.PipelineStartHandler{EventBus: eb},
				command.PipelineStepFinishHandler{EventBus: eb},
				command.PipelineStepQueueHandler{EventBus: eb},
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
				handler.PipelineCanceled{CommandBus: cb},
				handler.PipelineFailed{CommandBus: cb},
				handler.PipelineFinished{CommandBus: cb},
				handler.PipelineLoaded{CommandBus: cb},
				handler.PipelinePaused{CommandBus: cb},
				handler.PipelinePlanned{CommandBus: cb},
				handler.PipelineQueued{CommandBus: cb},
				handler.PipelineResumed{CommandBus: cb},
				handler.PipelineStarted{CommandBus: cb},
				handler.PipelineStepFinished{CommandBus: cb},
				handler.PipelineStepQueued{CommandBus: cb},
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
		Logger:                wLogger,
	})
	if err != nil {
		return err
	}

	if cqrsFacade == nil {
		return fperr.InternalWithMessage("cqrsFacade is nil")
	}

	runID := util.NewProcessID()

	es.runID = runID
	es.commandBus = cqrsFacade.CommandBus()

	// processors are based on router, so they will work when router will start
	go func() {
		err := router.Run(es.ctx)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}
