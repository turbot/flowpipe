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

func (api *APIService) NotifierRegisterAPI(router *gin.RouterGroup) {
	router.GET("/notifier", api.listNotifiers)
	router.GET("/notifier/:notifier_name", api.getNotifier)
}

// @Summary List notifiers
// @Description Lists notifiers
// @ID   notifier_list
// @Tags Notifier
// @Accept json
// @Produce json
// @Param limit query int false "The max number of items to fetch per page of data, subject to a min and max of 1 and 100 respectively. If not specified will default to 25." default(25) minimum(1) maximum(100)
// @Param next_token query string false "When list results are truncated, next_token will be returned, which is a cursor to fetch the next page of data. Pass next_token to the subsequent list request to fetch the next page of data."
// @Success 200 {object} types.ListNotifierResponse
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /notifier [get]
func (api *APIService) listNotifiers(c *gin.Context) {
	// Get paging parameters
	nextToken, limit, err := common.ListPagingRequest(c)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	slog.Info("received list notifiers request", "next_token", nextToken, "limit", limit)

	result, err := ListNotifiers(api.EsService.RootMod.Name())
	if err != nil {
		common.AbortWithError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func ListNotifiers(rootMod string) (*types.ListNotifierResponse, error) {
	notifiers, err := db.ListAllNotifiers()
	if err != nil {
		return nil, err
	}

	var listNotifierResponseItems []types.FpNotifier
	for _, notifier := range notifiers {
		item, err := types.FpNotifierFromModNotifier(notifier, rootMod)
		if err != nil {
			return nil, err
		}
		listNotifierResponseItems = append(listNotifierResponseItems, *item)
	}

	sort.Slice(listNotifierResponseItems, func(i, j int) bool {
		return listNotifierResponseItems[i].Name < listNotifierResponseItems[j].Name
	})

	return &types.ListNotifierResponse{
		Items: listNotifierResponseItems,
	}, nil
}

// @Summary Get notifier
// @Description Get notifier
// @ID   notifier_get
// @Tags Notifier
// @Accept json
// @Produce json
// @Param notifier_name path string true "Notifier name"
// @Success 200 {object} types.FpNotifier
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /notifier/{notifier_name} [get]
func (api *APIService) getNotifier(c *gin.Context) {
	notifierName := c.Param("notifier_name")
	slog.Info("received get notifier request", "notifier_name", notifierName)

	notifier, err := db.GetNotifier(notifierName)
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	item, err := types.FpNotifierFromModNotifier(notifier, api.EsService.RootMod.Name())
	if err != nil {
		common.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}
