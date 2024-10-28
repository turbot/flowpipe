//nolint:forbidigo // CLI command, expect some fmt.Println
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/flowpipe/internal/fperr"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/modinstaller"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/utils"
)

// mod management commands
func modCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "mod [command]",
		Args:  cobra.NoArgs,
		Short: "Flowpipe mod management",
		Long: `Flowpipe mod management.

Mods enable you to run, build, and share dashboards, benchmarks and other resources.

Find pre-built mods in the public registry at https://hub.flowpipe.io.

Examples:

    # Create a new mod in the current directory
    flowpipe mod init

    # Install a mod
    flowpipe mod install github.com/turbot/flowpipe-mod-github

    # Update a mod
    flowpipe mod update github.com/turbot/flowpipe-mod-github

    # List installed mods
    flowpipe mod list

    # Uninstall a mod
    flowpipe mod uninstall github.com/turbot/flowpipe-mod-github
	`,
	}

	cmd.AddCommand(modInstallCmd())
	cmd.AddCommand(modUninstallCmd())
	cmd.AddCommand(modUpdateCmd())
	cmd.AddCommand(modListCmd())
	cmd.AddCommand(modInitCmd())
	cmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for mod")

	return cmd
}

// install
func modInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "install",
		Run:   runModInstallCmd,
		Short: "Install one or more mods and their dependencies",
		Long:  `Install one or more mods and their dependencies.`,
	}

	// variable used to assign the output mode flag
	var updateStrategy = constants.ModUpdateIdLatest

	// setup hooks and flags
	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, false, "Help for init", cmdconfig.FlagOptions.WithShortHand("h")).
		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after update is complete").
		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be updated without modifying them").
		AddBoolFlag(constants.ArgForce, false, "Install mods even if cli version requirements are not met (cannot be used with --dry-run)").
		AddVarFlag(enumflag.New(&updateStrategy, constants.ArgPull, constants.ModUpdateStrategyIds, enumflag.EnumCaseInsensitive),
			constants.ArgPull,
			fmt.Sprintf("Update strategy; one of: %s", strings.Join(constants.FlagValues(constants.ModUpdateStrategyIds), ", ")))

	return cmd
}

func runModInstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			err := helpers.ToError(r)
			error_helpers.ShowError(ctx, err)
			os.Exit(constants.ExitCodeModInstallFailed)
		}
	}()

	// try to load the workspace mod definition
	// - if it does not exist, this will return a nil mod and a nil error
	workspacePath := viper.GetString(constants.ArgModLocation)
	workspaceMod, err := parse.LoadModfile(workspacePath)
	fperr.FailOnErrorWithMessage(err, "failed to load mod definition", nil, fperr.ErrorCodeModLoadFailed)

	// if no mod was loaded, create a default
	if workspaceMod == nil {
		workspaceMod, err = createWorkspaceMod(ctx, cmd, workspacePath)
		fperr.FailOnError(err, nil, fperr.ErrorCodeModLoadFailed)

	}

	// if any mod names were passed as args, convert into formed mod names
	installOpts := modinstaller.NewInstallOpts(workspaceMod, args...)
	installOpts.UpdateStrategy = viper.GetString(constants.ArgPull)

	slog.Debug("Mod install installOpts", "installOpts", installOpts)

	installData, err := modinstaller.InstallWorkspaceDependencies(ctx, installOpts)
	if err != nil {
		fperr.FailOnError(err, nil, fperr.ErrorCodeModInstallFailed)
	}

	fmt.Println(modinstaller.BuildInstallSummary(installData))
}

// uninstall
func modUninstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "uninstall",
		Run:   runModUninstallCmd,
		Short: "Uninstall a mod and its dependencies",
		Long:  `Uninstall a mod and its dependencies.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after uninstallation is complete").
		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be uninstalled without modifying them").
		AddBoolFlag(constants.ArgHelp, false, "Help for uninstall", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func runModUninstallCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModInstallCmd")
	defer func() {
		utils.LogTime("cmd.runModInstallCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			os.Exit(constants.ExitCodeUnknownErrorPanic)
		}
	}()

	// try to load the workspace mod definition
	// - if it does not exist, this will return a nil mod and a nil error
	workspaceMod, err := parse.LoadModfile(viper.GetString(constants.ArgModLocation))
	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")
	if workspaceMod == nil {
		fmt.Println("No mods installed.")
		return
	}
	opts := modinstaller.NewInstallOpts(workspaceMod, args...)
	installData, err := modinstaller.UninstallWorkspaceDependencies(ctx, opts)
	fperr.FailOnError(err, nil, "")
	fmt.Println(modinstaller.BuildUninstallSummary(installData))
}

// update
func modUpdateCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update",
		Run:   runModUpdateCmd,
		Short: "Update one or more mods and their dependencies",
		Long:  `Update one or more mods and their dependencies.`,
	}

	// variable used to assign the output mode flag
	var updateStrategy = constants.ModUpdateIdLatest

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgForce, false, "Update mods even if cli version requirements are not met (cannot be used with --dry-run)").
		AddBoolFlag(constants.ArgPrune, true, "Remove unused dependencies after update is complete").
		AddBoolFlag(constants.ArgDryRun, false, "Show which mods would be updated without modifying them").
		AddVarFlag(enumflag.New(&updateStrategy, constants.ArgPull, constants.ModUpdateStrategyIds, enumflag.EnumCaseInsensitive),
			constants.ArgPull,
			fmt.Sprintf("Update strategy; one of: %s", strings.Join(constants.FlagValues(constants.ModUpdateStrategyIds), ", "))).
		AddBoolFlag(constants.ArgHelp, false, "Help for update", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func runModUpdateCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModUpdateCmd")
	defer func() {
		utils.LogTime("cmd.runModUpdateCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			os.Exit(constants.ExitCodeUnknownErrorPanic)
		}
	}()

	// try to load the workspace mod definition
	// - if it does not exist, this will return a nil mod and a nil error
	workspaceMod, err := parse.LoadModfile(viper.GetString(constants.ArgModLocation))
	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")
	if workspaceMod == nil {
		fmt.Println("No mods installed.")
		return
	}

	installOpts := modinstaller.NewInstallOpts(workspaceMod, args...)
	installOpts.UpdateStrategy = viper.GetString(constants.ArgPull)

	slog.Debug("Mod update installOpts", "installOpts", installOpts)

	installData, err := modinstaller.InstallWorkspaceDependencies(ctx, installOpts)
	error_helpers.FailOnError(err)

	fmt.Println(modinstaller.BuildInstallSummary(installData))
}

// list
func modListCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Run:   runModListCmd,
		Short: "List currently installed mods",
		Long:  `List currently installed mods.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for list", cmdconfig.FlagOptions.WithShortHand("h"))
	return cmd
}

func runModListCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runModListCmd")
	defer func() {
		utils.LogTime("cmd.runModListCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			os.Exit(constants.ExitCodeUnknownErrorPanic)
		}
	}()

	// try to load the workspace mod definition
	// - if it does not exist, this will return a nil mod and a nil error
	workspaceMod, err := parse.LoadModfile(viper.GetString(constants.ArgModLocation))
	error_helpers.FailOnErrorWithMessage(err, "failed to load mod definition")
	if workspaceMod == nil {
		fmt.Println("No mods installed.")
		return
	}

	opts := modinstaller.NewInstallOpts(workspaceMod)
	installer, err := modinstaller.NewModInstaller(opts)
	error_helpers.FailOnError(err)

	treeString := installer.GetModList()
	if len(strings.Split(treeString, "\n")) > 1 {
		fmt.Println()
	}
	fmt.Println(treeString)
}

// // init
func modInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Run:   runModInitCmd,
		Short: "Initialize the current directory with a mod.sp file",
		Long:  `Initialize the current directory with a mod.sp file.`,
	}

	cmdconfig.OnCmd(cmd).AddBoolFlag(constants.ArgHelp, false, "Help for init", cmdconfig.FlagOptions.WithShortHand("h"))

	return cmd
}

func createWorkspaceMod(ctx context.Context, cmd *cobra.Command, workspacePath string) (*modconfig.Mod, error) {
	if !modinstaller.ValidateModLocation(ctx, workspacePath) {
		return nil, fmt.Errorf("mod %s cancelled", cmd.Name())
	}

	if _, exists := parse.ModFileExists(workspacePath); exists {
		fmt.Println("Working folder already contains a mod definition file")
		return nil, nil
	}
	mod := modconfig.CreateDefaultMod(workspacePath)
	if err := mod.Save(); err != nil {
		return nil, err
	}

	// load up the written mod file so that we get the updated
	// block ranges
	mod, err := parse.LoadModfile(workspacePath)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func runModInitCmd(cmd *cobra.Command, args []string) {
	workspacePath := viper.GetString(constants.ArgModLocation)
	_, err := createWorkspaceMod(cmd.Context(), cmd, workspacePath)
	if err != nil {
		slog.Error("Error creating mod", "error", err)
		os.Exit(1)
	}
}
