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

func (api *APIService) IntegrationRegisterAPI(router *gin.RouterGroup) {
	router.GET("/integration", api.listIntegrations)
	router.GET("/integration/:integration_name", api.getIntegration)

	// integration specific handlers
	router.POST("/integration/slack/:id/:hash", api.slackPostHandler) // Slack
	router.POST("/integration/teams/:id/:hash", api.teamsPostHandler) // Teams
}

// @Summary List integrations
// @Description Lists integrations
// @ID   integration_list
// @Tags Integration
// @Accept json
// @Produce json
// / ...
// @Param limit query int false "The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25." default(25) minimum(1) maximum(100)
// @Param next_token query string false "When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data."
// ...
// @Success 200 {object} types.ListIntegrationResponse
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /integration [get]
func (api *APIService) listIntegrations(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slog.Info("received list integrations request", "next_token", nextToken, "limit", limit)

	result, err := ListIntegrations()
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func ListIntegrations() (*types.ListIntegrationResponse, error) {
	integrations, err := db.ListAllIntegrations()
	if err != nil {
		return nil, err
	}

	var listIntegrationResponseItems []types.FpIntegration

	for _, integration := range integrations {
		item, err := types.FpIntegrationFromModIntegration(integration)
		if err != nil {
			return nil, err
		}
		listIntegrationResponseItems = append(listIntegrationResponseItems, *item)
	}

	sort.Slice(listIntegrationResponseItems, func(i, j int) bool {
		return listIntegrationResponseItems[i].Name < listIntegrationResponseItems[j].Name
	})

	// TODO: paging, filter, sorting
	result := &types.ListIntegrationResponse{
		Items: listIntegrationResponseItems,
	}
	return result, nil
}

// @Summary Get integration
// @Description Get integration
// @ID   integration_get
// @Tags Integration
// @Accept json
// @Produce json
// / ...
// @Param integration_name path string true "The name of the integration" format(^[a-z_]{0,32}$)
// / ...
// @Success 200 {object} types.FpIntegration
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /integration/{integration_name} [get]
func (api *APIService) getIntegration(c *gin.Context) {
	var uri types.IntegrationRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	integration, err := db.GetIntegration(uri.IntegrationName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	integrationResponse, err := types.FpIntegrationFromModIntegration(integration)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, integrationResponse)
}

func GetIntegration(name string) (*types.FpIntegration, error) {
	integration, err := db.GetIntegration(name)
	if err != nil {
		return nil, err
	}
	return types.FpIntegrationFromModIntegration(integration)
}
