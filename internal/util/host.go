package util

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
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
		port := viper.GetInt(constants.ArgPort)
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
	return filepath.Join(baseUrl, "webform", "input", joinId, hash), nil
}
