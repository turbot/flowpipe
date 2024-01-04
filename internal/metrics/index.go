package metrics

import (
	"sync"
	"time"
)

type ExecutionRun struct {
	ExecutionID    string    `json:"execution_id"`
	Pipeline       string    `json:"pipeline"`
	StartTimestamp time.Time `json:"start_timestamp"`
	EndTimestamp   time.Time `json:"end_timestamp"`
}

type RunMetric struct {
	executionRuns sync.Map
}

var RunMetricInstance = &RunMetric{
	executionRuns: sync.Map{},
}

func (m *RunMetric) StartExecution(executionID, pipeline string) {
	pipelineRun := &ExecutionRun{
		ExecutionID:    executionID,
		Pipeline:       pipeline,
		StartTimestamp: time.Now(),
	}

	m.executionRuns.Store(executionID, pipelineRun)
}

func (m *RunMetric) RunningExecutions() []ExecutionRun {
	executions := []ExecutionRun{}
	m.executionRuns.Range(func(key, value interface{}) bool {
		executions = append(executions, *value.(*ExecutionRun))
		return true
	})
	return executions
}

func (m *RunMetric) EndExecution(executionID string) {
	pipelineRun, ok := m.executionRuns.Load(executionID)
	if !ok {
		return
	}

	pipelineRun.(*ExecutionRun).EndTimestamp = time.Now()

	m.executionRuns.Delete(executionID)
}
