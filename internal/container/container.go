package container

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/pkg/archive"
	"github.com/radovskyb/watcher"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/constants"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fqueue"
	"github.com/turbot/pipe-fittings/perr"
)

type ContainerRunConfig struct {
	Cmd             []string          `json:"cmd"`
	Env             map[string]string `json:"env"`
	EntryPoint      []string          `json:"entrypoint"`
	RetainArtifacts bool              `json:"retain_artifacts"`
	Timeout         *int64            `json:"timeout"`
	CpuShares       *int64            `json:"cpu_shares"`
	User            string            `json:"user"`
	Workdir         string            `json:"workdir"`

	// Host configuration
	Memory            *int64 `json:"memory"`
	MemoryReservation *int64 `json:"memory_reservation"`
	MemorySwap        *int64 `json:"memory_swap"`
	MemorySwappiness  *int64 `json:"memory_swappiness"`
	ReadOnly          *bool  `json:"read_only"`
}

func (crc *ContainerRunConfig) GetEnv() []string {
	var env []string
	for k, v := range crc.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

type Container struct {

	// Configuration
	Name   string `json:"name"`
	Image  string `json:"image"`
	Source string `json:"source"`

	fqueue *fqueue.FunctionQueue

	// Runtime information
	CreatedAt   *time.Time               `json:"created_at,omitempty"`
	UpdatedAt   *time.Time               `json:"updated_at,omitempty"`
	ImageExists bool                     `json:"image_exists"`
	Runs        map[string]*ContainerRun `json:"runs"`

	// Internal
	ctx          context.Context
	runCtx       context.Context
	dockerClient *docker.DockerClient
	watcher      *watcher.Watcher
	runsMutex    sync.Mutex
}

type ContainerRun struct {
	ContainerID string       `json:"container_id"`
	Status      string       `json:"status"`
	Stdout      string       `json:"stdout"`
	Stderr      string       `json:"stderr"`
	Lines       []OutputLine `json:"lines"`
}

// ContainerOption defines a function signature for configuring the Container.
type ContainerOption func(*Container) error

// WithContext configures the Container with a specific context.
func WithContext(ctx context.Context) ContainerOption {
	return func(c *Container) error {
		c.ctx = ctx
		return nil
	}
}

// WithRunContext configures the Container with a specific run context.
func WithRunContext(ctx context.Context) ContainerOption {
	return func(c *Container) error {
		c.runCtx = ctx
		return nil
	}
}

// WithDockerClient configures the Docker client.
func WithDockerClient(client *docker.DockerClient) ContainerOption {
	return func(c *Container) error {
		c.dockerClient = client
		return nil
	}
}

func WithName(name string) ContainerOption {
	return func(c *Container) error {
		c.Name = name
		return nil
	}
}

// NewContainer creates a new Container with the provided ContainerOption.
func NewContainer(options ...ContainerOption) (*Container, error) {

	now := time.Now()

	fc := &Container{
		CreatedAt: &now,
		Runs:      map[string]*ContainerRun{},
		// ImageExists: true,
	}

	for _, option := range options {
		if err := option(fc); err != nil {
			return nil, err
		}
	}

	if fc.ctx == nil {
		fc.ctx = context.Background()
	}

	fc.fqueue = fqueue.NewFunctionQueue(fc.Name)

	return fc, nil
}

// SetUpdatedAt sets the updated at time.
func (c *Container) SetUpdatedAt() {
	now := time.Now()
	c.UpdatedAt = &now
}

// Load loads the function config.
func (c *Container) Load() error {
	if err := c.Validate(); err != nil {
		return err
	}

	if c.IsFromSource() {
		if err := c.Watch(); err != nil {
			return err
		}
		if err := c.Build(); err != nil {
			return err
		}
		c.ImageExists = true
	}

	return nil
}

// Unload unloads the function config.
func (c *Container) Unload() error {
	// Cleanup artifacts from Docker
	return c.CleanupArtifacts(false)
}

// Validate validates the function config.
func (c *Container) Validate() error {

	if c.Name == "" {
		return perr.BadRequestWithMessage("name required for container")
	}

	if c.Image == "" && c.Source == "" {
		return perr.BadRequestWithMessage("image or source required for container: " + c.Name)
	}

	return nil
}

func (c *Container) SetRunStatus(containerID string, newStatus string) error {
	c.runsMutex.Lock()
	defer c.runsMutex.Unlock()
	if c.Runs[containerID] == nil {
		c.Runs[containerID] = &ContainerRun{
			ContainerID: containerID,
		}
	}
	c.Runs[containerID].Status = newStatus
	return nil
}

func (c *Container) Run(cConfig ContainerRunConfig) (string, int, error) {
	containerID := ""

	start := time.Now()

	// Pull the Docker image if it's not already available
	if !c.ImageExists && !c.IsFromSource() {
		imageExistsStart := time.Now()
		imageExists, err := c.dockerClient.ImageExists(c.Image)

		slog.Debug("image exists check completed", "since", time.Since(imageExistsStart), "image", c.Image)

		if err != nil {
			slog.Error("image exists check error", "error", err)
			return containerID, -1, perr.InternalWithMessage("Error checking if image exists: " + err.Error())
		}

		if !imageExists {
			imagePullStart := time.Now()
			err = c.dockerClient.ImagePull(c.Image)
			slog.Debug("image pull completed", "elapsed", time.Since(imagePullStart), "image", c.Image)

			if err != nil {
				return containerID, -1, perr.InternalWithMessage("Error pulling image: " + err.Error())
			}

			// Image has been pulled, so now exists locally
			c.ImageExists = true
		}
	}

	// Enforce a timeout to prevent runaway containers
	timeout := 60
	if cConfig.Timeout != nil {
		timeout = int(*cConfig.Timeout)
	}

	// Create a container using the specified image
	imageName := c.Image
	if c.IsFromSource() {
		imageName = c.GetImageTag()
	}
	createConfig := container.Config{
		Image: imageName,
		Cmd:   cConfig.Cmd,
		Labels: map[string]string{
			// Set on the container since it's not on the image
			"io.flowpipe.type": "container",
			"io.flowpipe.name": c.Name,
			// Is this standard for containers?
			"org.opencontainers.container.created": time.Now().Format(time.RFC3339),
		},
		Env:         cConfig.GetEnv(),
		StopTimeout: &timeout,

		// I'm confused about how this works, and we're seeing a lot of control characters
		// with AWS CLI output.
		//
		// Tty: true, --no-cli-pager
		// Works! Format is good. But it's complicated and unexpected to need these.
		//
		// Tty: false, --no-cli-pager
		// Control chars in output.
		//
		// Tty: false
		// Control chars in output.
		//
		// Tty: true
		// Hangs, I presume because it's waiting for input on the paging.
		//
		// Overall, I trust this StackOverflow answer which says we want to use
		// docker run (without -it) to avoid control chars. But, we're still getting
		// the chars for some reason.
		// https://stackoverflow.com/questions/65824304/aws-cli-returns-json-with-control-codes-making-jq-fail
		//
		// The control characters seem to be at the start of each line:
		// [1 0 0 0 0 0 0 2 123 10 1 0 0 0 0 0 0 14 32 32 32 32 34 86 112 99 115 34 58 32 91 10
		// Specifically, see the "1 0 0 0 0 0 0 <number>" at the start of each line, which is:
		// 1 0 0 0 0 0 0  2 {
		// 1 0 0 0 0 0 0 14     "Vpcs": [
		//
		// OK - here is the answer - https://github.com/moby/moby/issues/7375#issuecomment-51462963
		// Docker adds a control character to each line of output to indicate if
		// it's stdout or stderr.
		Tty:          false, // Turn off interactive mode
		OpenStdin:    false, // Turn off stdin
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
	}

	// Only override Entrypoint if we pass content to c.EntryPoint
	if len(cConfig.EntryPoint) != 0 {
		createConfig.Entrypoint = cConfig.EntryPoint
	}

	if cConfig.User != "" {
		createConfig.User = cConfig.User
	}

	if cConfig.Workdir != "" {
		createConfig.WorkingDir = cConfig.Workdir
	}

	// Create the host configuration
	hostConfig := container.HostConfig{}

	if cConfig.CpuShares != nil {
		hostConfig.Resources.CPUShares = *cConfig.CpuShares
	}

	// Defaults to 128MB
	hostConfig.Resources.Memory = 128 * 1024 * 1024 // in bytes
	if cConfig.Memory != nil {
		hostConfig.Resources.Memory = *cConfig.Memory * 1024 * 1024 // in bytes
	}

	if cConfig.MemoryReservation != nil {
		hostConfig.Resources.MemoryReservation = *cConfig.MemoryReservation * 1024 * 1024
	}

	if cConfig.MemorySwap != nil {
		hostConfig.Resources.MemorySwap = *cConfig.MemorySwap * 1024 * 1024 // in bytes
	}

	if cConfig.MemorySwappiness != nil {
		hostConfig.Resources.MemorySwappiness = cConfig.MemorySwappiness
	}

	if cConfig.ReadOnly != nil {
		hostConfig.ReadonlyRootfs = *cConfig.ReadOnly
	}

	containerCreateStart := time.Now()
	containerResp, err := c.dockerClient.CLI.ContainerCreate(c.ctx, &createConfig, &hostConfig, &network.NetworkingConfig{}, nil, "")
	slog.Debug("container create", "elapsed", time.Since(containerCreateStart), "image", c.Image, "container", containerResp.ID)
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error creating container: " + err.Error())
	}
	containerID = containerResp.ID
	err = c.SetRunStatus(containerID, "created")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to created: " + err.Error())
	}

	// Start the container
	containerStartStart := time.Now()
	err = c.dockerClient.CLI.ContainerStart(c.ctx, containerID, types.ContainerStartOptions{})
	slog.Debug("container start", "elapsed", time.Since(containerStartStart), "image", c.Image, "container", containerResp.ID)
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error starting container: " + err.Error())
	}
	err = c.SetRunStatus(containerID, "started")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to started: " + err.Error())
	}

	// Wait for the container to finish
	var exitCode int64
	containerWaitStart := time.Now()
	statusCh, errCh := c.dockerClient.CLI.ContainerWait(c.ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return containerID, 1, perr.InternalWithMessage("Error waiting for container: " + err.Error())
		}
	case status := <-statusCh:
		// Set the status code of the container run
		exitCode = status.StatusCode
	}
	slog.Debug("container wait", "elapsed", time.Since(containerWaitStart), "image", c.Image, "container", containerResp.ID)

	err = c.SetRunStatus(containerID, "finished")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to finished: " + err.Error())
	}

	// Retrieve the container output
	containerLogsOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		// Timstamps inject timestamp text into the output, making it hard to parse
		Timestamps: false,
		// Get all logs from the container, not just the last X lines
		Tail: "all",
	}
	containerLogsStart := time.Now()
	reader, err := c.dockerClient.CLI.ContainerLogs(c.ctx, containerID, containerLogsOptions)
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error getting container logs: " + err.Error())
	}
	defer reader.Close()

	o := NewOutput()
	err = o.FromDockerLogsReader(reader)
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error reading container logs: " + err.Error())
	}

	c.runsMutex.Lock()
	c.Runs[containerID].Stdout = o.Stdout()
	c.Runs[containerID].Stderr = o.Stderr()
	c.Runs[containerID].Lines = o.Lines
	c.runsMutex.Unlock()

	slog.Info("container logs", "elapsed", time.Since(containerLogsStart), "image", c.Image, "container", containerResp.ID, "combined", c.Runs[containerID].Lines, "cmd", cConfig.Cmd)

	err = c.SetRunStatus(containerID, "logged")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to logged: " + err.Error())
	}

	// Remove the container

	if cConfig.RetainArtifacts {
		slog.Debug("retain artifacts", "name", c.Name)
	} else {
		containerRemoveStart := time.Now()
		err = c.dockerClient.CLI.ContainerRemove(c.ctx, containerID, types.ContainerRemoveOptions{})

		slog.Debug("container remove", "elapsed", time.Since(containerRemoveStart), "image", c.Image, "container", containerResp.ID)
		if err != nil {
			// TODO - do we have to fail here? Perhaps things like not found can be ignored?
			return containerID, -1, perr.InternalWithMessage("Error removing container: " + err.Error())
		}
	}

	err = c.SetRunStatus(containerID, "removed")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to removed: " + err.Error())
	}

	slog.Debug("container run", "elapsed", time.Since(start), "image", c.Image, "container", containerResp.ID)

	// If the container exited with a non-zero exit code, return an execution error
	if exitCode != 0 {

		// Get the Stderr and truncate it to 256 chars
		stdErr := o.Stderr()
		truncatedStdErr := truncateString(stdErr, 256)

		return containerID, int(exitCode), perr.ExecutionErrorWithMessage(truncatedStdErr)
	}

	return containerID, 0, nil
}

type StreamLines struct {
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

// truncateString truncates the string to the given length
func truncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength]
	}
	return s
}

// CleanupArtifacts will clean up all docker artifacts for the given container
func (c *Container) CleanupArtifacts(keepLatest bool) error {
	slog.Debug("cleanup artifacts", "name", c.Name)
	return c.dockerClient.CleanupArtifactsForLabel("io.flowpipe.name", c.Name, docker.WithSkipLatest(keepLatest))
}

func (c *Container) IsFromSource() bool {
	return c.Image == "" && c.Source != ""
}

func (c *Container) Watch() error {
	if c.Image != "" {
		return nil
	}

	wd := viper.GetString(constants.ArgModLocation)
	df := filepath.Join(wd, c.Source)

	c.watcher = watcher.New()
	c.watcher.SetMaxEvents(1)
	if err := c.watcher.Add(df); err != nil {
		return perr.BadRequestWithMessage("failed to add watch for container source: " + c.Source)
	}

	// watch for changes
	go func() {
		for {
			select {
			case event := <-c.watcher.Event:
				go func() {
					slog.Debug("container watch event", "event", event)
					if err := c.Build(); err != nil {
						slog.Error(fmt.Sprintf("failed to build container image %s, got error: %v", c.Name, err), "error", err, "containerName", c.Name)
					}
				}()
			case err := <-c.watcher.Error:
				slog.Error("file watcher error", "error", err)
			case <-c.watcher.Closed:
				return
			}
		}
	}()

	// watcher in background
	go func() {
		if err := c.watcher.Start(time.Millisecond * 1000); err != nil {
			slog.Error("failed to start file watcher", "container", c.Name, "error", err)
		}
	}()

	return nil
}

func (c *Container) Build() error {
	// if we want to wait for the result, we can do so like this
	receiveChannel := make(chan error)
	c.fqueue.RegisterCallback(receiveChannel)

	c.fqueue.Enqueue(c.buildOne)

	// execute returns immediately
	c.fqueue.Execute()

	err := <-receiveChannel
	return err
}

func (c *Container) buildOne() error {
	c.SetUpdatedAt()
	if err := c.buildImage(); err != nil {
		return err
	}

	return c.CleanupArtifacts(true)
}

// buildImage actually builds the container image. Should only be called by build.
func (c *Container) buildImage() error {
	wd := viper.GetString(constants.ArgModLocation)
	df := filepath.Join(wd, c.Source)
	dockerFilePath := strings.TrimSuffix(df, "/Dockerfile")

	buildCtx, err := archive.TarWithOptions(dockerFilePath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	buildOptions := types.ImageBuildOptions{
		Tags: []string{
			c.GetImageTag(),
			c.GetImageLatestTag(),
		},
		PullParent:     true,
		SuppressOutput: true,
		Remove:         true,
		Labels: map[string]string{
			"io.flowpipe.type":                 "container",
			"io.flowpipe.name":                 c.Name,
			"org.opencontainers.image.created": time.Now().Format(time.RFC3339),
		},
	}

	slog.Info("Building image ...", "container", c.Name)

	resp, err := c.dockerClient.CLI.ImageBuild(c.ctx, buildCtx, buildOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stderr, resp.Body)
	if err != nil {
		return err
	}

	slog.Info("Docker image built successfully.", "container", c.Name)

	return nil
}

// GetImageName returns the docker image name (e.g. flowpipe/my_func) for the function.
func (c *Container) GetImageName() string {
	return fmt.Sprintf("flowpipe/%s", c.Name)
}

// GetImageTag returns the docker image name and a timestamped tag
func (c *Container) GetImageTag() string {
	tagTimestampFormat := "20060102T150405.000"
	tag := c.CreatedAt.Format(tagTimestampFormat)
	if c.UpdatedAt != nil {
		tag = c.UpdatedAt.Format(tagTimestampFormat)
	}
	tag = strings.ReplaceAll(tag, ".", "")
	return fmt.Sprintf("%s:%s", c.GetImageName(), tag)
}

// GetImageLatestTag returns the docker image name with latest as tag
func (c *Container) GetImageLatestTag() string {
	return fmt.Sprintf("%s:%s", c.GetImageName(), "latest")
}
