package api

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/es/command"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

func (api *APIService) FormRegisterAPI(router *gin.RouterGroup) {
	router.GET("/form/:id/:hash", api.getFormData)          // used by UI to get data to populate form
	router.POST("/form/:id/:hash/submit", api.postFormData) // used by UI, cURL, etc for form response
}

func (api *APIService) getFormData(c *gin.Context) {
	var uri types.InputIDHash
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// verify hash
	salt, err := util.GetGlobalSalt()
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}
	hashString, err := util.CalculateHash(uri.ID, salt)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("error calculating hash"))
		return
	}
	if hashString != uri.Hash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash"))
		return
	}

	output, err := httpFormDataFromId(uri.ID)
	if err != nil {
		common.AbortWithError(c, err) // will be perr type
		return
	}

	// get step exec
	exec, err := execution.GetExecution(output.ExecutionID)
	if err != nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("execution %s not found", output.ExecutionID)))
		return
	}
	pExec := exec.PipelineExecutions[output.PipelineExecutionID]
	if helpers.IsNil(pExec) {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution %s not found", output.PipelineExecutionID)))
		return
	}
	sExec := pExec.StepExecutions[output.StepExecutionID]
	if helpers.IsNil(sExec) {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("pipeline step execution %s not found", output.StepExecutionID)))
		return
	}

	stepFullName := sExec.Name
	stepName := strings.Split(stepFullName, ".")[len(strings.Split(stepFullName, "."))-1]
	stepType := strings.Split(stepFullName, ".")[len(strings.Split(stepFullName, "."))-2]

	switch stepType {
	case "input":
		output.Inputs[stepName] = httpFormDataInputFromInputStep(sExec.Input)
	case "form":
		// TODO: implement
		common.AbortWithError(c, perr.InternalWithMessage("form is not yet implemented"))
		return
	default:
		common.AbortWithError(c, perr.InternalWithMessage(fmt.Sprintf("step type %s is not supported", stepType)))
		return
	}

	output.Status = sExec.Status

	c.JSON(200, output)
}

func (api *APIService) postFormData(c *gin.Context) {
	var uri types.InputIDHash
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// verify hash
	salt, err := util.GetGlobalSalt()
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("salt not found"))
		return
	}
	hashString, err := util.CalculateHash(uri.ID, salt)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage("error calculating hash"))
		return
	}
	if hashString != uri.Hash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash"))
		return
	}

	output, err := httpFormDataFromId(uri.ID)
	if err != nil {
		common.AbortWithError(c, err) // will be perr type
		return
	}

	plannerMutex := event.GetEventStoreMutex(output.ExecutionID)
	plannerMutex.Lock()
	defer func() {
		if plannerMutex != nil {
			plannerMutex.Unlock()
		}
	}()

	ex, err := execution.GetExecution(output.ExecutionID)
	if err != nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("execution %s not found", output.ExecutionID)))
		return
	}

	pipelineExecution := ex.PipelineExecutions[output.PipelineExecutionID]
	if pipelineExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution %s not found", output.PipelineExecutionID)))
		return
	}

	stepExecution := pipelineExecution.StepExecutions[output.StepExecutionID]
	if stepExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("step execution %s not found", output.StepExecutionID)))
		return
	}

	pipelineDefn, err := ex.PipelineDefinition(output.PipelineExecutionID)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage(fmt.Sprintf("error getting pipeline definition: %s", err.Error())))
		return
	}

	stepDefn := pipelineDefn.GetStep(stepExecution.Name)

	if !httpFormValidateNotifiers(stepExecution) {
		// if not a valid notifier for this endpoint return NotFound
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("step execution %s not found", output.StepExecutionID)))
		return
	}

	if pipelineExecution.IsFinished() || pipelineExecution.IsFinishing() || stepExecution.Status == "finished" {
		common.AbortWithError(c, perr.ConflictWithMessage(fmt.Sprintf("step %s has already been processed or is no longer required due to pipeline completion", output.StepExecutionID)))
		return
	}

	stepFullName := stepExecution.Name
	stepName := strings.Split(stepFullName, ".")[len(strings.Split(stepFullName, "."))-1]
	stepType := strings.Split(stepFullName, ".")[len(strings.Split(stepFullName, "."))-2]

	var parsedBody map[string]any
	switch c.ContentType() {
	case "application/x-www-form-urlencoded":
		err = c.Bind(&parsedBody)
		if err != nil {
			common.AbortWithError(c, perr.InternalWithMessage("error parsing body content"))
			return
		}
	default:
		err = c.BindJSON(&parsedBody)
		if err != nil {
			common.AbortWithError(c, perr.InternalWithMessage("error parsing body content"))
			return
		}
	}

	switch stepType {
	case "input":
		output.Inputs[stepName] = httpFormDataInputFromInputStep(stepExecution.Input)

		if parsedBody[stepName] != nil {
			val := parsedBody[stepName]
			inputType := *output.Inputs[stepName].InputType
			if inputType != constants.InputTypeText {
				var allowedValues []string
				for _, o := range output.Inputs[stepName].Options {
					allowedValues = append(allowedValues, *o.Value)
				}
				if !httpFormDataValidateResponse(val, allowedValues) {
					common.AbortWithError(c, perr.BadRequestWithMessage(fmt.Sprintf("submitted value %v contains invalid option value(s).", val)))
					return
				}
			}
			if i, ok := output.Inputs[stepName]; ok {
				switch *i.InputType {
				case constants.InputTypeMultiSelect:
					if v, ok := val.(string); ok {
						val = []string{v}
					}
				}
			}
			err := api.finishInputStepFromForm(ex, stepExecution, pipelineDefn, stepDefn, val)
			if err != nil {
				common.AbortWithError(c, err)
				return
			}
			output.Status = "finished"
			c.JSON(200, output)
		} else {
			common.AbortWithError(c, perr.BadRequestWithMessage(fmt.Sprintf("missing expected key %s", stepName)))
			return
		}
	case "form":
		// TODO: implement
		common.AbortWithError(c, perr.InternalWithMessage("form is not yet implemented"))
		return
	default:
		common.AbortWithError(c, perr.InternalWithMessage(fmt.Sprintf("step type %s is not supported", stepType)))
		return
	}
}

// TODO: consider struct naming / relocation to types?
type httpFormData struct {
	ExecutionID         string                       `json:"execution_id"`
	PipelineExecutionID string                       `json:"pipeline_execution_id"`
	StepExecutionID     string                       `json:"step_execution_id"`
	Status              string                       `json:"status"`
	Inputs              map[string]httpFormDataInput `json:"inputs"`
}

type httpFormDataInput struct {
	Prompt    *string                    `json:"prompt,omitempty"`
	InputType *string                    `json:"input_type,omitempty"`
	Options   []httpFormDataInputOptions `json:"options,omitempty"`
}

type httpFormDataInputOptions struct {
	Label    *string `json:"label,omitempty"`
	Value    *string `json:"value,omitempty"`
	Selected *bool   `json:"selected,omitempty"`
	Style    *string `json:"style,omitempty"`
}

func httpFormDataFromId(id string) (httpFormData, error) {
	var output httpFormData
	output.Inputs = make(map[string]httpFormDataInput)

	executionID, pipelineExecutionID, stepExecutionID, ok := db.ResolveShortStepExecutionID(id)
	if !ok {
		return output, perr.NotFoundWithMessage(fmt.Sprintf("unable to find step for id %s - id may be incorrect or step may already be completed.", id))
	}
	output.ExecutionID = executionID
	output.PipelineExecutionID = pipelineExecutionID
	output.StepExecutionID = stepExecutionID

	return output, nil
}

func httpFormDataInputFromInputStep(input modconfig.Input) httpFormDataInput {
	var output httpFormDataInput

	if p, ok := input[schema.AttributeTypePrompt].(string); ok {
		output.Prompt = &p
	}
	if t, ok := input[schema.AttributeTypeType].(string); ok {
		output.InputType = &t
	}
	if !helpers.IsNil(input[schema.AttributeTypeOptions]) {
		for _, o := range input[schema.AttributeTypeOptions].([]any) {
			opt := o.(map[string]any)
			option := httpFormDataInputOptions{}
			if l, ok := opt[schema.AttributeTypeLabel].(string); ok {
				option.Label = &l
			}
			if v, ok := opt[schema.AttributeTypeValue].(string); ok {
				option.Value = &v
				if helpers.IsNil(option.Label) {
					option.Label = &v
				}
			}
			if s, ok := opt[schema.AttributeTypeSelected].(bool); ok {
				option.Selected = &s
			}
			if s, ok := opt[schema.AttributeTypeStyle].(string); ok {
				option.Style = &s
			}
			output.Options = append(output.Options, option)
		}
	}

	return output
}

func httpFormDataValidateResponse(val any, allowedOptions []string) bool {
	switch v := val.(type) {
	case string:
		return slices.Contains(allowedOptions, v)
	case []any:
		for _, x := range v {
			if !slices.Contains(allowedOptions, x.(string)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func httpFormValidateNotifiers(sexec *execution.StepExecution) bool {
	validNotifiers := []string{"http", "email"}
	input := sexec.Input
	if notifier, ok := input[schema.AttributeTypeNotifier].(map[string]any); ok {
		if notifies, ok := notifier[schema.AttributeTypeNotifies].([]any); ok {
			for _, n := range notifies {
				if notify, ok := n.(map[string]any); ok {
					if integration, ok := notify["integration"].(map[string]any); ok {
						integrationType := integration["type"].(string)
						if slices.Contains(validNotifiers, integrationType) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (api *APIService) finishInputStepFromForm(ex *execution.ExecutionInMemory, stepExecution *execution.StepExecution, pipelineDefn *modconfig.Pipeline, stepDefn modconfig.PipelineStep, value any) error {
	out := modconfig.Output{
		Data: map[string]any{
			"value": value,
		},
		Status: "finished",
	}

	err := command.EndStepFromApi(ex, stepExecution, pipelineDefn, stepDefn, &out, api.EsService.EventBus)
	if err != nil {
		return err
	}

	err = execution.ReleasePipelineExecutionStepSemaphore(stepExecution.PipelineExecutionID, stepDefn)
	return err
}
