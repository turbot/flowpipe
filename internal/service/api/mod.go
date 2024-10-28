package api

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) ModRegisterAPI(router *gin.RouterGroup) {

	router.GET("/mod/:mod_name", api.getMod)
}

// @Summary Get mod
// @Description Get mod
// @ID   mod_get
// @Tags Mod
// @Accept json
// @Produce json
// / ...
// @Param mod_name path string true "The name of the mod" format(^[a-z]{0,32}$)
// ...
// @Success 200 {object} types.Mod
// @Failure 400 {object} perr.ErrorModel
// @Failure 401 {object} perr.ErrorModel
// @Failure 403 {object} perr.ErrorModel
// @Failure 404 {object} perr.ErrorModel
// @Failure 429 {object} perr.ErrorModel
// @Failure 500 {object} perr.ErrorModel
// @Router /mod/{mod_name} [get]
func (api *APIService) getMod(c *gin.Context) {

	var uri types.ModRequestURI
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	rootMod := api.EsService.RootMod

	// TODO: need to be able to return the dependent mod?
	if rootMod.ShortName != uri.ModName {
		common.AbortWithError(c, perr.NotFoundWithMessage("not found"))
		return
	}

	fpMod := types.NewModFromModConfigMod(rootMod)

	c.JSON(http.StatusOK, fpMod)
}
