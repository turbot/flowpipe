package filepaths

import (
	"fmt"
	"path"
	"path/filepath"

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

func ModDir() string {
	return viper.GetString(constants.ArgModLocation)
}

func FlowpipeDBFileName() string {
	modLocation := ModDir()
	dbPath := filepath.Join(modLocation, "flowpipe.db")
	return dbPath
}

func EventStoreFilePath(executionId string) string {
	return path.Join(EventStoreDir(), fmt.Sprintf("%s.jsonl", executionId))
}

func SnapshotFilePath(executionId string) string {
	return path.Join(EventStoreDir(), fmt.Sprintf("%s.sps", executionId))
}

func GlobalInternalDir() string {
	return path.Join(app_specific.InstallDir, "internal")
}
