package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/primitive"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/templates"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) InputRegisterAPI(router *gin.RouterGroup) {
	// router.POST("/input/:input/:hash", api.runInputPost)
	router.POST("/input/slack/:input/:hash", api.runSlackInputPost)
	router.GET("/input/email/:input/:hash", api.runInputEmailGet)
}

type JSONPayload struct {
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	ExecutionID         string `json:"execution_id"`
}

func (api *APIService) runPipeline(c *gin.Context, inputType primitive.InputType, executionID, pipelineExecutionID, stepExecutionID string) {
	logger := fplog.Logger(api.ctx)

	ex, err := execution.NewExecution(api.ctx)
	if err != nil {
		logger.Error("error creating execution", "error", err)
		common.AbortWithError(c, err)
		return
	}

	var stepOutput *modconfig.Output

	evt := &event.Event{
		ExecutionID: executionID,
	}

	err = ex.LoadProcess(evt)
	if err != nil {
		logger.Error("error loading process", "error", err)
		common.AbortWithError(c, err)
		return
	}

	// Find the step start for the step execution id
	pipelineExecution := ex.PipelineExecutions[pipelineExecutionID]
	if pipelineExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage("pipeline execution "+pipelineExecutionID+" not found"))
		return
	}

	stepExecution := pipelineExecution.StepExecutions[stepExecutionID]
	if stepExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage("step execution "+stepExecutionID+" not found"))
		return
	}

	if c.Request.Body != nil && inputType == primitive.InputTypeSlack {
		var prompt, userName string
		var err error
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}

		decodedValue, err := url.QueryUnescape(string(bodyBytes))
		if err != nil {
			log.Fatal(err)
			return
		}

		decodedValue = decodedValue[8:]

		// TODO: Remove this log
		// logger.Info("decodedValue", "decodedValue>>>>>", decodedValue)

		var bodyJSON map[string]interface{}
		err = json.Unmarshal([]byte(decodedValue), &bodyJSON)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}

		// Decode the callback_id to extract the execution_id, pipeline_execution_id and step_execution_id
		rawDecodedText, err := base64.StdEncoding.DecodeString(bodyJSON["callback_id"].(string))
		if err != nil {
			common.AbortWithError(c, err)
			return
		}

		var decodedText JSONPayload
		err = json.Unmarshal(rawDecodedText, &decodedText)
		if err != nil {
			common.AbortWithError(c, err)
			return
		}
		// stepOutput, err = input.ProcessOutput(c, inputType, bodyBytes)
		// if err != nil {
		// 	common.AbortWithError(c, err)
		// 	return
		// }

		pipelineExecutionID = decodedText.PipelineExecutionID
		stepExecutionID = decodedText.StepExecutionID

		// TODO: Refactor to extract the prompt in a better way
		if bodyJSON["original_message"] != nil {
			originMessage := bodyJSON["original_message"].(map[string]interface{})
			if originMessage != nil {
				attachments := originMessage["attachments"].([]interface{})
				for _, attachment := range attachments {
					prompt = attachment.(map[string]interface{})["text"].(string)
					break
				}
			}
		}

		if bodyJSON["user"] != nil {
			userData := bodyJSON["user"].(map[string]interface{})
			userName = userData["id"].(string)
		}

		var value interface{}
		if bodyJSON["actions"] != nil {
			for _, action := range bodyJSON["actions"].([]interface{}) {
				if action.(map[string]interface{})["type"] == "button" {
					value = action.(map[string]interface{})["value"]
				}
				if action.(map[string]interface{})["type"] == "select" {
					/*
							"actions": [
						    {
						      "name": "Choose an option",
						      "type": "select",
						      "selected_options": [
						        {
						          "value": "Admin"
						        }
						      ]
						    }
						  ],
					*/
					selectedOptions := action.(map[string]interface{})["selected_options"]
					if selectedOptions != nil {
						for _, selectedOption := range selectedOptions.([]interface{}) {
							value = selectedOption.(map[string]interface{})["value"]
						}
					}
				}
			}
		}

		output := modconfig.Output{
			Data: map[string]interface{}{
				"value": value,
			},
		}

		stepOutput = &output

		// stepOutput, err = input.ProcessOutput(api.ctx, inputType, bodyBytes)
		// if err != nil {
		// 	logger.Error("error processing output", "error", err)
		// 	common.AbortWithError(c, err)
		// 	return
		// }

		logger.Debug("stepOutput", "stepOutput", &output)

		c.String(http.StatusOK, fmt.Sprintf("%s <@%s> has selected `%v`", prompt, userName, value))
	} else {
		if pipelineExecution.Status == "finished" {
			alreadyAcknowledgedInputTemplate, err := templates.HTMLTemplate("already-acknowledged-input.html")
			if err != nil {
				logger.Error("error reading the template file", "error", err)
				common.AbortWithError(c, err)
				return
			}
			renderHTMLWithValues(c, string(alreadyAcknowledgedInputTemplate), gin.H{})
		} else {
			input := primitive.Input{}
			stepOutput, err = input.ProcessOutput(c, inputType, nil)
			if err != nil {
				logger.Error("error processing output", "error", err)
				common.AbortWithError(c, err)
				return
			}

			acknowledgeInputTemplate, err := templates.HTMLTemplate("acknowledge-input.html")
			if err != nil {
				logger.Error("error reading the template file", "error", err)
				common.AbortWithError(c, err)
				return
			}
			renderHTMLWithValues(c, string(acknowledgeInputTemplate), gin.H{"response": stepOutput.Data["value"]})
		}
	}

	pipelineStepFinishedEvent, err := event.NewStepFinished()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	pipelineStepFinishedEvent.Event = evt
	pipelineStepFinishedEvent.PipelineExecutionID = pipelineExecutionID
	pipelineStepFinishedEvent.StepExecutionID = stepExecutionID
	pipelineStepFinishedEvent.StepForEach = stepExecution.StepForEach
	pipelineStepFinishedEvent.StepOutput = map[string]interface{}{}
	pipelineStepFinishedEvent.Output = stepOutput

	err = api.EsService.Raise(pipelineStepFinishedEvent)
	if err != nil {
		common.AbortWithError(c, err)
	}
}

func (api *APIService) runInputEmailGet(c *gin.Context) {
	logger := fplog.Logger(api.ctx)

	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}

	logger.Info("executionMode", "executionMode", executionMode)

	inputName := inputUri.Input
	inputHash := inputUri.Hash

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		logger.Error("salt not found")
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	hashString := util.CalculateHash(inputName, salt.(string))

	if hashString != inputHash {
		logger.Warn("invalid hash, but we're ignoring it for now ... ", "hash", inputHash, "expected", hashString)
		// common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash for "+inputName))
		// return
	}

	api.runPipeline(c, primitive.InputTypeEmail, inputQuery.ExecutionID, inputQuery.PipelineExecutionID, inputQuery.StepExecutionID)
}

func (api *APIService) runSlackInputPost(c *gin.Context) {
	logger := fplog.Logger(api.ctx)

	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}

	logger.Info("executionMode", "executionMode", executionMode)

	inputName := inputUri.Input
	inputHash := inputUri.Hash

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		logger.Error("salt not found")
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}

	hashString := util.CalculateHash(inputName, salt.(string))

	if hashString != inputHash {
		logger.Warn("invalid hash, but we're ignoring it for now ... ", "hash", inputHash, "expected", hashString)
		// common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash for "+inputName))
		// return
	}

	api.runPipeline(c, primitive.InputTypeSlack, inputQuery.ExecutionID, inputQuery.PipelineExecutionID, inputQuery.StepExecutionID)
}

// Custom function to render HTML with values
func renderHTMLWithValues(c *gin.Context, templateContent string, data interface{}) {
	tmpl, err := template.New("html").Parse(templateContent)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse template")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Failed to execute template")
		return
	}
}
