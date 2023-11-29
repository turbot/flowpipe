package cmdconfig

import (
	"fmt"
	"maps"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
)

func initGlobalConfig() *modconfig.FlowpipeConfig {
	// load workspace profile from the configured install dir
	loader, err := cmdconfig.GetWorkspaceProfileLoader[*modconfig.FlowpipeWorkspaceProfile]()
	error_helpers.FailOnError(err)

	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)

	// get the config defaults for this command
	configDefaults := getConfigDefaults(cmd)
	// set-up viper with defaults from the env and default workspace profile
	cmdconfig.BootstrapViper(loader, cmd,
		cmdconfig.WithConfigDefaults(configDefaults),
		cmdconfig.WithDirectoryEnvMappings(dirEnvMappings))

	// set the rest of the defaults from ENV
	// ENV takes precedence over any default configuration
	cmdconfig.SetDefaultsFromEnv(envMappings)

	// if an explicit workspace profile was set, add to viper as highest precedence default
	// NOTE: if install_dir/mod_location are set these will already have been passed to viper by BootstrapViper
	// since the "ConfiguredProfile" is passed in through a cmdline flag, it will always take precedence
	if loader.ConfiguredProfile != nil {
		cmdconfig.SetDefaultsFromConfig(loader.ConfiguredProfile.ConfigMap(cmd))
	}

	installDir := viper.GetString(constants.ArgInstallDir)
	ensureInstallDir(installDir)

	return nil
}

// build defaults, combine global and cmd specific defaults
func getConfigDefaults(cmd *cobra.Command) map[string]any {
	var res = map[string]any{}
	maps.Copy(res, configDefaults)

	cmdSpecificDefaults, ok := cmdSpecificDefaults[cmd.Name()]
	if ok {
		maps.Copy(res, cmdSpecificDefaults)
	}
	return res
}

// todo KAI use filepaths???
func ensureInstallDir(installDir string) {
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}
	app_specific.InstallDir = installDir
}
