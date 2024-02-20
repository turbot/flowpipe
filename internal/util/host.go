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
		host := viper.GetString(constants.ArgListen)
		// when running in CLI mode, the default ArgListen is not bound to Viper (because it's part of the server command, not the root command)
		if host == "" {
			host = localconstants.DefaultListen
		}
		port := viper.GetInt(constants.ArgPort)
		if port == 0 {
			port = localconstants.DefaultServerPort
		}

		return fmt.Sprintf("http://%s:%d", host, port)
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
