package cmdconfig

import (
	serviceconfig "github.com/turbot/flowpipe/internal/service/config"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
)

// global config defaults
var configDefaults = map[string]any{}

// command specific config defaults (keyed by comand name)
var cmdSpecificDefaults = map[string]map[string]any{
	"server": serviceconfig.ConfigDefaults,
}

// environment variable mappings for directory paths which must be set as part of the viper bootstrap process
var dirEnvMappings = map[string]cmdconfig.EnvMapping{
	app_specific.EnvConfigPath:  {ConfigVar: []string{constants.ArgConfigPath}, VarType: cmdconfig.EnvVarTypeString},
	app_specific.EnvModLocation: {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
}

// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
// (EnvConfigPath has already been set at same time but we set it again to make sure it has the correct precedence)

// a map of known environment variables to map to viper keys - these are set as part of LoadGlobalConfig
var envMappings = map[string]cmdconfig.EnvMapping{
	app_specific.EnvConfigPath:  {ConfigVar: []string{constants.ArgConfigPath}, VarType: cmdconfig.EnvVarTypeString},
	app_specific.EnvModLocation: {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
	app_specific.EnvMemoryMaxMb: {ConfigVar: []string{constants.ArgMemoryMaxMb}, VarType: cmdconfig.EnvVarTypeInt},
}
