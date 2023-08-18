package pipeparser

import (
	"context"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/pipeparser/cmdconfig"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/filepaths"
	filehelpers "github.com/turbot/go-kit/files"
)

func LoadWorkspaceProfile(runCtx context.Context) (*WorkspaceProfileLoader, error) {
	// set viper default for workspace profile, using STEAMPIPE_WORKSPACE env var
	SetDefaultFromEnv(constants.EnvWorkspaceProfile, constants.ArgWorkspaceProfile, cmdconfig.String)
	// set viper default for install dir, using STEAMPIPE_INSTALL_DIR env var
	SetDefaultFromEnv(constants.EnvInstallDir, constants.ArgInstallDir, cmdconfig.String)

	// resolve the workspace profile dir
	installDir, err := filehelpers.Tildefy(viper.GetString(constants.ArgInstallDir))
	if err != nil {
		return nil, err
	}

	workspaceProfileDir, err := filepaths.WorkspaceProfileDir(installDir)
	if err != nil {
		return nil, err
	}

	// create loader
	loader, err := NewWorkspaceProfileLoader(runCtx, workspaceProfileDir)
	if err != nil {
		return nil, err
	}

	return loader, nil
}
