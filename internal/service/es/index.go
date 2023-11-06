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
	"github.com/turbot/flowpipe/internal/es/command"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"

	"github.com/turbot/flowpipe/internal/service/es/middleware"
	"github.com/turbot/flowpipe/internal/util"
)

type ESService struct {
	ctx        context.Context
	runID      string
	commandBus *handler.FpCommandBus
	eventBus   *command.FpEventBus
	router     *message.Router

	RootMod   *modconfig.Mod
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

func (es *ESService) Raise(evt interface{}) error {
	err := es.eventBus.Publish(es.ctx, evt)
	return err
}

func (es *ESService) Start() error {
	// Convenience
	logger := fplog.Logger(es.ctx)

	logger.Debug("ES starting")
	defer logger.Debug("ES started")

	outputDir := viper.GetString(constants.ArgOutputDir)
	logDir := viper.GetString(constants.ArgLogDir)

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

	//
	// IMPORTANT: middleware order:
	// 1. panic recoverer
	// 2. retry (to ack message after panic recoverer)
	// 3. log event (?)
	//
	// Recoverer handles panics from handlers.
	router.AddMiddleware(middleware.PanicRecovererMiddleware(es.ctx))

	retryer := middleware.Retry{
		MaxRetries: 0,
	}

	router.AddMiddleware(retryer.Middleware)

	// Log to file for creation of state
	// ! Ensure that the log event middleware is the first middleware to be added in the router
	// ! so the log entry is written ASAP
	// router.AddMiddleware(middleware.LogEventMiddlewareWithContext(es.ctx))

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
				command.PipelineCancelHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineFailHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineFinishHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineLoadHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelinePauseHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelinePlanHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineQueueHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineResumeHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineStartHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineStepFinishHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineStepQueueHandler{EventBus: &command.FpEventBus{Eb: eb}},
				command.PipelineStepStartHandler{EventBus: &command.FpEventBus{Eb: eb}},
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
				handler.PipelineCanceled{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineFailed{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineFinished{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineLoaded{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelinePaused{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelinePlanned{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineQueued{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineResumed{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineStarted{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineStepFinished{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineStepQueued{CommandBus: &handler.FpCommandBus{Cb: cb}},
				handler.PipelineStepStarted{CommandBus: &handler.FpCommandBus{Cb: cb}},
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
		return perr.InternalWithMessage("cqrsFacade is nil")
	}

	runID := util.NewProcessID()

	es.runID = runID
	es.commandBus = &handler.FpCommandBus{Cb: cqrsFacade.CommandBus()}
	es.eventBus = &command.FpEventBus{Eb: cqrsFacade.EventBus()}

	es.router = router

	// processors are based on router, so they will work when router will start
	go func() {
		err := router.Run(es.ctx)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (es *ESService) Stop() error {
	logger := fplog.Logger(es.ctx)

	logger.Debug("ES stopping")
	defer logger.Debug("ES stopped")

	es.router.Close()
	return nil
}
