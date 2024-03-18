package util

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
	localconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/pipe-fittings/constants"
)

func GetHost() string {
	host := viper.GetString(constants.ArgHost)
	if host == "local" {
		return localconstants.DefaultFlowpipeHost
	}

	return host
}

func GetBaseUrl() string {
	baseUrl := viper.GetString(constants.ArgBaseUrl)
	if baseUrl == "" || baseUrl == localconstants.DefaultFlowpipeHost {
		port := viper.GetInt(constants.ArgPort)
		if port == 0 {
			port = localconstants.DefaultServerPort
		}

		return fmt.Sprintf("http://localhost:%d", port)
	}
	return baseUrl
}

func GetHttpFormUrl(execId string, pExecId string, sExecId string) (string, error) {

	if strings.HasPrefix(os.Getenv("RUN_MODE"), "TEST") {
		// in test env there's no global salt
		return localconstants.DefaultFlowpipeHost + "/form/" + sExecId + "/abcdefg", nil
	}

	baseUrl := GetBaseUrl()
	key := db.MapStepExecutionID(execId, pExecId, sExecId)

	salt, err := GetGlobalSalt()
	if err != nil {
		return "", err
	}
	hash, err := CalculateHash(key, salt)
	if err != nil {
		return "", err
	}
	return url.JoinPath(baseUrl, "form", key, hash)
}
