package es

import (
	"context"
	"time"

	"github.com/turbot/flowpipe/config"
	"github.com/turbot/flowpipe/fplog"
	"github.com/turbot/flowpipe/pipeline"
)

type ESService struct {
	ctx context.Context

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

func (es *ESService) Start() error {
	// Convenience
	logger := fplog.Logger(es.ctx)

	logger.Debug("ES starting")
	defer logger.Debug("ES started")

	c := config.GetConfigFromContext(es.ctx)

	pipelineDir := c.Viper.GetString("pipeline.dir")

	logger.Debug("Pipeline dir", "dir", pipelineDir)

	_, err := pipeline.LoadPipelines(es.ctx, pipelineDir)
	if err != nil {
		return err
	}

	return nil
}
