package manager

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	fpconstants "github.com/turbot/flowpipe/internal/constants"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/es/db"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/output"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/scheduler"
	"github.com/turbot/flowpipe/internal/store"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/cmdconfig"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/flowpipeconfig"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/sanitize"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
)

type ExecutionMode int

type StartupFlag int

const (
	startAPI       StartupFlag = 1 << iota // 1
	startES                                // 2
	startScheduler                         // 4
)

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	RootMod   *modconfig.Mod
	workspace *workspace.Workspace

	// Services
	ESService        *es.ESService
	apiService       *api.APIService
	schedulerService *scheduler.SchedulerService

	triggers map[string]*modconfig.Trigger

	HTTPAddress string
	HTTPPort    int

	startup StartupFlag

	Status    string
	StartedAt *time.Time
	StoppedAt *time.Time

	fpConfigLoadLock *sync.Mutex
	rootModLoadLock  *sync.Mutex
}

// NewManager creates a new Manager.
func NewManager(ctx context.Context, opts ...ManagerOption) *Manager {
	// Defaults
	m := &Manager{
		ctx:              ctx,
		Status:           "initialized",
		fpConfigLoadLock: &sync.Mutex{},
		rootModLoadLock:  &sync.Mutex{},
	}
	// Set options
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Start initializes tha manage and starts services managed by the Manager.
func (m *Manager) Start() (*Manager, error) {

	slog.Debug("Manager starting")
	defer slog.Debug("Manager started")

	if err := m.initializeModDirectory(); err != nil {
		return nil, err
	}

	// initializeResources - load and cache triggers and pipelines
	// if we are in server mode and there is a modfile, setup the file watcher
	if err := m.initializeResources(); err != nil {
		return nil, err
	}

	if m.shouldStartES() {
		err := m.startESService()
		if err != nil {
			return nil, err
		}
		for {
			slog.Info("Waiting for Flowpipe service to start ...")
			if m.ESService.IsRunning() {
				break
			}

			time.Sleep(time.Duration(100) * time.Millisecond)
		}

		slog.Info("Flowpipe service started ...")
	}

	if m.shouldStartAPI() {
		if err := m.startAPIService(); err != nil {
			return nil, err
		}
	}

	if m.shouldStartScheduler() {
		if err := m.startSchedulerService(); err != nil {
			return nil, err
		}
	}

	m.StartedAt = utils.TimeNow()
	m.Status = "running"

	if output.IsServerMode {
		m.renderServerStartOutput()
	}

	return m, nil
}

func (m *Manager) shouldStartAPI() bool {
	return m.startup&startAPI == startAPI
}

func (m *Manager) shouldStartES() bool {
	return m.startup&startES != 0
}

func (m *Manager) shouldStartScheduler() bool {
	return m.startup&startScheduler != 0
}

func (m *Manager) initializeModDirectory() error {
	modLocation := viper.GetString(constants.ArgModLocation)
	slog.Debug("Initializing mod directory", "modLocation", modLocation)

	modFlowpipeDir := path.Join(modLocation, app_specific.WorkspaceDataDir)
	err := util.EnsureDir(modFlowpipeDir)
	if err != nil {
		return err
	}

	internalDir := filepaths.ModInternalDir()
	modSaltPath := filepath.Join(internalDir, "salt")
	if files.DirectoryExists(internalDir) && files.FileExists(modSaltPath) {
		saltBytes, err := os.ReadFile(modSaltPath)
		if err != nil {
			return err
		}
		modSalt := string(saltBytes)
		if modSalt != "" {
			cache.GetCache().SetWithTTL("mod_salt", modSalt, 24*7*52*99*time.Hour)
		}
	}

	// Force cleanup if it hasn't run for 1 day
	store.ForceCleanup()

	return nil
}

// load and cache triggers and pipelines
// if we are in server mode and there is a modfile, setup the file watcher
func (m *Manager) initializeResources() error {
	modLocation := viper.GetString(constants.ArgModLocation)
	slog.Info("Starting Flowpipe", "modLocation", modLocation)

	var mod *modconfig.Mod

	var rootModName string

	if _, exists := parse.ModFileExists(modLocation); exists {
		// build the list of possible config path locations
		configPath, err := cmdconfig.GetConfigPath()
		error_helpers.FailOnError(err)

		flowpipeConfig, ew := flowpipeconfig.LoadFlowpipeConfig(configPath)

		// check for error
		error_helpers.FailOnError(ew.Error)
		ew.ShowWarnings()

		// Add the "Credentials" in the context
		// effectively forever .. we don't want to expire the config
		if flowpipeConfig != nil {
			cache.GetCache().SetWithTTL(fpconstants.FlowpipeConfigCacheKey, flowpipeConfig, 24*7*52*99*time.Hour)
			for _, c := range flowpipeConfig.Integrations {
				slog.Debug("Integration loaded", "name", c.GetHclResourceImpl().FullName)
			}

			for _, c := range flowpipeConfig.Credentials {
				if !strings.HasSuffix(c.GetHclResourceImpl().FullName, ".default") {
					slog.Debug("Credential loaded", "name", c.GetHclResourceImpl().FullName)
				}
			}

			for _, c := range flowpipeConfig.PipelingConnections {
				if !strings.HasSuffix(c.Name(), ".default") {
					slog.Debug("Connection loaded", "name", c.Name())
				}
			}
		}

		err = m.cacheConfigData()
		if err != nil {
			return err
		}

		if m.shouldStartAPI() {
			flowpipeConfig.OnFileWatcherEvent = m.flowpipeConfigUpdated
			err := flowpipeConfig.SetupWatcher(context.TODO(), func(c context.Context, e error) {
				slog.Error("error watching flowpipe config", "error", e)
			})

			if err != nil {
				return err
			}
		}

		err = m.loadMod()
		if err != nil {
			return err
		}

	} else {
		// there is no mod, just load pipelines and triggers from the directory
		var err error
		pipelines, triggers, err := load_mod.LoadPipelines(m.ctx, modLocation)
		if err != nil {
			return err
		}

		rootModName = "local"
		mod = &modconfig.Mod{
			ResourceMaps: &modconfig.ResourceMaps{
				Pipelines: pipelines,
				Triggers:  triggers,
			},
		}

		m.triggers = triggers
		m.RootMod = mod

		cache.GetCache().SetWithTTL("#rootmod.name", rootModName, 24*7*52*99*time.Hour)
		err = m.cacheModData(mod)
		if err != nil {
			return err
		}

		slog.Info("Pipelines and triggers loaded", "pipelines", len(pipelines), "triggers", len(triggers))
	}

	return nil
}

func (m *Manager) startESService() error {
	// start event sourcing service
	esService, err := es.NewESService(m.ctx)
	if err != nil {
		return err
	}
	err = esService.Start()
	if err != nil {
		return err
	}
	esService.Status = "running"
	esService.StartedAt = utils.TimeNow()
	esService.RootMod = m.RootMod

	m.ESService = esService
	return nil
}

func (m *Manager) startAPIService() error {
	// Define the API service
	apiService, err := api.NewAPIService(m.ctx, m.ESService,
		api.WithHTTPAddress(m.HTTPAddress),
		api.WithHTTPPort(m.HTTPPort))

	if err != nil {
		return err
	}
	m.apiService = apiService

	// Start API
	return apiService.Start()
}

func (m *Manager) startSchedulerService() error {
	s := scheduler.NewSchedulerService(m.ctx, m.ESService, m.triggers)
	if err := s.Start(); err != nil {
		slog.Error("error starting scheduler service", "error", err)
		return err
	}

	err := s.ScheduleCoreServices()
	if err != nil {
		return err
	}

	m.schedulerService = s
	return nil
}

// Stop stops services managed by the Manager.
func (m *Manager) Stop() error {
	slog.Debug("manager stopping")
	defer slog.Debug("manager stopped")

	// Ensure any log messages are synced before we exit
	defer func() {
		// TODO do we need this for slog
		// _ = slog.Sync()
	}()

	if m.apiService != nil {
		if err := m.apiService.Stop(); err != nil {
			// Log and continue stopping other services
			slog.Error("error stopping api service", "error", err)
		}
	}

	if m.ESService != nil {
		if err := m.ESService.Stop(); err != nil {
			// Log and continue stopping other services
			slog.Error("error stopping es service", "error", err)
		}
	}

	// Cleanup docker artifacts
	// TODO - Can we remove this since we cleanup per function etc?
	if docker.GlobalDockerClient != nil {
		if err := docker.GlobalDockerClient.CleanupArtifacts(); err != nil {
			slog.Error("Failed to cleanup flowpipe docker artifacts", "error", err)
		}
	}

	m.StoppedAt = utils.TimeNow()

	if output.IsServerMode {
		m.renderServerShutdownOutput()
	}

	return nil
}

func (m *Manager) InterruptHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		slog.Info("Manager exiting", "signal", sig)
		err := m.Stop()
		if err != nil {
			panic(err)
		}

		done <- true
	}()
	<-done
	slog.Debug("Manager exited")
}

func (m *Manager) cacheConfigData() error {

	fpConfig, err := db.GetFlowpipeConfig()
	if err != nil {
		return err
	}

	err = cacheHclResource("integration", fpConfig.Integrations, true, integrationUrlProcessor)
	if err != nil {
		return err
	}

	err = cacheHclResource("notifier", fpConfig.Notifiers, true, nil)
	if err != nil {
		return err
	}

	// Credential must be resolved at runtime, i.e. reading env var or temp creds
	err = cacheHclResource("credential", fpConfig.Credentials, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) cacheModData(mod *modconfig.Mod) error {

	err := cacheHclResource("pipeline", mod.ResourceMaps.Pipelines, true, nil)
	if err != nil {
		return err
	}

	triggers := mod.ResourceMaps.Triggers
	err = cacheHclResource("trigger", triggers, true, triggerUrlProcessor)
	if err != nil {
		return err
	}

	variables := mod.ResourceMaps.Variables
	err = cacheHclResource("variable", variables, true, nil)
	if err != nil {
		return err
	}

	return nil
}

func triggerUrlProcessor(trigger *modconfig.Trigger) error {
	if strings.HasPrefix(os.Getenv("RUN_MODE"), "TEST") {
		// don't calculate trigger url in test mode, there's no global salt and it's not needed
		return nil
	}

	_, ok := trigger.Config.(*modconfig.TriggerHttp)
	if ok {
		triggerUrl, err := calculateTriggerUrl(trigger)
		if err != nil {
			slog.Error("error calculating trigger url", "error", err)
			return err
		}
		trigger.Config.(*modconfig.TriggerHttp).Url = triggerUrl
	}

	return nil
}

func integrationUrlProcessor(integration modconfig.Integration) error {
	if strings.HasPrefix(os.Getenv("RUN_MODE"), "TEST") {
		// don't calculate trigger url in test mode, there's no global salt and it's not needed
		return nil
	}

	salt, err := util.GetGlobalSalt()
	if err != nil {
		slog.Error("salt not found", "error", err)
		return err
	}

	switch integration.GetIntegrationType() {
	case schema.IntegrationTypeSlack:
		integrationName := integration.GetHclResourceImpl().FullName
		hashString, err := util.CalculateHash(integrationName, salt)
		if err != nil {
			slog.Error("error computing hash", "error", err)
			return err
		}
		shortName := strings.TrimPrefix(integrationName, "slack.")
		integrationUrl := fmt.Sprintf("%s/api/latest/integration/slack/%s/%s", util.GetBaseUrl(), shortName, hashString)
		integration.SetUrl(integrationUrl)
	case schema.IntegrationTypeMsTeams:
		integrationName := integration.GetHclResourceImpl().FullName
		hashString, err := util.CalculateHash(integrationName, salt)
		if err != nil {
			slog.Error("error computing hash", "error", err)
			return err
		}
		shortName := strings.TrimPrefix(integrationName, "msteams.")
		integrationUrl := fmt.Sprintf("%s/api/latest/integration/msteams/%s/%s", util.GetBaseUrl(), shortName, hashString)
		integration.SetUrl(integrationUrl)
	}
	return nil
}

// TODO: rethink this approach. The cache is not deleted if the resource is removed. We also cache FlowpipeConfig in memory so this function is redundant
// TODO: except where we calculate the server side attribute such as URL
func cacheHclResource[T modconfig.HclResource](resourceType string, items map[string]T, cacheIndividualResource bool, individualProcessor func(T) error) error {
	inMemoryCache := cache.GetCache()

	var names []string
	for _, item := range items {
		name := item.Name()
		names = append(names, name)

		if individualProcessor != nil {
			err := individualProcessor(item)
			if err != nil {
				return err
			}
		}

		if cacheIndividualResource {
			inMemoryCache.SetWithTTL(name, item, 24*7*52*99*time.Hour)
		}
	}

	cacheName := fmt.Sprintf("#%s.names", resourceType)
	inMemoryCache.SetWithTTL(cacheName, names, 24*7*52*99*time.Hour)
	return nil
}

func calculateTriggerUrl(trigger *modconfig.Trigger) (string, error) {
	shortName := strings.TrimPrefix(trigger.UnqualifiedName, "trigger.")
	salt, err := util.GetModSaltOrDefault()
	if err != nil {
		return "", perr.InternalWithMessage("salt not found")
	}
	hashString, err := util.CalculateHash(shortName, salt)
	if err != nil {
		return "", perr.InternalWithMessage("error calculating hash")
	}
	baseUrl := util.GetBaseUrl()
	return fmt.Sprintf("%s/api/latest/hook/%s/%s", baseUrl, shortName, hashString), nil
}

func (m *Manager) renderServerStartOutput() {
	var outputs []sanitize.SanitizedStringer
	startTime := time.Now()
	if !helpers.IsNil(m.StartedAt) {
		startTime = *m.StartedAt
	}
	outputs = append(outputs, types.NewServerOutputStatusChange(startTime, "Started", app_specific.AppVersion.String()))
	outputs = append(outputs, types.NewServerOutputStatusChangeWithAdditional(startTime, "Listening", m.HTTPAddress, m.HTTPPort))
	if m.RootMod != nil {
		outputs = append(outputs, types.NewServerOutputLoaded(types.NewServerOutputPrefix(startTime, "flowpipe"), m.RootMod.Name(), false))
	}
	outputs = append(outputs, renderServerTriggers(m.triggers)...)
	outputs = append(outputs, types.NewServerOutput(startTime, "flowpipe", "Press Ctrl+C to exit"))

	output.RenderServerOutput(m.ctx, outputs...)
}

func (m *Manager) renderServerShutdownOutput() {
	stopTime := time.Now()
	if !helpers.IsNil(m.StoppedAt) {
		stopTime = *m.StoppedAt
	}
	output.RenderServerOutput(m.ctx, types.NewServerOutputStatusChange(stopTime, "Stopped", ""))
}

func renderServerTriggers(triggers map[string]*modconfig.Trigger) []sanitize.SanitizedStringer {
	var outputs []sanitize.SanitizedStringer

	for key, t := range triggers {
		tt := modconfig.GetTriggerTypeFromTriggerConfig(t.Config)
		prefix := types.NewServerOutputPrefix(time.Now(), "trigger")
		o := types.NewServerOutputTrigger(prefix, key, tt, t.Enabled)
		switch tt {
		case schema.TriggerTypeHttp:
			if tc, ok := t.Config.(*modconfig.TriggerHttp); ok {
				// TODO: Add Payload Requirements?
				methods := strings.Join(utils.SortedMapKeys(tc.Methods), " ")
				o.Method = &methods
				o.Url = &tc.Url
				outputs = append(outputs, o)
			}
		case schema.TriggerTypeSchedule:
			if tc, ok := t.Config.(*modconfig.TriggerSchedule); ok {
				o.Schedule = &tc.Schedule
				outputs = append(outputs, o)
			}
		case schema.TriggerTypeQuery:
			if tc, ok := t.Config.(*modconfig.TriggerQuery); ok {
				o.Schedule = &tc.Schedule
				o.Sql = &tc.Sql
				outputs = append(outputs, o)
			}
		}
	}

	return outputs
}
