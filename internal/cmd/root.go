package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/pipe-fittings/parse"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/cmd/mod"
	"github.com/turbot/flowpipe/internal/cmd/pipeline"
	"github.com/turbot/flowpipe/internal/cmd/process"
	"github.com/turbot/flowpipe/internal/cmd/service"
	"github.com/turbot/flowpipe/internal/cmd/trigger"
	"github.com/turbot/flowpipe/internal/config"
	"github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/load_mod"

	pcconstants "github.com/turbot/pipe-fittings/constants"
)

// ④ Now use the FooMode enum flag. If you want a non-zero default, then
// simply set it here, such as in "foomode = Bar".
var outputMode types.OutputMode

// Build the cobra command that handles our command line tool.
func RootCommand(ctx context.Context) (*cobra.Command, error) {

	// Define our command
	rootCmd := &cobra.Command{
		Use:     constants.Name,
		Short:   constants.ShortDescription,
		Long:    constants.LongDescription,
		Version: viper.GetString("main.version"),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(pcconstants.ConfigKeyActiveCommand, cmd)

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
	rootCmd.PersistentFlags().String(constants.ArgApiHost, "http://localhost", "API server host")
	rootCmd.PersistentFlags().Int(constants.ArgApiPort, 7103, "API server port")
	rootCmd.PersistentFlags().Bool(constants.ArgTlsInsecure, false, "Skip TLS verification")

	// Common (steampipe, flowpipe) flags
	rootCmd.PersistentFlags().String(pcconstants.ArgInstallDir, filepaths.PipesComponentDefaultInstallDir, "Path to the Config Directory")
	rootCmd.PersistentFlags().String(pcconstants.ArgModLocation, cwd, "Path to the workspace working directory")

	// ⑤ Define the CLI flag parameters for your wrapped enum flag.
	rootCmd.PersistentFlags().Var(
		enumflag.New(&outputMode, constants.ArgOutput, types.OutputModeIds, enumflag.EnumCaseInsensitive),
		constants.ArgOutput,
		"Output format; one of: table, yaml, json")

	error_helpers.FailOnError(viper.BindPFlag(constants.ArgApiHost, rootCmd.PersistentFlags().Lookup(constants.ArgApiHost)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgApiPort, rootCmd.PersistentFlags().Lookup(constants.ArgApiPort)))
	error_helpers.FailOnError(viper.BindPFlag(constants.ArgTlsInsecure, rootCmd.PersistentFlags().Lookup(constants.ArgTlsInsecure)))

	error_helpers.FailOnError(viper.BindPFlag(pcconstants.ArgInstallDir, rootCmd.PersistentFlags().Lookup(pcconstants.ArgInstallDir)))
	error_helpers.FailOnError(viper.BindPFlag(pcconstants.ArgModLocation, rootCmd.PersistentFlags().Lookup(pcconstants.ArgModLocation)))

	// disable auto completion generation, since we don't want to support
	// powershell yet - and there's no way to disable powershell in the default generator
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// add all the subcommands
	err = addCommands(ctx, rootCmd)
	error_helpers.FailOnError(err)

	return rootCmd, nil
}

func addCommands(ctx context.Context, rootCmd *cobra.Command) error {
	// flowpipe service
	serviceCmd, err := service.ServiceCmd(ctx)
	if err != nil {
		return err
	}
	rootCmd.AddCommand(serviceCmd)

	pipelineCmd, err := pipeline.PipelineCmd(ctx)
	if err != nil {
		return err
	}
	rootCmd.AddCommand(pipelineCmd)

	triggerCmd, err := trigger.TriggerCmd(ctx)
	if err != nil {
		return err
	}
	rootCmd.AddCommand(triggerCmd)

	processCmd, err := process.ProcessCmd(ctx)
	if err != nil {
		return err
	}
	rootCmd.AddCommand(processCmd)

	modCmd, err := mod.ModCmd(ctx)
	if err != nil {
		return err
	}
	rootCmd.AddCommand(modCmd)

	return nil
}

// initConfig reads in config file and ENV variables if set.
func initGlobalConfig() *error_helpers.ErrorAndWarnings {

	// Steampipe CLI loads the Workspace Profile here, but it also loads the mod in the parse context.
	//
	// set global containing the configured install dir (create directory if needed)

	// load workspace profile from the configured install dir
	loader, err := parse.LoadWorkspaceProfiles(context.TODO())
	error_helpers.FailOnError(err)

	// set global workspace profile
	load_mod.GlobalWorkspaceProfile = loader.GetActiveWorkspaceProfile()

	var cmd = viper.Get(pcconstants.ConfigKeyActiveCommand).(*cobra.Command)
	// set-up viper with defaults from the env and default workspace profile
	err = load_mod.BootstrapViper(loader, cmd)
	error_helpers.FailOnError(err)

	installDir := viper.GetString(pcconstants.ArgInstallDir)
	ensureInstallDir(filepath.Join(installDir, "internal"))

	salt, err := flowpipeSalt(filepath.Join(installDir, filepaths.PipesComponentInternal, "salt"), 32)
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

	// store as SteampipeDir
	filepaths.SteampipeDir = installDir
}
