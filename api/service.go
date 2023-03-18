package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/gin-gonic/gin"
	"github.com/turbot/steampipe-pipelines/es/event"
	"github.com/turbot/steampipe-pipelines/es/execution"
)

type RunPipeline struct {
	Name string `json:"name"`
}

func StartService(ctx context.Context, runID string, commandBus *cqrs.CommandBus) {
	r := gin.Default()

	// curl -X POST http://localhost:8080/pipeline_execution -H 'Content-Type: application/json' -d '{"name": "simple_parallel"}'
	r.POST("/pipeline_execution", func(c *gin.Context) {

		var input RunPipeline
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		pipelineCmd := &event.PipelineQueue{
			Event: event.NewExecutionEvent(ctx),
			Name:  input.Name,
		}
		if err := commandBus.Send(ctx, pipelineCmd); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"cmd": pipelineCmd})

	})

	// curl http://localhost:8080/pipeline_execution/exec_cg9oujlnsevmt3umtas0
	r.GET("/pipeline_execution/:id", func(c *gin.Context) {
		id := c.Param("id")
		e := event.NewExecutionEvent(ctx)
		e.ExecutionID = id
		ex, err := execution.NewExecution(ctx, execution.WithEvent(e))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": ex,
		})
	})

	// List all exec_* files from the directory and get the execution details
	// for each ID (format exec_).
	// Example request - curl http://localhost:8080/pipeline_execution/exec_cg9oujlnsevmt3umtas0
	r.GET("/pipeline_execution", func(c *gin.Context) {

		executions := []*execution.Execution{}

		err := filepath.Walk("logs", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			logName := info.Name()
			if len(logName) <= 5 || logName[0:5] != "exec_" {
				return nil
			}
			id := strings.Split(logName, ".")[0]
			e := event.NewExecutionEvent(ctx)
			e.ExecutionID = id
			ex, err := execution.NewExecution(ctx, execution.WithEvent(e))
			if err != nil {
				return err
			}
			executions = append(executions, ex)
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"items": executions})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
