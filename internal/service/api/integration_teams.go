package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/service/api/common"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/perr"
)

func (api *APIService) teamsPostHandler(c *gin.Context) {
	var uri types.InputIDHash
	if err := c.ShouldBindUri(&uri); err != nil {
		common.AbortWithError(c, err)
		return
	}

	// support for omitting the type in the url
	if !strings.HasPrefix(uri.ID, "teams.") {
		uri.ID = fmt.Sprintf("teams.%s", uri.ID)
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

	// TODO: #TeamsIntegrationImplementation - handler

	c.Status(http.StatusNotImplemented)
}
