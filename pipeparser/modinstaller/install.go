package modinstaller

import (
	"context"

	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/utils"
	"github.com/turbot/go-kit/helpers"
)

func InstallWorkspaceDependencies(ctx context.Context, opts *InstallOpts) (_ *InstallData, err error) {
	utils.LogTime("cmd.InstallWorkspaceDependencies")
	defer func() {
		utils.LogTime("cmd.InstallWorkspaceDependencies end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	// install workspace dependencies
	installer, err := NewModInstaller(opts)
	if err != nil {
		return nil, err
	}

	if err := installer.InstallWorkspaceDependencies(ctx); err != nil {
		return nil, err
	}

	return installer.installData, nil
}
