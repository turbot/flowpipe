package cmdconfig

import (
	"maps"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	con "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/log"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
)

func initGlobalConfig() *flowpipeconfig.FlowpipeConfig {
	// check defaults prior to loading into workspace
	defaultIntegrationNotifierFiles()

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

func defaultIntegrationNotifierFiles() {
	installPath := app_specific.DefaultInstallDir
	// configPath := strings.Split(app_specific.DefaultConfigPath, ":")[len(strings.Split(app_specific.DefaultConfigPath, ":"))-1]
	configPath := filepath.Join(installPath, "config")
	internalStateFile := filepath.Join(installPath, "internal", ".webform_initialized")
	integrationFile := filepath.Join(configPath, "integrations.fpc")
	notifierFile := filepath.Join(configPath, "notifiers.fpc")
	if !files.FileExists(internalStateFile) {
		if !files.FileExists(integrationFile) {
			_ = os.WriteFile(integrationFile, []byte(con.DefaultFlowpipeIntegrationContent), 0755)
		}
		if !files.FileExists(notifierFile) {
			_ = os.WriteFile(notifierFile, []byte(con.DefaultFlowpipeNotifierContent), 0755)
		}

		ts := time.Now().Format(time.RFC3339)
		_ = os.WriteFile(internalStateFile, []byte(ts), 0755)
	}
}
