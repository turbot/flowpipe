package cmdconfig

import (
	serviceconfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
)

// global config defaults
func configDefaults() map[string]any {
	return map[string]any{
		constants.ArgMemoryMaxMb: 1024,
		constants.ArgTelemetry:   constants.TelemetryInfo,
		constants.ArgUpdateCheck: true,
		constants.ArgInstallDir:  app_specific.DefaultInstallDir,
	}
}

// command specific config defaults (keyed by comand name)
func cmdSpecificDefaults() map[string]map[string]any {
	return map[string]map[string]any{
		"server": serviceconfig.ConfigDefaults,
	}
}

// environment variable mappings for directory paths which must be set as part of the viper bootstrap process
func dirEnvMappings() map[string]cmdconfig.EnvMapping {
	return map[string]cmdconfig.EnvMapping{
		app_specific.EnvConfigPath:  {ConfigVar: []string{constants.ArgConfigPath}, VarType: cmdconfig.EnvVarTypeString},
		app_specific.EnvModLocation: {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
	}
}

// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
// (EnvConfigPath has already been set at same time but we set it again to make sure it has the correct precedence)

// a map of known environment variables to map to viper keys - these are set as part of LoadGlobalConfig
func envMappings() map[string]cmdconfig.EnvMapping {
	return map[string]cmdconfig.EnvMapping{
		app_specific.EnvConfigPath:           {ConfigVar: []string{constants.ArgConfigPath}, VarType: cmdconfig.EnvVarTypeString},
		app_specific.EnvModLocation:          {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
		app_specific.EnvMemoryMaxMb:          {ConfigVar: []string{constants.ArgMemoryMaxMb}, VarType: cmdconfig.EnvVarTypeInt},
		app_specific.EnvTelemetry:            {ConfigVar: []string{constants.ArgTelemetry}, VarType: cmdconfig.EnvVarTypeInt},
		app_specific.EnvUpdateCheck:          {ConfigVar: []string{constants.ArgUpdateCheck}, VarType: cmdconfig.EnvVarTypeBool},
		"FLOWPIPE_MAX_CONCURRENCY_HTTP":      {ConfigVar: []string{constants.ArgMaxConcurrencyHttp}, VarType: cmdconfig.EnvVarTypeInt},
		"FLOWPIPE_MAX_CONCURRENCY_QUERY":     {ConfigVar: []string{constants.ArgMaxConcurrencyQuery}, VarType: cmdconfig.EnvVarTypeInt},
		"FLOWPIPE_MAX_CONCURRENCY_CONTAINER": {ConfigVar: []string{constants.ArgMaxConcurrencyContainer}, VarType: cmdconfig.EnvVarTypeInt},
		"FLOWPIPE_MAX_CONCURRENCY_FUNCTION":  {ConfigVar: []string{constants.ArgMaxConcurrencyFunction}, VarType: cmdconfig.EnvVarTypeInt},
	}
}
