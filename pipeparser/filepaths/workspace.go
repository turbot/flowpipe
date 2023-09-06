package filepaths

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/turbot/flowpipe/pipeparser/constants/runtime"
)

// mod related constants
const (
	WorkspaceDataDir            = ".steampipe"
	WorkspaceModDir             = "mods"
	WorkspaceModShadowDirPrefix = ".mods."
	WorkspaceConfigFileName     = "workspace.spc"
	WorkspaceIgnoreFile         = ".steampipeignore"
	ModFileName                 = "mod.sp"
	DefaultVarsFileName         = "steampipe.spvars"
	WorkspaceLockFileName       = ".mod.cache.json"
)

var PipesComponentWorkspaceDataDir = WorkspaceDataDir
var PipesComponentModsFileName = ModFileName
var PipesComponentWorkspaceIgnoreFiles = WorkspaceIgnoreFile
var PipesComponentDefaultVarsFileName = DefaultVarsFileName

var PipesComponentValidModFiles = []string{"mod.sp", "mod.hcl"}

func WorkspaceModPath(workspacePath string) string {
	return path.Join(workspacePath, PipesComponentWorkspaceDataDir, WorkspaceModDir)
}

func WorkspaceModShadowPath(workspacePath string) string {
	return path.Join(workspacePath, PipesComponentWorkspaceDataDir, fmt.Sprintf("%s%s", WorkspaceModShadowDirPrefix, runtime.ExecutionID))
}

func IsModInstallShadowPath(dirName string) bool {
	return strings.HasPrefix(dirName, WorkspaceModShadowDirPrefix)
}

func WorkspaceLockPath(workspacePath string) string {
	return path.Join(workspacePath, WorkspaceLockFileName)
}

func DefaultVarsFilePath(workspacePath string) string {
	return path.Join(workspacePath, PipesComponentDefaultVarsFileName)
}

func ModFilePath(modFolder string) string {
	modFilePath := filepath.Join(modFolder, PipesComponentModsFileName)
	return modFilePath
}
