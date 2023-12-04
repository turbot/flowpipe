package manager

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/cache"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/filepaths"
	"github.com/turbot/flowpipe/internal/service/api"
	"github.com/turbot/flowpipe/internal/service/es"
	"github.com/turbot/flowpipe/internal/service/scheduler"
	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/steampipeconfig"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/pipe-fittings/workspace"
)

type ExecutionMode int

type StartupFlag int

const (
	startAPI       StartupFlag = 1 << iota // 1
	startES                                // 2
	startScheduler                         // 4
	startDocker                            // 8
)

// Manager manages and represents the status of the service.
type Manager struct {
	ctx context.Context

	RootMod *modconfig.Mod

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
}

// NewManager creates a new Manager.
func NewManager(ctx context.Context, opts ...ManagerOption) *Manager {
	// Defaults
	m := &Manager{
		ctx:    ctx,
		Status: "initialized",
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

	if m.shouldStartDocker() {
		if err := docker.Initialize(m.ctx); err != nil {
			return nil, err
		}
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

	return m, nil
}

func (m *Manager) shouldStartAPI() bool {
	return m.startup&startAPI == startAPI
}

func (m *Manager) shouldStartES() bool {
	return m.startup&startES != 0
}

func (m *Manager) shouldStartDocker() bool {
	return m.startup&startDocker != 0
}

func (m *Manager) shouldStartScheduler() bool {
	return m.startup&startScheduler != 0
}

func ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return perr.InternalWithMessage(fmt.Sprintf("error creating directory %s", dir))
		}
	}
	return nil
}

func (m *Manager) initializeModDirectory() error {

	modLocation := viper.GetString(constants.ArgModLocation)
	slog.Debug("Initializing mod directory", "modLocation", modLocation)

	modFlowpipeDir := path.Join(modLocation, app_specific.WorkspaceDataDir)
	err := ensureDir(modFlowpipeDir)
	if err != nil {
		return err
	}

	eventStoreDir := filepaths.EventStoreDir()
	err = ensureDir(eventStoreDir)
	if err != nil {
		return err
	}

	internalDir := filepaths.ModInternalDir()
	err = ensureDir(internalDir)
	if err != nil {
		return err
	}

	saltFileFullPath := filepath.Join(internalDir, "salt")
	salt, err := flowpipeSalt(saltFileFullPath, 32)
	if err != nil {
		return err
	}

	cache.GetCache().SetWithTTL("salt", salt, 24*7*52*99*time.Hour)

	return nil
}

// Assumes that the dir exists
//
// The function creates the salt if it does not exist, or it returns the existing
// salt if it's already there
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

// load and cache triggers and pipelines
// if we are in server mode and there is a modfile, setup the file watcher
func (m *Manager) initializeResources() error {
	modLocation := viper.GetString(constants.ArgModLocation)
	slog.Info("Starting Flowpipe", "modLocation", modLocation)

	var pipelines map[string]*modconfig.Pipeline
	var triggers map[string]*modconfig.Trigger
	var modInfo *modconfig.Mod

	if load_mod.ModFileExists(modLocation, app_specific.ModFileName) {

		workspacePath := viper.GetString(constants.ArgModLocation)
		flowpipeConfig, ew := steampipeconfig.LoadFlowpipeConfig(workspacePath)
		if ew != nil {
			ew.ShowWarnings()
			// check for error
			error_helpers.FailOnError(ew.Error)
		}

		// Add the "Credentials" in the context

		// TODO: this isn't the way .. we shouldn't be passing the credentials in the context
		var credentials map[string]modconfig.Credential

		if flowpipeConfig == nil {
			credentials = make(map[string]modconfig.Credential)
		} else {
			credentials = flowpipeConfig.Credentials
		}

		// effectively forever .. we don't want to expire the config
		if flowpipeConfig != nil {
			cache.GetCache().SetWithTTL("#flowpipeconfig", flowpipeConfig, 24*7*52*99*time.Hour)
		}

		w, errorAndWarning := workspace.LoadWorkspacePromptingForVariables(m.ctx, modLocation, credentials, app_specific.ModDataExtension)
		if errorAndWarning.Error != nil {
			return errorAndWarning.Error
		}

		// if we are running in server mode, setup the file watcher
		if m.shouldStartAPI() {
			if err := m.setupWatcher(w); err != nil {
				return err
			}
		}

		mod := w.Mod
		modInfo = mod

		pipelines = workspace.GetWorkspaceResourcesOfType[*modconfig.Pipeline](w)
		triggers = workspace.GetWorkspaceResourcesOfType[*modconfig.Trigger](w)

	} else {
		// there is no mod, just load pipelines and triggers from the directory
		var err error
		pipelines, triggers, err = load_mod.LoadPipelines(m.ctx, modLocation)
		if err != nil {
			return err
		}
	}

	m.triggers = triggers

	var rootModName string
	if modInfo != nil {
		rootModName = modInfo.ShortName
		if modInfo.Require != nil && modInfo.Require.FlowpipeVersionConstraint() != nil {
			flowpipeCliVersion := viper.GetString("main.version")
			flowpipeSemverVersion := semver.MustParse(flowpipeCliVersion)
			if !modInfo.Require.FlowpipeVersionConstraint().Check(flowpipeSemverVersion) {
				return perr.BadRequestWithMessage(fmt.Sprintf("flowpipe version %s does not satisfy %s which requires version %s", flowpipeCliVersion, rootModName, modInfo.Require.Flowpipe.MinVersionString))
			}
		}
	} else {
		rootModName = "local"
	}

	cache.GetCache().SetWithTTL("#rootmod.name", rootModName, 24*7*52*99*time.Hour)
	err := m.cachePipelinesAndTriggers(pipelines, triggers)
	if err != nil {
		return err
	}

	slog.Info("Pipelines and triggers loaded", "pipelines", len(pipelines), "triggers", len(triggers), "rootMod", rootModName)

	m.RootMod = modInfo

	return nil
}

func (m *Manager) setupWatcher(w *workspace.Workspace) error {
	err := w.SetupWatcher(m.ctx, func(c context.Context, e error) {
		slog.Error("error watching workspace", "error", e)
		m.apiService.ModMetadata.IsStale = true
	})
	if err != nil {
		return err
	}

	w.SetOnFileWatcherEventMessages(func() {
		slog.Info("caching pipelines and triggers")
		m.triggers = w.Mod.ResourceMaps.Triggers
		err = m.cachePipelinesAndTriggers(w.Mod.ResourceMaps.Pipelines, w.Mod.ResourceMaps.Triggers)
		if err != nil {
			slog.Error("error caching pipelines and triggers", "error", err)
		} else {
			slog.Info("cached pipelines and triggers")
			m.apiService.ModMetadata.IsStale = false
			m.apiService.ModMetadata.LastLoaded = time.Now()
		}

		// Reload scheduled triggers
		slog.Info("rescheduling triggers")
		if m.schedulerService != nil {
			m.schedulerService.Triggers = w.Mod.ResourceMaps.Triggers
			err := m.schedulerService.RescheduleTriggers()
			if err != nil {
				slog.Error("error rescheduling triggers", "error", err)
			} else {
				slog.Info("rescheduled triggers")
			}
		}
	})
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
	if !viper.GetBool(constants.ArgNoScheduler) {
		if err := s.Start(); err != nil {
			return err
		}
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
		//_ = slog.Sync()
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

	return nil
}

func (m *Manager) InterruptHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		slog.Debug("Manager exiting", "signal", sig)
		err := m.Stop()
		if err != nil {
			panic(err)
		}

		done <- true
	}()
	<-done
	slog.Debug("Manager exited")
}

func (m *Manager) cachePipelinesAndTriggers(pipelines map[string]*modconfig.Pipeline, triggers map[string]*modconfig.Trigger) error {
	inMemoryCache := cache.GetCache()
	var pipelineNames []string

	for _, p := range pipelines {
		pipelineNames = append(pipelineNames, p.Name())

		// TODO: how do we want to do this?
		inMemoryCache.SetWithTTL(p.Name(), p, 24*7*52*99*time.Hour)
	}

	inMemoryCache.SetWithTTL("#pipeline.names", pipelineNames, 24*7*52*99*time.Hour)

	var triggerNames []string
	for _, trigger := range triggers {
		triggerNames = append(triggerNames, trigger.Name())

		// if it's a webhook trigger, calculate the URL
		_, ok := trigger.Config.(*modconfig.TriggerHttp)
		if ok && !strings.HasPrefix(os.Getenv("RUN_MODE"), "TEST") {
			triggerUrl, err := calculateTriggerUrl(trigger)
			if err != nil {
				return err
			}
			trigger.Config.(*modconfig.TriggerHttp).Url = triggerUrl
		}

		inMemoryCache.SetWithTTL(trigger.Name(), trigger, 24*7*52*99*time.Hour)
	}
	inMemoryCache.SetWithTTL("#trigger.names", triggerNames, 24*7*52*99*time.Hour)

	return nil
}

func calculateTriggerUrl(trigger *modconfig.Trigger) (string, error) {
	salt, ok := cache.GetCache().Get("salt")
	if !ok {
		return "", perr.InternalWithMessage("salt not found")
	}

	hashString := util.CalculateHash(trigger.FullName, salt.(string))

	return "/hook/" + trigger.FullName + "/" + hashString, nil
}
