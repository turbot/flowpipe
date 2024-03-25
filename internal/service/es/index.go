package es

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	slogwatermill "github.com/denisss025/slog-watermill"
	_ "github.com/garsue/watermillzap"
	"github.com/turbot/flowpipe/internal/es/command"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/es/handler"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/flowpipe/internal/service/es/middleware"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

type ESService struct {
	ctx        context.Context
	runID      string
	CommandBus handler.FpCommandBus
	EventBus   command.FpEventBus
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
	err := es.CommandBus.Send(es.ctx, cmd)
	return err
}

func (es *ESService) Raise(evt interface{}) error {
	err := es.EventBus.Publish(es.ctx, evt)
	return err
}

func (es *ESService) IsRunning() bool {
	if es.router == nil {
		return false
	}

	return es.router.IsRunning()
}

func (es *ESService) Start() error {
	slog.Debug("ES starting")
	defer slog.Debug("ES started")

	execution.InitGlobalStepSemaphores()

	cqrsMarshaler := cqrs.JSONMarshaler{}

	goChannelConfig := gochannel.Config{
		//TODO - I really don't understand this and I'm not sure it's necessary.
		// OutputChannelBuffer: 10000,
		// Persistent:          true,
	}
	wLogger := slogwatermill.New(log.FlowpipeLogger())
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
	// 2. anything else
	//
	// Recoverer handles panics from handlers.
	router.AddMiddleware(middleware.PanicRecovererMiddleware(es.ctx))

	// Do not remove this middleware. This stops Watermill from doing infinite loop if we return an error on the handler (which may be a valid case)
	retryer := middleware.Retry{
		MaxRetries: 0,
	}

	router.AddMiddleware(retryer.Middleware)

	// router.AddMiddleware(middleware.EventMiddleware(es.ctx))

	// cqrs.Facade is facade for Command and Event buses and processors.
	// You can use facade, or create buses and processors manually (you can inspire with cqrs.NewFacade)
	//nolint:staticcheck // TODO victor look at this
	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			// we are using queue RabbitMQ config, so we need to have topic per command type
			return commandName
		},

		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				command.PipelineCancelHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelineFailHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelineFinishHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelineLoadHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelinePauseHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelinePlanHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelineQueueHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelineResumeHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.PipelineStartHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.StepPipelineFinishHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.StepQueueHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.StepStartHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
				command.StepForEachPlanHandler{EventBus: &command.FpEventBusImpl{Eb: eb}},
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
				handler.PipelineCanceled{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelineFailed{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelineFinished{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelineLoaded{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelinePaused{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelinePlanned{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelineQueued{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelineResumed{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.PipelineStarted{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.StepFinished{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.StepQueued{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.StepPipelineStarted{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
				handler.StepForEachPlanned{CommandBus: &handler.FpCommandBusImpl{Cb: cb}},
			}
		},
		EventsPublisher: eventsPubSub,
		EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			// we can reuse subscriber, because all commands have separated topics
			return eventsPubSub, nil
		},
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
	es.CommandBus = &handler.FpCommandBusImpl{Cb: cqrsFacade.CommandBus()}
	es.EventBus = &command.FpEventBusImpl{Eb: cqrsFacade.EventBus()}

	es.router = router

	// processors are based on router, so they will work when router will start
	go func() {
		err := router.Run(es.ctx)
		if err != nil {
			slog.Error("Error running event sourcing enging", "error", err)
			os.Exit(1)
		}
	}()

	return nil
}

func (es *ESService) Stop() error {

	slog.Debug("ES stopping")
	defer slog.Debug("ES stopped")

	return es.router.Close()
}
