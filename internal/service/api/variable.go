package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
)

func (api *APIService) VariableRegisterAPI(router *gin.RouterGroup) {
	router.GET("/variable", api.listVariables)
	router.GET("/variable/:variable_name", api.getVariable)
}

// @Summary List variables
// @Description Lists variables
// @ID   variable_list
// @Tags Variable
// @Accept json
// @Produce json
// / ...
// @Param limit query int false "The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25." default(25) minimum(1) maximum(100)
// @Param next_token query string false "When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data."
// ...
// @Success 200 {object} types.ListVariableResponse
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
// @Router /variable [get]
func (api *APIService) listVariables(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	fplog.Logger(api.ctx).Info("received list variable request", "next_token", nextToken, "limit", limit)

	result := types.ListVariableResponse{
		Items: []types.Variable{},
	}

	result.Items = append(result.Items, types.Variable{Type: "variable_webhook", Name: "webhookvariable"}, types.Variable{Type: "variable_manual", Name: "manualvariable"})

	c.JSON(http.StatusOK, result)
}

// @Summary Get variable
// @Description Get variable
// @ID   variable_get
// @Tags Variable
// @Accept json
// @Produce json
// / ...
// @Param variable_name path string true "The name of the variable" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.Variable
// @Failure 400 {object} pcerr.ErrorModel
// @Failure 401 {object} pcerr.ErrorModel
// @Failure 403 {object} pcerr.ErrorModel
// @Failure 404 {object} pcerr.ErrorModel
// @Failure 429 {object} pcerr.ErrorModel
// @Failure 500 {object} pcerr.ErrorModel
// @Router /variable/{variable_name} [get]
func (api *APIService) getVariable(c *gin.Context) {

	var uri types.VariableRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}
	result := types.Variable{Type: "variable_" + uri.VariableName, Name: uri.VariableName}
	c.JSON(http.StatusOK, result)
}
