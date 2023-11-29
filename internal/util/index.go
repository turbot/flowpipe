package util

import (
	"path"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
)

func EventStoreDir() string {
	modLocation := viper.GetString(constants.ArgModLocation)
	modFlowpipeDir := path.Join(modLocation, app_specific.WorkspaceDataDir)
	eventStoreDir := path.Join(modFlowpipeDir, "store")

	return eventStoreDir
}

func ModInternalDir() string {
	modLocation := viper.GetString(constants.ArgModLocation)
	modFlowpipeDir := path.Join(modLocation, app_specific.WorkspaceDataDir)
	modInternalDir := path.Join(modFlowpipeDir, "internal")

	return modInternalDir
}
