package cmdconfig

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/task"
	"github.com/turbot/pipe-fittings/utils"
)

var waitForTasksChannel chan struct{}
var tasksCancelFn context.CancelFunc

// postRunHook is a function that is executed after the PostRun of every command handler
func postRunHook(cmd *cobra.Command, args []string) error {
	utils.LogTime("cmdhook.postRunHook start")
	defer utils.LogTime("cmdhook.postRunHook end")

	if waitForTasksChannel != nil {
		// wait for the async tasks to finish
		select {
		case <-time.After(100 * time.Millisecond):
			tasksCancelFn()
			return nil
		case <-waitForTasksChannel:
			return nil
		}
	}

	return nil
}

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

	checkUpdate(cmd)

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

func checkUpdate(cmd *cobra.Command) {
	updateCheck := viper.GetBool(constants.ArgUpdateCheck)
	updateCheck = true //nolint:ineffassign // remove when we enable update check
	if updateCheck {
		// runScheduledTasks skips running tasks if this instance is the plugin manager
		waitForTasksChannel = runScheduledTasks(cmd.Context(), cmd, []string{})
	}
}

// runScheduledTasks runs the task runner and returns a channel which is closed when
// task run is complete
//
// runScheduledTasks skips running tasks if this instance is the plugin manager
func runScheduledTasks(ctx context.Context, cmd *cobra.Command, args []string) chan struct{} {
	taskUpdateCtx, cancelFn := context.WithCancel(ctx)
	tasksCancelFn = cancelFn

	return task.RunTasks(
		taskUpdateCtx,
		cmd,
		args,
		// pass the config value in rather than runRasks querying viper directly - to avoid concurrent map access issues
		// (we can use the update-check viper config here, since initGlobalConfig has already set it up
		// with values from the config files and ENV settings - update-check cannot be set from the command line)
		// task.WithUpdateCheck(viper.GetBool(constants.ArgUpdateCheck)),
		task.WithUpdateCheck(true),
	)
}
