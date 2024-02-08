package api

import (
	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/flowpipe/templates"
	"github.com/turbot/pipe-fittings/perr"
	"html/template"
	"log/slog"
	"net/http"
)

func (api *APIService) InputRegisterAPI(router *gin.RouterGroup) {
	// router.POST("/input/:input/:hash", api.runInputPost)
	router.GET("/input/email/:input/:hash", api.runInputEmailGet)
}

type ParsedSlackResponse struct {
	Prompt   string
	UserName string
	Value    any
}

func (api *APIService) runInputEmailGet(c *gin.Context) {
	inputUri := types.InputRequestUri{}
	if err := c.ShouldBindUri(&inputUri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	_ = validateInputHash(inputUri)
	// TODO: uncomment if hash validation required
	// if err != nil {
	//   common.AbortWithError(c, err)
	//   return
	// }

	inputQuery := types.InputRequestQuery{}
	if err := c.ShouldBindQuery(&inputQuery); err != nil {
		common.AbortWithError(c, err)
		return
	}

	executionMode := "asynchronous"
	if inputQuery.ExecutionMode != nil {
		executionMode = *inputQuery.ExecutionMode
	}

	slog.Info("executionMode", "executionMode", executionMode)

	fired, err := api.finishInputStep(inputQuery.ExecutionID, inputQuery.PipelineExecutionID, inputQuery.StepExecutionID, inputQuery.Value)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	if !fired {
		alreadyAcknowledgedInputTemplate, err := templates.HTMLTemplate("already-acknowledged-input.html")
		if err != nil {
			slog.Error("error reading the template file", "error", err)
			common.AbortWithError(c, err)
			return
		}
		renderHTMLWithValues(c, string(alreadyAcknowledgedInputTemplate), gin.H{})
	} else {
		acknowledgeInputTemplate, err := templates.HTMLTemplate("acknowledge-input.html")
		if err != nil {
			slog.Error("error reading the template file", "error", err)
			common.AbortWithError(c, err)
			return
		}
		renderHTMLWithValues(c, string(acknowledgeInputTemplate), gin.H{"response": inputQuery.Value})
	}
}

func validateInputHash(inputUri types.InputRequestUri) error {
	inputName := inputUri.Input
	inputHash := inputUri.Hash

	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		slog.Error("salt not found")
		return perr.InternalWithMessage("salt not found")
	}

	hashString := util.CalculateHash(inputName, salt.(string))
	if hashString != inputHash {
		slog.Error("invalid hash", "hash", inputHash, "input_name", inputName, "expected", hashString)
		return perr.UnauthorizedWithMessage("invalid hash for " + inputName)
	}

	return nil
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
