package flowpipe_mod_install_tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modinstaller"
	"github.com/turbot/pipe-fittings/parse"
)

// TestModReinstallOverReadOnlyPacks installs a mod and then force-reinstalls it. go-git
// (>= v5.17) writes git pack files (*.pack/*.idx) read-only; the reinstall must overwrite
// the previously-cloned packs via the mod installer's shadow-directory commit rather than
// fail with "permission denied". The first install always works (nothing to overwrite),
// which is why this upgrade/reinstall path stayed broken in flowpipe for months with no
// coverage. Requires network to clone the mod; skipped under -short.
func TestModReinstallOverReadOnlyPacks(t *testing.T) {
	if testing.Short() {
		t.Skip("clones a mod from github; skipped under -short")
	}

	workspace := t.TempDir()
	t.Setenv("FLOWPIPE_INSTALL_DIR", t.TempDir())

	viper.SetDefault("main.version", "0.0.0-test.0")
	viper.Set(constants.ArgModLocation, workspace)
	viper.Set(constants.ConfigKeyActiveCommand, &cobra.Command{Use: "install"})
	localcmdconfig.SetAppSpecificConstants()

	require.NoError(t, os.WriteFile(
		filepath.Join(workspace, "mod.fp"),
		[]byte("mod \"test_workspace\" {\n}\n"), 0o600))

	ctx := context.Background()
	const testMod = "github.com/turbot/flowpipe-mod-reallyfreegeoip@v1.0.0"

	// first install: clones the mod, leaving read-only pack files in the mods dir
	// and persisting the require to the workspace mod
	mod, err := parse.LoadModfile(workspace)
	require.NoError(t, err)
	_, err = modinstaller.InstallWorkspaceDependencies(ctx, modinstaller.NewInstallOpts(mod, testMod))
	require.NoError(t, err, "first install should succeed")

	// remove the install cache so the next install re-resolves and re-clones, then
	// re-install with no args: the shadow-directory commit copies the freshly-cloned
	// mod over the still-present read-only packs from the first install
	require.NoError(t, os.Remove(filepath.Join(workspace, ".mod.cache.json")))

	mod, err = parse.LoadModfile(workspace)
	require.NoError(t, err)
	_, err = modinstaller.InstallWorkspaceDependencies(ctx, modinstaller.NewInstallOpts(mod))
	require.NoError(t, err, "reinstall over read-only git packs must not fail with permission denied")
}
