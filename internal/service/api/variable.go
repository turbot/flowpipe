package api

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/es/db"
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
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /variable [get]
func (api *APIService) listVariables(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slog.Info("received list variable request", "next_token", nextToken, "limit", limit)
	result, err := ListVariables()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func ListVariables() (*types.ListVariableResponse, error) {

	variables, err := db.ListAllVariables()
	if err != nil {
		return nil, err
	}

	var fpVars []*types.FpVariable
	for _, v := range variables {
		fpVars = append(fpVars, types.FpVariableFromModVariable(v))
	}

	sort.Slice(fpVars, func(i, j int) bool {
		return fpVars[i].QualifiedName < fpVars[j].QualifiedName
	})

	result := types.ListVariableResponse{
		Items: fpVars,
	}

	return &result, nil

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
// @Success 200 {object} types.FpVariable
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /variable/{variable_name} [get]
func (api *APIService) getVariable(c *gin.Context) {

	var uri types.VariableRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	variableName := ConstructFullyQualifiedName("var", uri.VariableName)
	res, err := GetVariable(variableName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func GetVariable(name string) (*types.FpVariable, error) {
	variable, err := db.GetVariable(name)
	if err != nil {
		return nil, err
	}

	res := types.FpVariableFromModVariable(variable)
	return res, nil
}
