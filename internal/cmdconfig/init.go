package cmdconfig

import (
	"maps"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	constant "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func initGlobalConfig() *flowpipeconfig.FlowpipeConfig {
	ensureFlowpipeSample()

	// load workspace profile from the configured install dir
	loader, err := cmdconfig.GetWorkspaceProfileLoader[*modconfig.FlowpipeWorkspaceProfile]()
	error_helpers.FailOnError(err)

	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)

	// get the config defaults for this command
	configDefaults := getConfigDefaults(cmd)
	// set-up viper with defaults from the env and default workspace profile
	cmdconfig.BootstrapViper(loader, cmd,
		cmdconfig.WithConfigDefaults(configDefaults),
		cmdconfig.WithDirectoryEnvMappings(dirEnvMappings()))

	// set the rest of the defaults from ENV
	// ENV takes precedence over any default configuration
	cmdconfig.SetDefaultsFromEnv(envMappings())

	// if an explicit workspace profile was set, add to viper as highest precedence default
	// NOTE: if install_dir/mod_location are set these will already have been passed to viper by BootstrapViper
	// since the "ConfiguredProfile" is passed in through a cmdline flag, it will always take precedence
	if loader.ConfiguredProfile != nil {
		cmdconfig.SetDefaultsFromConfig(loader.ConfiguredProfile.ConfigMap(cmd))
	}

	validateConfig()

	// reset log level after reading the workspace config
	log.SetDefaultLogger()

	return nil
}

func validateConfig() {
	validOutputFormats := map[string]struct{}{
		constants.OutputFormatPretty: {},
		constants.OutputFormatPlain:  {},
		constants.OutputFormatJSON:   {},
		constants.OutputFormatYAML:   {},
	}
	output := viper.GetString(constants.ArgOutput)
	if _, ok := validOutputFormats[output]; !ok {
		error_helpers.FailOnError(perr.BadRequestWithMessage("invalid output format " + output))
	}

	modLocation := viper.GetString(constants.ArgModLocation)
	if _, err := os.Stat(modLocation); os.IsNotExist(err) {
		error_helpers.FailOnError(perr.BadRequestWithMessage("invalid mod location " + modLocation))
	}
}

// build defaults, combine global and cmd specific defaults
func getConfigDefaults(cmd *cobra.Command) map[string]any {
	var res = map[string]any{}
	maps.Copy(res, configDefaults())

	cmdSpecificDefaults, ok := cmdSpecificDefaults()[cmd.Name()]
	if ok {
		maps.Copy(res, cmdSpecificDefaults)
	}
	return res
}

func ensureFlowpipeSample() {
	sampleFile := filepath.Join(app_specific.InstallDir, "config", "flowpipe.fpc.sample")
	sampleFileContent := constant.FlowpipeSampleContent
	//nolint: gosec // this file is safe to be read by all users
	_ = os.WriteFile(sampleFile, []byte(sampleFileContent), 0755)
}
