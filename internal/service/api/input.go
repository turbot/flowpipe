package api

import (
	"fmt"
	"github.com/turbot/pipe-fittings/schema"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/es/execution"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) InputRegisterAPI(router *gin.RouterGroup) {
	router.GET("/input/:id/:hash", api.getInputStepInput)
}

type inputStepInput struct {
	ExecutionID         string `json:"execution_id"`
	PipelineExecutionID string `json:"pipeline_execution_id"`
	StepExecutionID     string `json:"step_execution_id"`
	Status              string `json:"status"`

	Prompt      *string                 `json:"prompt,omitempty"`
	InputType   *string                 `json:"input_type,omitempty"`
	Options     []inputStepInputOptions `json:"options,omitempty"`
	ResponseURL *string                 `json:"response_url,omitempty"`
}

type inputStepInputOptions struct {
	Label    *string `json:"label,omitempty"`
	Value    *string `json:"value,omitempty"`
	Selected *bool   `json:"selected,omitempty"`
}

func (api *APIService) getInputStepInput(c *gin.Context) {
	var output inputStepInput
	var uri types.InputIdHash
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
	hashString := util.CalculateHash(uri.Id, salt)
	if hashString != uri.Hash {
		common.AbortWithError(c, perr.UnauthorizedWithMessage("invalid hash"))
		return
	}

	// parse ids
	ids := strings.Split(uri.Id, ".")
	if len(ids) != 3 {
		common.AbortWithError(c, perr.BadRequestWithMessage("unable to parse identifiers provided"))
		return
	}
	output.ExecutionID = ids[0]
	output.PipelineExecutionID = ids[1]
	output.StepExecutionID = ids[2]

	// get step exec
	exec, err := execution.GetExecution(output.ExecutionID)
	if err != nil {
		common.AbortWithError(c, err)
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

	// verify step is type input
	sDef, err := exec.StepDefinition(pExec.ID, sExec.ID)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	if sDef.GetType() != "input" {
		common.AbortWithError(c, perr.InternalWithMessage(fmt.Sprintf("step %s is not input step type", sExec.ID)))
		return
	}

	// map status
	output.Status = sExec.Status

	// build response object
	if p, ok := sExec.Input[schema.AttributeTypePrompt].(string); ok {
		output.Prompt = &p
	}
	if t, ok := sExec.Input[schema.AttributeTypeType].(string); ok {
		output.InputType = &t
	}
	if !helpers.IsNil(sExec.Input[schema.AttributeTypeOptions]) {
		for _, o := range sExec.Input[schema.AttributeTypeOptions].([]any) {
			opt := o.(map[string]any)
			option := inputStepInputOptions{}
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

	// TODO: devise a better approach
	name := "integration.webform.default"
	hash := util.CalculateHash(name, salt)
	rUrl, _ := url.JoinPath(util.GetBaseUrl(), "api", "latest", "hook", name, hash)
	// rUrl := fmt.Sprintf("http://%s/api/latest/hook/%s/%s", c.Request.Host, name, hash)
	output.ResponseURL = &rUrl

	c.JSON(http.StatusOK, output)
}
