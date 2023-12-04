package main

import (
	"context"

	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/cmd"
	localcmdconfig "github.com/turbot/flowpipe/internal/cmdconfig"
	"github.com/turbot/flowpipe/internal/sanitize"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
)

var (
	// These variables will be set by GoReleaser. We have them in main package because we put everything else in internal
	// and  I couldn't get Go Release to modify the internal packages
	version = "0.0.1-local.1"
	commit  = "none"
	date    = "unknown"
	builtBy = "local"
)

func main() {
	// Create a single, global context for the application
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	localcmdconfig.SetAppSpecificConstants()
	setupLogger()
	cache.InMemoryInitialize(nil)

	// TODO kai can we pass these into SetAppSpecificConstants?
	//  look into namespacing of config
	viper.SetDefault("main.version", version)
	viper.SetDefault("main.commit", commit)
	viper.SetDefault("main.date", date)
	viper.SetDefault("main.builtBy", builtBy)

	// Run the CLI
	cmd.RunCLI(ctx)
}

func setupLogger() {
	handlerOptions := &slog.HandlerOptions{
		Level: getLogLevel(),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			sanitized := sanitize.Instance.SanitizeKeyValue(a.Key, a.Value.Any())

			return slog.Attr{
				Key:   a.Key,
				Value: slog.AnyValue(sanitized),
			}
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOptions))
	slog.SetDefault(logger)
}

func getLogLevel() slog.Leveler {
	levelEnv := os.Getenv(app_specific.EnvLogLevel)

	switch strings.ToLower(levelEnv) {
	case "trace":
		return constants.LevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}
