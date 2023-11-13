package cmdconfig

import (
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
)

// TODO add in service defaults
var configDefaults = map[string]any{
	// TODO default max memory????
}

// environment variable mappings for directory paths which must be set as part of the viper bootstrap process
var dirEnvMappings = map[string]cmdconfig.EnvMapping{
	constants.EnvInstallDir:  {ConfigVar: []string{constants.ArgInstallDir}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvModLocation: {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
}

// NOTE: EnvWorkspaceProfile has already been set as a viper default as we have already loaded workspace profiles
// (EnvInstallDir has already been set at same time but we set it again to make sure it has the correct precedence)

// a map of known environment variables to map to viper keys - these are set as part of LoadGlobalConfig
var envMappings = map[string]cmdconfig.EnvMapping{
	constants.EnvInstallDir:  {ConfigVar: []string{constants.ArgInstallDir}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvModLocation: {ConfigVar: []string{constants.ArgModLocation}, VarType: cmdconfig.EnvVarTypeString},
	constants.EnvMemoryMaxMb: {ConfigVar: []string{constants.ArgMemoryMaxMb}, VarType: cmdconfig.EnvVarTypeInt},
}
