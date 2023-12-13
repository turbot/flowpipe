package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/pipe-fittings/perr"
)

type Container struct {

	// Configuration
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	Source          string            `json:"source"`
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

	// Runtime information
	CreatedAt   *time.Time               `json:"created_at,omitempty"`
	UpdatedAt   *time.Time               `json:"updated_at,omitempty"`
	ImageExists bool                     `json:"image_exists"`
	Runs        map[string]*ContainerRun `json:"runs"`

	// Internal
	ctx          context.Context
	runCtx       context.Context
	dockerClient *docker.DockerClient
}

type ContainerRun struct {
	ContainerID string       `json:"container_id"`
	Status      string       `json:"status"`
	Stdout      string       `json:"stdout"`
	Stderr      string       `json:"stderr"`
	Lines       []OutputLine `json:"lines"`
}

// Option defines a function signature for configuring the Docker client.
type ContainerOption func(*Container) error

// WithContext configures the Docker client with a specific context.
func WithContext(ctx context.Context) ContainerOption {
	return func(c *Container) error {
		c.ctx = ctx
		return nil
	}
}

func WithRunContext(ctx context.Context) ContainerOption {
	return func(c *Container) error {
		c.runCtx = ctx
		return nil
	}
}

// WithConfigDockerClient configures the Docker client.
func WithDockerClient(client *docker.DockerClient) ContainerOption {
	return func(c *Container) error {
		c.dockerClient = client
		return nil
	}
}

// NewConfig creates a new Function Config with the provided options.
func NewContainer(options ...ContainerOption) (*Container, error) {

	now := time.Now()

	fc := &Container{
		CreatedAt: &now,
		Cmd:       []string{},
		Env:       map[string]string{},
		Runs:      map[string]*ContainerRun{},
		// ImageExists: true,
		EntryPoint: []string{},
	}

	for _, option := range options {
		if err := option(fc); err != nil {
			return nil, err
		}
	}

	if fc.ctx == nil {
		fc.ctx = context.Background()
	}

	return fc, nil
}

func (c *Container) GetEnv() []string {
	var env []string
	for k, v := range c.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
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
	return nil
}

// Unload unloads the function config.
func (c *Container) Unload() error {
	// Cleanup artifacts from Docker
	return c.CleanupArtifacts()
}

// Validate validates the function config.
func (c *Container) Validate() error {

	if c.Name == "" {
		return perr.BadRequestWithMessage("name required for container")
	}

	if c.Image == "" {
		return perr.BadRequestWithMessage("image required for container: " + c.Name)
	}

	return nil
}

func (c *Container) SetRunStatus(containerID string, newStatus string) error {
	if c.Runs[containerID] == nil {
		c.Runs[containerID] = &ContainerRun{
			ContainerID: containerID,
		}
	}
	c.Runs[containerID].Status = newStatus
	return nil
}

func (c *Container) Run() (string, int, error) {
	containerID := ""

	start := time.Now()

	// Pull the Docker image if it's not already available
	if !c.ImageExists {
		imageExistsStart := time.Now()
		imageExists, err := c.dockerClient.ImageExists(c.Image)

		slog.Info("image exists check completed", "since", time.Since(imageExistsStart), "image", c.Image)

		if err != nil {
			slog.Error("image exists check error", "error", err)
			return containerID, -1, perr.InternalWithMessage("Error checking if image exists: " + err.Error())
		}

		if !imageExists {
			imagePullStart := time.Now()
			err = c.dockerClient.ImagePull(c.Image)
			slog.Info("image pull completed", "elapsed", time.Since(imagePullStart), "image", c.Image)

			if err != nil {
				return containerID, -1, perr.InternalWithMessage("Error pulling image: " + err.Error())
			}

			// Image has been pulled, so now exists locally
			c.ImageExists = true
		}
	}

	// Enforce a timeout to prevent runaway containers
	timeout := 60
	if c.Timeout != nil {
		timeout = int(*c.Timeout)
	}

	// Create a container using the specified image
	createConfig := container.Config{
		Image: c.Image,
		Cmd:   c.Cmd,
		Labels: map[string]string{
			// Set on the container since it's not on the image
			"io.flowpipe.type": "container",
			"io.flowpipe.name": c.Name,
			// Is this standard for containers?
			"org.opencontainers.container.created": time.Now().Format(time.RFC3339),
		},
		Env:         c.GetEnv(),
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
	if len(c.EntryPoint) != 0 {
		createConfig.Entrypoint = c.EntryPoint
	}

	if c.User != "" {
		createConfig.User = c.User
	}

	if c.Workdir != "" {
		createConfig.WorkingDir = c.Workdir
	}

	// Create the host configuration
	hostConfig := container.HostConfig{}

	if c.CpuShares != nil {
		hostConfig.Resources.CPUShares = *c.CpuShares
	}

	// Defaults to 128MB
	hostConfig.Resources.Memory = 128 * 1024 * 1024 // in bytes
	if c.Memory != nil {
		hostConfig.Resources.Memory = *c.Memory * 1024 * 1024 // in bytes
	}

	if c.MemoryReservation != nil {
		hostConfig.Resources.MemoryReservation = *c.MemoryReservation * 1024 * 1024
	}

	if c.MemorySwap != nil {
		hostConfig.Resources.MemorySwap = *c.MemorySwap * 1024 * 1024 // in bytes
	}

	if c.MemorySwappiness != nil {
		hostConfig.Resources.MemorySwappiness = c.MemorySwappiness
	}

	if c.ReadOnly != nil {
		hostConfig.ReadonlyRootfs = *c.ReadOnly
	}

	containerCreateStart := time.Now()
	containerResp, err := c.dockerClient.CLI.ContainerCreate(c.ctx, &createConfig, &hostConfig, &network.NetworkingConfig{}, nil, "")
	slog.Info("container create", "elapsed", time.Since(containerCreateStart), "image", c.Image, "container", containerResp.ID)
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
	slog.Info("container start", "elapsed", time.Since(containerStartStart), "image", c.Image, "container", containerResp.ID)
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
	slog.Info("container wait", "elapsed", time.Since(containerWaitStart), "image", c.Image, "container", containerResp.ID)

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

	c.Runs[containerID].Stdout = o.Stdout()
	c.Runs[containerID].Stderr = o.Stderr()
	c.Runs[containerID].Lines = o.Lines

	slog.Info("container logs", "elapsed", time.Since(containerLogsStart), "image", c.Image, "container", containerResp.ID, "combined", c.Runs[containerID].Lines)

	err = c.SetRunStatus(containerID, "logged")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to logged: " + err.Error())
	}

	// Remove the container

	if c.RetainArtifacts {
		slog.Info("retain artifacts", "name", c.Name)
	} else {
		containerRemoveStart := time.Now()
		err = c.dockerClient.CLI.ContainerRemove(c.ctx, containerID, types.ContainerRemoveOptions{})

		slog.Info("container remove", "elapsed", time.Since(containerRemoveStart), "image", c.Image, "container", containerResp.ID)
		if err != nil {
			// TODO - do we have to fail here? Perhaps things like not found can be ignored?
			return containerID, -1, perr.InternalWithMessage("Error removing container: " + err.Error())
		}
	}

	err = c.SetRunStatus(containerID, "removed")
	if err != nil {
		return containerID, -1, perr.InternalWithMessage("Error setting run status to removed: " + err.Error())
	}

	slog.Info("container run", "elapsed", time.Since(start), "image", c.Image, "container", containerResp.ID)

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

// Cleanup all docker containers for the given container
func (c *Container) CleanupArtifacts() error {
	slog.Info("cleanup artifacts", "name", c.Name)
	return c.dockerClient.CleanupArtifactsForLabel("io.flowpipe.name", c.Name)
}
