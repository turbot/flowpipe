package cmdconfig

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/modconfig"
	"path/filepath"
)

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
)

func initGlobalConfig() {
	// load workspace profile from the configured install dir
	loader, err := cmdconfig.GetWorkspaceProfileLoader[*modconfig.FlowpipeWorkspaceProfile]()
	error_helpers.FailOnError(err)

	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	// set-up viper with defaults from the env and default workspace profile
	cmdconfig.BootstrapViper(loader, cmd,
		cmdconfig.WithConfigDefaults(configDefaults),
		cmdconfig.WithDirectoryEnvMappings(dirEnvMappings))

	error_helpers.FailOnError(err)

	installDir := viper.GetString(constants.ArgInstallDir)
	ensureInstallDir(filepath.Join(installDir, "internal"))

	saltDir := filepath.Join(filepaths.EnsureInternalDir(), "salt")
	salt, err := flowpipeSalt(saltDir, 32)
	error_helpers.FailOnError(err)

	cache.GetCache().SetWithTTL("salt", salt, 24*7*52*99*time.Hour)
}

// Assumes that the install dir exists
func flowpipeSalt(filename string, length int) (string, error) {
	// Check if the salt file exists
	if _, err := os.Stat(filename); err == nil {
		// If the file exists, read the salt from it
		saltBytes, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(saltBytes), nil
	}

	// If the file does not exist, generate a new salt
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// Encode the salt as a hexadecimal string
	saltHex := hex.EncodeToString(salt)

	// Write the salt to the file
	err = os.WriteFile(filename, []byte(saltHex), 0600)
	if err != nil {
		return "", err
	}

	return saltHex, nil
}

// todo KAI use filepaths???
func ensureInstallDir(installDir string) {
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}
	app_specific.InstallDir = installDir
}
