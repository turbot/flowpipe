package filepaths

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
)

func EventStoreDir() string {
	dataDir := viper.GetString(constants.ArgDataDir)
	if strings.Trim(dataDir, " ") != "" {
		return dataDir
	}

	modFlowpipeDir := ModFlowpipeDir()

	return modFlowpipeDir
}

func ModInternalDir() string {
	modFlowpipeDir := ModFlowpipeDir()
	modInternalDir := path.Join(modFlowpipeDir, "internal")

	return modInternalDir
}

func ModFlowpipeDir() string {
	modLocation := viper.GetString(constants.ArgModLocation)
	modFlowpipeDir := path.Join(modLocation, app_specific.WorkspaceDataDir)

	return modFlowpipeDir
}

func ModDir() string {
	return viper.GetString(constants.ArgModLocation)
}

func LegacyFlowpipeDBFileName() string {
	modLocation := ModDir()
	dbPath := filepath.Join(modLocation, "flowpipe.db")
	return dbPath
}

func FlowpipeDBFileName() string {

	dbPath := viper.GetString(constants.ArgDataDir)
	if strings.Trim(dbPath, " ") != "" {
		dbPath = filepath.Join(dbPath, "flowpipe.db")
		return dbPath
	}

	modLocation := ModFlowpipeDir()
	dbPath = filepath.Join(modLocation, "flowpipe.db")
	return dbPath
}

func GlobalInternalDir() string {
	return path.Join(app_specific.InstallDir, "internal")
}

func EventStoreFilePath(executionId string) string {
	return path.Join(EventStoreDir(), fmt.Sprintf("%s.jsonl", executionId))
}
