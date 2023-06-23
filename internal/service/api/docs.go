package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/turbot/flowpipe/internal/service/api/common"
)

func (api *APIService) DocsRegisterAPI(router *gin.RouterGroup) {
	router.GET("/docs/openapi.json", api.getOpenAPIDocs)
}

func (api *APIService) getOpenAPIDocs(c *gin.Context) {
	file := "service/api/docs/openapi.json"

	b, err := os.ReadFile(file)
	if err != nil {
		common.AbortWithError(c, err)
	}

	var spec map[string]interface{}
	err = json.Unmarshal(b, &spec)
	if err != nil {
		common.AbortWithError(c, err)
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, spec)
}
