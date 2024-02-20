package util

import (
	"fmt"
	"github.com/spf13/viper"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/pipe-fittings/constants"
	"net/url"
)

func GetHost() string {
	host := viper.GetString(constants.ArgHost)
	if host == "local" {
		return "http://localhost:7103"
	}

	return host
}

func GetBaseUrl() string {
	baseUrl := viper.GetString(constants.ArgBaseUrl)
	if baseUrl == "" {
		port := viper.GetInt(constants.ArgPort)
		if port == 0 {
			port = localconstants.DefaultServerPort
		}

		return fmt.Sprintf("http://localhost:%d", port)
	}
	return baseUrl
}

func GetWebformUrl(execId string, pExecId string, sExecId string) (string, error) {
	baseUrl := GetBaseUrl()
	joinId := fmt.Sprintf("%s.%s.%s", execId, pExecId, sExecId)
	salt, err := GetGlobalSalt()
	if err != nil {
		return "", err
	}
	hash := CalculateHash(joinId, salt)
	return url.JoinPath(baseUrl, "webform", "input", joinId, hash)
}
