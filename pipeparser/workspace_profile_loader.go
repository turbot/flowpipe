package pipeparser

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/utils"
)

var GlobalWorkspaceProfile *modconfig.WorkspaceProfile

type WorkspaceProfileLoader struct {
	workspaceProfiles    map[string]*modconfig.WorkspaceProfile
	workspaceProfilePath string
	DefaultProfile       *modconfig.WorkspaceProfile
	ConfiguredProfile    *modconfig.WorkspaceProfile
}

func NewWorkspaceProfileLoader(runCtx context.Context, workspaceProfilePath string) (*WorkspaceProfileLoader, error) {
	loader := &WorkspaceProfileLoader{workspaceProfilePath: workspaceProfilePath}
	workspaceProfiles, err := loader.load(runCtx)
	if err != nil {
		return nil, err
	}
	loader.workspaceProfiles = workspaceProfiles

	defaultProfile, err := loader.get("default")
	if err != nil {
		// there must always be a default - this should have been added by parse.LoadWorkspaceProfiles
		return nil, err
	}
	loader.DefaultProfile = defaultProfile

	if viper.IsSet(constants.ArgWorkspaceProfile) {
		configuredProfile, err := loader.get(viper.GetString(constants.ArgWorkspaceProfile))
		if err != nil {
			// could not find configured profile
			return nil, err
		}
		loader.ConfiguredProfile = configuredProfile
	}

	return loader, nil
}

func (l *WorkspaceProfileLoader) GetActiveWorkspaceProfile() *modconfig.WorkspaceProfile {
	if l.ConfiguredProfile != nil {
		return l.ConfiguredProfile
	}
	return l.DefaultProfile
}

func (l *WorkspaceProfileLoader) get(name string) (*modconfig.WorkspaceProfile, error) {
	if workspaceProfile, ok := l.workspaceProfiles[name]; ok {
		return workspaceProfile, nil
	}

	if implicitWorkspace := l.getImplicitWorkspace(name); implicitWorkspace != nil {
		return implicitWorkspace, nil
	}

	return nil, fmt.Errorf("workspace profile %s does not exist", name)
}

func (l *WorkspaceProfileLoader) load(runCtx context.Context) (map[string]*modconfig.WorkspaceProfile, error) {
	// get all the config files in the directory
	return parse.LoadWorkspaceProfiles(runCtx, l.workspaceProfilePath)
}

/*
Named workspaces follow normal standards for hcl identities, thus they cannot contain the slash (/) character.

If you pass a value to --workspace or STEAMPIPE_WORKSPACE in the form of {identity_handle}/{workspace_handle},
it will be interpreted as an implicit workspace.

Implicit workspaces, as the name suggests, do not need to be specified in the workspaces.spc file.

Instead they will be assumed to refer to a Steampipe Cloud workspace,
which will be used as both the database and snapshot location.

Essentially, --workspace acme/dev is equivalent to:

	workspace "acme/dev" {
	  workspace_database = "acme/dev"
	  snapshot_location  = "acme/dev"
	}
*/
func (l *WorkspaceProfileLoader) getImplicitWorkspace(name string) *modconfig.WorkspaceProfile {
	if IsCloudWorkspaceIdentifier(name) {
		log.Printf("[TRACE] getImplicitWorkspace - %s is implicit workspace: SnapshotLocation=%s, WorkspaceDatabase=%s", name, name, name)
		return &modconfig.WorkspaceProfile{
			SnapshotLocation:  utils.ToStringPointer(name),
			WorkspaceDatabase: utils.ToStringPointer(name),
		}
	}
	return nil
}
