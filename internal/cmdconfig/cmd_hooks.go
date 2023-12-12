package cmdconfig

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
)

// preRunHook is a function that is executed before the PreRun of every command handler
func preRunHook(cmd *cobra.Command, args []string) error {
	viper.Set(constants.ConfigKeyActiveCommand, cmd)
	viper.Set(constants.ConfigKeyActiveCommandArgs, args)

	// set up the global viper config with default values from
	// config files and ENV variables
	_ = initGlobalConfig()

	// set the max memory if specified
	setMemoryLimit(cmd.Context())

	// check telemetry setting
	telemetrySetting(cmd.Context())

	checkUpdateSetting(cmd.Context())

	return nil
}

func setMemoryLimit(ctx context.Context) {
	maxMemoryMb := viper.GetInt64(constants.ArgMemoryMaxMb)
	maxMemoryBytes := maxMemoryMb * 1024 * 1024
	if maxMemoryBytes > 0 {
		slog.Debug("setting memory limit", "max memory MB", maxMemoryMb)
		// set the max memory
		debug.SetMemoryLimit(maxMemoryBytes)
	}
}

func telemetrySetting(ctx context.Context) {
	telemetry := viper.GetBool(constants.ArgTelemetry)
	if telemetry {
		slog.Debug("enabling telemetry")
	}
}

func checkUpdateSetting(ctx context.Context) {
	updateCheck := viper.GetBool(constants.ArgUpdateCheck)
	if updateCheck {
		slog.Debug("enabling update check")
	}
}
