package cmdconfig

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/error_helpers"
)


// SetAppSpecificConstants sets app specific constants defined in pipe-fittings
func SetAppSpecificConstants() {
	app_specific.AppName = "flowpipe"

	// TODO unify version logic with steampipe and powerpipe
	//app_specific.AppVersion

	// set all app specific env var keys
	app_specific.SetAppSpecificEnvVarKeys("FLOWPIPE_")

	app_specific.AutoVariablesExtension = ".auto.fpvars"
	app_specific.DefaultVarsFileName = "flowpipe.fpvars"
	app_specific.EnvInputVarPrefix = "FP_VAR_"

	app_specific.ConfigExtension = ".fpc"
	app_specific.ModDataExtension = ".fp"
	app_specific.ModFileName = "mod.fp"
	app_specific.VariablesExtension = ".fpvars"
	app_specific.WorkspaceIgnoreFile = ".flowpipeignore"
	app_specific.WorkspaceDataDir = ".flowpipe"

	// set the command pre and post hooks
	cmdconfig.CustomPreRunHook = preRunHook
	cmdconfig.CustomPostRunHook = postRunHook

	// set the default install dir
	defaultInstallDir, err := files.Tildefy("~/.flowpipe")
	error_helpers.FailOnError(err)
	app_specific.DefaultInstallDir = defaultInstallDir

	// set the default config path
	globalConfigPath := filepath.Join(defaultInstallDir, "config")
	// check whether install-dir env has been set - if so, respect it
	if envInstallDir, ok := os.LookupEnv(app_specific.EnvInstallDir); ok {
		globalConfigPath = filepath.Join(envInstallDir, "config")
	}
	app_specific.DefaultConfigPath = strings.Join([]string{".", globalConfigPath}, ":")

	// Version check
	app_specific.VersionCheckHost = "hub.flowpipe.io"
	app_specific.VersionCheckPath = "api/cli/version/latest"


}
