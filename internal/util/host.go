package util

import (
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
