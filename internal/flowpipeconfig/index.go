package flowpipeconfig

import (
	"context"
	"log/slog"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/turbot/flowpipe/internal/resources"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/app_specific_connection"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/credential"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type FlowpipeConfig struct {
	ConfigPaths []string

	CredentialImports   map[string]credential.CredentialImport
	Credentials         map[string]credential.Credential
	Integrations        map[string]resources.Integration
	Notifiers           map[string]resources.Notifier
	ConnectionImports   map[string]modconfig.ConnectionImport
	PipelingConnections map[string]connection.PipelingConnection

	watcher                 *filewatcher.FileWatcher
	fileWatcherErrorHandler func(context.Context, error)

	// Hooks
	OnFileWatcherError func(context.Context, error)
	OnFileWatcherEvent func(context.Context, *FlowpipeConfig)

	loadLock *sync.Mutex
}

func (f *FlowpipeConfig) updateResources(other *FlowpipeConfig) {
	f.loadLock.Lock()
	defer f.loadLock.Unlock()

	f.CredentialImports = other.CredentialImports
	f.Credentials = other.Credentials
	f.Integrations = other.Integrations
	f.Notifiers = other.Notifiers
	f.PipelingConnections = other.PipelingConnections
	f.ConnectionImports = other.ConnectionImports

}

func (f *FlowpipeConfig) Equals(other *FlowpipeConfig) bool {
	if len(f.Credentials) != len(other.Credentials) {
		return false
	}

	for k, v := range f.Credentials {
		if _, ok := other.Credentials[k]; !ok {
			return false
		}

		if !other.Credentials[k].Equals(v) {
			return false
		}
	}

	if len(f.Integrations) != len(other.Integrations) {
		return false
	}

	for k, v := range f.Integrations {
		if _, ok := other.Integrations[k]; !ok {
			return false
		}

		if !other.Integrations[k].Equals(v) {
			return false
		}
	}

	if len(f.Notifiers) != len(other.Notifiers) {
		return false
	}

	for k, v := range f.Notifiers {
		if _, ok := other.Notifiers[k]; !ok {
			return false
		}

		if !other.Notifiers[k].Equals(v) {
			return false
		}
	}

	if len(f.CredentialImports) != len(other.CredentialImports) {
		return false
	}

	for k, v := range f.CredentialImports {

		if _, ok := other.CredentialImports[k]; !ok {
			return false
		}

		if !other.CredentialImports[k].Equals(v) {
			return false
		}
	}

	if len(f.PipelingConnections) != len(other.PipelingConnections) {
		return false
	}

	for k, v := range f.PipelingConnections {
		if _, ok := other.PipelingConnections[k]; !ok {
			return false
		}

		if !other.PipelingConnections[k].Equals(v) {
			return false
		}
	}
	if len(f.ConnectionImports) != len(other.ConnectionImports) {
		return false
	}
	for k, v := range f.ConnectionImports {
		if _, ok := other.ConnectionImports[k]; !ok {
			return false
		}
		if !other.ConnectionImports[k].Equals(v) {
			return false
		}
	}

	return true
}

func (f *FlowpipeConfig) SetupWatcher(ctx context.Context, errorHandler func(context.Context, error)) error {
	watcherOptions := &filewatcher.WatcherOptions{
		Directories: f.ConfigPaths,
		Include:     filehelpers.InclusionsFromExtensions([]string{app_specific.ConfigExtension}),
		ListFlag:    filehelpers.FilesRecursive,
		EventMask:   fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write,
		// we should look into passing the callback function into the underlying watcher
		// we need to analyze the kind of errors that come out from the watcher and
		// decide how to handle them
		// OnError: errCallback,
		OnChange: func(events []fsnotify.Event) {
			f.handleFileWatcherEvent(ctx)
		},
	}
	watcher, err := filewatcher.NewWatcher(watcherOptions)
	if err != nil {
		return err
	}
	f.watcher = watcher

	// start the watcher
	watcher.Start()

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	f.fileWatcherErrorHandler = errorHandler

	return nil
}

func (f *FlowpipeConfig) handleFileWatcherEvent(ctx context.Context) {
	slog.Debug("FlowpipeConfig handleFileWatcherEvent")

	newFpConfig, errAndWarnings := LoadFlowpipeConfig(f.ConfigPaths)

	if errAndWarnings.GetError() != nil {
		// call error hook
		if f.OnFileWatcherError != nil {
			f.OnFileWatcherError(ctx, errAndWarnings.Error)
		}

		// Flag on workspace?
		return
	}

	if !newFpConfig.Equals(f) {
		f.updateResources(newFpConfig)

		// call hook
		if f.OnFileWatcherEvent != nil {
			f.OnFileWatcherEvent(ctx, newFpConfig)
		}
	}

}

func (f *FlowpipeConfig) NotifierValueMap() (map[string]cty.Value, error) {
	varValueNotifierMap := make(map[string]cty.Value)
	if f == nil {
		return varValueNotifierMap, nil
	}

	for k, i := range f.Notifiers {
		var err error
		varValueNotifierMap[k], err = i.CtyValue()
		if err != nil {
			slog.Warn("failed to convert notifier to cty value", "notifier", i.Name(), "error", err)
		}
	}

	return varValueNotifierMap, nil
}

func NewFlowpipeConfig(configPaths []string) *FlowpipeConfig {
	defaultCreds, err := credential.DefaultCredentials()
	if err != nil {
		slog.Error("Unable to create default credentials", "error", err)
		return nil
	}

	defaultIntegrations, err := resources.DefaultIntegrations()
	if err != nil {
		slog.Error("Unable to create default integrations", "error", err)
		return nil
	}

	defaultNotifiers, err := resources.DefaultNotifiers(defaultIntegrations["http.default"])
	if err != nil {
		slog.Error("Unable to create default notifiers", "error", err)
		return nil
	}

	defaultPipelingConnections, err := app_specific_connection.DefaultPipelingConnections()
	if err != nil {
		slog.Error("Unable to create default pipeling connections", "error", err)
		return nil
	}

	fpConfig := FlowpipeConfig{
		CredentialImports:   make(map[string]credential.CredentialImport),
		Credentials:         defaultCreds,
		Integrations:        defaultIntegrations,
		Notifiers:           defaultNotifiers,
		ConfigPaths:         configPaths,
		PipelingConnections: defaultPipelingConnections,
		ConnectionImports:   make(map[string]modconfig.ConnectionImport),
		loadLock:            &sync.Mutex{},
	}

	return &fpConfig
}
