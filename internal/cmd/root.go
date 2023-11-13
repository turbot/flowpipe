package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/turbot/pipe-fittings/modconfig"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/config"
	internalconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
)

// ④ Now use the FooMode enum flag. If you want a non-zero default, then
// simply set it here, such as in "foomode = Bar".
var outputMode types.OutputMode

// Build the cobra command that handles our command line tool.
func RootCommand(ctx context.Context) (*cobra.Command, error) {

	// Define our command
	rootCmd := &cobra.Command{
		Use:     internalconstants.Name,
		Short:   internalconstants.ShortDescription,
		Long:    internalconstants.LongDescription,
		Version: viper.GetString("main.version"),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(constants.ConfigKeyActiveCommand, cmd)

			// set up the global viper config with default values from
			// config files and ENV variables

			// TODO: this creates '~' directory in the source when we run the test. Find a solution.
			_ = initGlobalConfig()

			return nil
		},
	}
	rootCmd.SetVersionTemplate("Flowpipe v{{.Version}}\n")

	c := config.GetConfigFromContext(ctx)

	cwd, err := os.Getwd()
	error_helpers.FailOnError(err)

	rootCmd.Flags().StringVar(&c.ConfigPath, "config-path", "", "config file (default is $HOME/.flowpipe/flowpipe.yaml)")

	// Flowpipe API
	rootCmd.PersistentFlags().String(internalconstants.ArgApiHost, "http://localhost", "API server host")
	rootCmd.PersistentFlags().Int(internalconstants.ArgApiPort, 7103, "API server port")
	rootCmd.PersistentFlags().Bool(internalconstants.ArgTlsInsecure, false, "Skip TLS verification")

	// Common (steampipe, flowpipe) flags
	rootCmd.PersistentFlags().String(constants.ArgInstallDir, app_specific.DefaultInstallDir, "Path to the Config Directory")
	rootCmd.PersistentFlags().String(constants.ArgModLocation, cwd, "Path to the workspace working directory")

	// ⑤ Define the CLI flag parameters for your wrapped enum flag.
	rootCmd.PersistentFlags().Var(
		enumflag.New(&outputMode, constants.ArgOutput, types.OutputModeIds, enumflag.EnumCaseInsensitive),
		constants.ArgOutput,
		"Output format; one of: table, yaml, json")

	error_helpers.FailOnError(viper.BindPFlag(internalconstants.ArgApiHost, rootCmd.PersistentFlags().Lookup(internalconstants.ArgApiHost)))
	error_helpers.FailOnError(viper.BindPFlag(internalconstants.ArgApiPort, rootCmd.PersistentFlags().Lookup(internalconstants.ArgApiPort)))
	error_helpers.FailOnError(viper.BindPFlag(internalconstants.ArgTlsInsecure, rootCmd.PersistentFlags().Lookup(internalconstants.ArgTlsInsecure)))

	error_helpers.FailOnError(viper.BindPFlag(constants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(constants.ArgInstallDir)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgModLocation, rootCmd.PersistentFlags().Lookup(constants.ArgModLocation)))

	// disable auto completion generation, since we don't want to support
	// powershell yet - and there's no way to disable powershell in the default generator
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// add all the subcommands
	err = addCommands(rootCmd)
	error_helpers.FailOnError(err)

	return rootCmd, nil
}

func addCommands(rootCmd *cobra.Command) error {
	rootCmd.AddCommand(serviceCmd())
	rootCmd.AddCommand(pipelineCmd())
	rootCmd.AddCommand(triggerCmd())
	rootCmd.AddCommand(processCmd())
	rootCmd.AddCommand(modCmd())

	return nil
}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() *error_helpers.ErrorAndWarnings {
	// load workspace profile from the configured install dir
	loader, err := cmdconfig.GetWorkspaceProfileLoader[*modconfig.FlowpipeWorkspaceProfile]()
	error_helpers.FailOnError(err)

	var cmd = viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	// set-up viper with defaults from the env and default workspace profile
	err = cmdconfig.BootstrapViper(loader, cmd)
	error_helpers.FailOnError(err)

	installDir := viper.GetString(constants.ArgInstallDir)
	ensureInstallDir(filepath.Join(installDir, "internal"))

	saltDir := filepath.Join(filepaths.EnsureInternalDir(), "salt")
	salt, err := flowpipeSalt(saltDir, 32)
	if err != nil {
		error_helpers.FailOnErrorWithMessage(err, err.Error())
	}

	cache.GetCache().SetWithTTL("salt", salt, 24*7*52*99*time.Hour)

	return nil
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

func ensureInstallDir(installDir string) {
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		err = os.MkdirAll(installDir, 0755)
		error_helpers.FailOnErrorWithMessage(err, fmt.Sprintf("could not create installation directory: %s", installDir))
	}
	app_specific.InstallDir = installDir
}
