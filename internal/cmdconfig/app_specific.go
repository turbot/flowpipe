package cmdconfig

import (
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/error_helpers"
	"path/filepath"
	"strings"
)

// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {
	// set the default install dir
	installDir, err := files.Tildefy("~/.flowpipe")
	error_helpers.FailOnError(err)

	app_specific.DefaultInstallDir = installDir
	app_specific.DefaultConfigPath = strings.Join([]string{".", filepath.Join(installDir, "config")}, ":")
	app_specific.AppName = "flowpipe"
	// TODO unify version logic with steampipe and powerpipe
	//app_specific.AppVersion
	app_specific.AutoVariablesExtension = ".auto.fpvars"
	//app_specific.ClientConnectionAppNamePrefix
	//app_specific.ClientSystemConnectionAppNamePrefix

	app_specific.DefaultVarsFileName = "flowpipe.fpvars"
	//app_specific.DefaultWorkspaceDatabase
	app_specific.SetAppSpecificEnvVarKeys("FLOWPIPE_")
	app_specific.EnvInputVarPrefix = "FP_VAR_"

	app_specific.ConfigExtension = ".fpc"
	app_specific.ModDataExtension = ".fp"
	app_specific.ModFileName = "mod.fp"
	app_specific.VariablesExtension = ".fpvars"
	//app_specific.ServiceConnectionAppNamePrefix
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"
	// set the command pre and post hooks
	cmdconfig.CustomPreRunHook = preRunHook
}
