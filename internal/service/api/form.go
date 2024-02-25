package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/es/event"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"net/url"
	"strings"
	"time"
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

	output, err := webFormDataFromId(uri.ID)
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

	sDef, err := exec.StepDefinition(pExec.ID, sExec.ID)
	if err != nil {
		common.AbortWithError(c, perr.InternalWithMessage(fmt.Sprintf("unable to ascertain step definition for step %s", sExec.ID)))
		return
	}

	stepType := sDef.GetType()
	stepFullName := sExec.Name
	stepName := strings.Split(stepFullName, ".")[len(strings.Split(stepFullName, "."))-1]

	switch stepType {
	case "input":
		output.Inputs[stepName] = webFormDataInputFromInputStep(sExec.Input)
	case "form":
		// TODO: implement
		common.AbortWithError(c, perr.InternalWithMessage("form is not yet implemented"))
		return
	default:
		common.AbortWithError(c, perr.InternalWithMessage(fmt.Sprintf("step type %s is not supported", stepType)))
		return
	}

	output.Status = sExec.Status
	output.ResponseURL, _ = url.JoinPath(util.GetBaseUrl(), c.Request.URL.Path, "submit")

	c.JSON(200, output)
}

func (api *APIService) postFormData(c *gin.Context) {
	var execID, pexecID, sexecID string
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

	if len(uri.ID) == 8 {
		e, p, s, ok := db.ResolveShortStepExecutionID(uri.ID)
		if !ok {
			common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution for %s not found", uri.ID)))
			return
		}
		execID = e
		pexecID = p
		sexecID = s
	} else {
		// TODO: Do we still require this code branch, should always use short ID
		ids := strings.Split(uri.ID, ".")
		if len(ids) != 3 {
			common.AbortWithError(c, perr.BadRequestWithMessage("unable to parse identifiers provided"))
			return
		}
		execID = ids[0]
		pexecID = ids[1]
		sexecID = ids[2]
	}

	ex, err := execution.GetExecution(execID)
	if err != nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("execution %s not found", execID)))
		return
	}

	pipelineExecution := ex.PipelineExecutions[pexecID]
	if pipelineExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution %s not found", pexecID)))
		return
	}

	stepExecution := pipelineExecution.StepExecutions[sexecID]
	if stepExecution == nil {
		common.AbortWithError(c, perr.NotFoundWithMessage(fmt.Sprintf("step execution %s not found", sexecID)))
		return
	}

	if pipelineExecution.IsFinished() || pipelineExecution.IsFinishing() || stepExecution.Status == "finished" {
		common.AbortWithError(c, perr.ConflictWithMessage(fmt.Sprintf("step %s has already been processed or is no longer required due to pipeline completion", sexecID)))
	}

	var parsedBody map[string]any
	switch c.ContentType() {
	case "application/x-www-form-urlencoded":
		// TODO: implement form encoded support for obtain body content
		common.AbortWithError(c, perr.InternalWithMessage("form-encoding not yet implemented"))
		return
	default:
		err = c.BindJSON(&parsedBody)
		if err != nil {
			common.AbortWithError(c, perr.InternalWithMessage("error parsing body content"))
			return
		}
	}

	stepType := strings.Split(stepExecution.Name, ".")[len(strings.Split(stepExecution.Name, "."))-2]
	stepFullName := stepExecution.Name
	stepName := strings.Split(stepFullName, ".")[len(strings.Split(stepFullName, "."))-1]
	switch stepType {
	case "input":
		if parsedBody[stepName] != nil {
			err := api.finishInputStepFromWebForm(execID, pexecID, stepExecution, parsedBody[stepName])
			if err != nil {
				common.AbortWithError(c, err)
				return
			}
			c.Status(200) // TODO: Return JSON payload
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
type webFormData struct {
	ExecutionID         string                      `json:"execution_id"`
	PipelineExecutionID string                      `json:"pipeline_execution_id"`
	StepExecutionID     string                      `json:"step_execution_id"`
	Status              string                      `json:"status"`
	ResponseURL         string                      `json:"response_url"`
	Inputs              map[string]webFormDataInput `json:"inputs"`
}

type webFormDataInput struct {
	Prompt    *string                   `json:"prompt,omitempty"`
	InputType *string                   `json:"input_type,omitempty"`
	Options   []webFormDataInputOptions `json:"options,omitempty"`
}

type webFormDataInputOptions struct {
	Label    *string `json:"label,omitempty"`
	Value    *string `json:"value,omitempty"`
	Selected *bool   `json:"selected,omitempty"`
}

func webFormDataFromId(id string) (webFormData, error) {
	var output webFormData
	output.Inputs = make(map[string]webFormDataInput)

	if len(id) == 8 {
		executionID, pipelineExecutionID, stepExecutionID, ok := db.ResolveShortStepExecutionID(id)
		if !ok {
			return output, perr.NotFoundWithMessage(fmt.Sprintf("pipeline execution %s not found", id))
		}
		output.ExecutionID = executionID
		output.PipelineExecutionID = pipelineExecutionID
		output.StepExecutionID = stepExecutionID
	} else {
		// TODO: Do we still require this code branch, should always use short ID
		ids := strings.Split(id, ".")
		if len(ids) != 3 {
			return output, perr.BadRequestWithMessage("unable to parse identifiers provided")
		}
		output.ExecutionID = ids[0]
		output.PipelineExecutionID = ids[1]
		output.StepExecutionID = ids[2]
	}

	return output, nil
}

func webFormDataInputFromInputStep(input modconfig.Input) webFormDataInput {
	var output webFormDataInput

	if p, ok := input[schema.AttributeTypePrompt].(string); ok {
		output.Prompt = &p
	}
	if t, ok := input[schema.AttributeTypeType].(string); ok {
		output.InputType = &t
	}
	if !helpers.IsNil(input[schema.AttributeTypeOptions]) {
		for _, o := range input[schema.AttributeTypeOptions].([]any) {
			opt := o.(map[string]any)
			option := webFormDataInputOptions{}
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
			output.Options = append(output.Options, option)
		}
	}

	return output
}

func (api *APIService) finishInputStepFromWebForm(execID, pexecID string, sexec *execution.StepExecution, value any) error {
	evt := &event.Event{ExecutionID: execID, CreatedAt: time.Now()}
	stepFinishedEvent, err := event.NewStepFinished()
	if err != nil {
		return perr.InternalWithMessage("unable to create step finished event: " + err.Error())
	}

	out := modconfig.Output{
		Data: map[string]any{
			"value": value,
		},
		Status: "finished",
	}

	stepFinishedEvent.Event = evt
	stepFinishedEvent.PipelineExecutionID = pexecID
	stepFinishedEvent.StepExecutionID = sexec.ID
	stepFinishedEvent.StepForEach = sexec.StepForEach
	stepFinishedEvent.StepLoop = sexec.StepLoop
	stepFinishedEvent.StepRetry = sexec.StepRetry
	stepFinishedEvent.StepOutput = make(map[string]any)
	stepFinishedEvent.Output = &out
	err = api.EsService.Raise(stepFinishedEvent)
	if err != nil {
		return perr.InternalWithMessage(fmt.Sprintf("error raising step finished event: %s", err.Error()))
	}
	return err
}
