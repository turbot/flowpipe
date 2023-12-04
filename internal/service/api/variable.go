package api

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/perr"
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
	logger := fplog.Logger(api.ctx)
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	rootMod := api.EsService.RootMod

	logger.Info("received list variable request", "next_token", nextToken, "limit", limit)

	variables := []types.Variable{}
	for _, v := range rootMod.ResourceMaps.Variables {
		variables = append(variables, types.Variable{
			Name:        v.ShortName,
			Description: v.Description,
			Type:        v.TypeString,
			Value:       v.ValueGo,
			Default:     v.DefaultGo,
		})
	}

	sort.Slice(variables, func(i, j int) bool {
		return variables[i].Name < variables[j].Name
	})

	result := types.ListVariableResponse{
		Items: variables,
	}

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

	rootMod := api.EsService.RootMod

	for _, v := range rootMod.ResourceMaps.Variables {
		if v.ShortName == uri.VariableName {
			result := types.Variable{
				Name:        v.ShortName,
				Type:        v.TypeString,
				Value:       v.ValueGo,
				Default:     v.DefaultGo,
				Description: v.Description,
			}
			c.JSON(http.StatusOK, result)
			return
		}
	}

	common.AbortWithError(c, perr.NotFoundWithMessage("not found"))
}
