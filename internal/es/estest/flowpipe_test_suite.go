package estest

import (
	"context"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/manager"
)

type FlowpipeTestSuite struct {
	esService *es.ESService
	manager   *manager.Manager
	ctx       context.Context
}
