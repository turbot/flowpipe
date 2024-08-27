package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/radovskyb/watcher"
	"github.com/spf13/viper"
	"github.com/turbot/flowpipe/internal/docker"
	"github.com/turbot/flowpipe/internal/fqueue"
	"github.com/turbot/flowpipe/internal/runtime"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/perr"
	putils "github.com/turbot/pipe-fittings/utils"
)

type Function struct {

	// fnuration information
	Name                  string                 `json:"name"`
	Runtime               string                 `json:"runtime"`
	Handler               string                 `json:"handler"`
	Source                string                 `json:"source"`
	Env                   map[string]string      `json:"env"`
	Event                 map[string]interface{} `json:"event"`
	Timeout               *int64                 `json:"timeout"`
	StartTimeoutInSeconds int                    `json:"start_timeout_in_seconds"`

	fqueue *fqueue.FunctionQueue

	// PullParentImagePeriod defines how often the parent image should be pulled.
	// This is useful for keeping the parent image up to date. Default is every
	// 24hrs. Accepts any valid golang duration string.
	PullParentImagePeriod string `json:"pull_parent_image_period"`

	// Runtime information
	AbsolutePath            string             `json:"absolute_path"`
	CreatedAt               *time.Time         `json:"created_at,omitempty"`
	UpdatedAt               *time.Time         `json:"updated_at,omitempty"`
	ParentImageLastPulledAt *time.Time         `json:"-"`
	CurrentVersionName      string             `json:"current_version_name"`
	Versions                map[string]Version `json:"versions"`

	// run context, need context.Background()
	ctx context.Context `json:"-"`

	// Flowpipe run context (e.g. for logging)
	runCtx       context.Context      `json:"-"`
	watcher      *watcher.Watcher     `json:"-"`
	dockerClient *docker.DockerClient `json:"-"`
}

const (
	// DefaultPullParentImagePeriod defines the default period for pulling the
	// parent image.
	DefaultPullParentImagePeriod = "24h"
)

// Option defines a function signature for fnuring the Docker client.
type FunctionOption func(*Function) error

// WithContext fnures the Docker client with a specific context.
func WithContext(ctx context.Context) FunctionOption {
	return func(c *Function) error {
		c.ctx = ctx
		return nil
	}
}

func WithRunContext(runContext context.Context) FunctionOption {
	return func(c *Function) error {
		c.runCtx = runContext
		return nil
	}
}

// WithFunctionDockerClient fnures the Docker client.
func WithDockerClient(client *docker.DockerClient) FunctionOption {
	return func(c *Function) error {
		c.dockerClient = client
		return nil
	}
}

func WithName(name string) FunctionOption {
	return func(c *Function) error {
		c.Name = name
		return nil
	}
}

func WithRuntime(runtime string) FunctionOption {
	return func(c *Function) error {
		c.Runtime = runtime
		return nil
	}
}

func WithStartTimeoutInSeconds(timeoutInSeconds int) FunctionOption {
	return func(c *Function) error {
		c.StartTimeoutInSeconds = timeoutInSeconds
		return nil
	}
}

// New creates a new Function fn with the provided options.
func New(options ...FunctionOption) (*Function, error) {

	now := time.Now()

	fc := &Function{
		CreatedAt: &now,
		Versions:  map[string]Version{},
		// By default, pull the parent image once per day.
		PullParentImagePeriod: DefaultPullParentImagePeriod,
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

// GetHandler returns the handler for the function.
func (fn *Function) GetHandler() string {
	if fn.Handler != "" {
		return fn.Handler
	}
	return "index.handler"
}

// GetImageName returns the docker image name (e.g. flowpipe/my_func) for the function.
func (fn *Function) GetImageName() string {
	return fmt.Sprintf("flowpipe/%s", fn.Name)
}

// GetImageTag returns the docker tag name (e.g.
// flowpipe/my_func:20230704150029969) for the function.
func (fn *Function) GetImageTag() string {
	tagTimestampFormat := "20060102T150405.000"
	tag := fn.CreatedAt.Format(tagTimestampFormat)
	if fn.UpdatedAt != nil {
		tag = fn.UpdatedAt.Format(tagTimestampFormat)
	}
	tag = strings.ReplaceAll(tag, ".", "")
	return fmt.Sprintf("%s:%s", fn.GetImageName(), tag)
}

// GetImageLatestTag returns the latest stream docker tag name (e.g.
// flowpipe/my_func:latest) for the function.
func (fn *Function) GetImageLatestTag() string {
	return fmt.Sprintf("%s:%s", fn.GetImageName(), "latest")
}

func (fn *Function) GetEnv() []string {
	env := []string{}
	for k, v := range fn.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

// SetUpdatedAt sets the updated at time.
func (fn *Function) SetUpdatedAt() {
	now := time.Now()
	fn.UpdatedAt = &now
}

// SetParentImageLastPulledAt sets the ParentImageLastPulledAt to the current time.
func (fn *Function) SetParentImageLastPulledAt() {
	now := time.Now()
	fn.ParentImageLastPulledAt = &now
}

// Load loads the function fn.
func (fn *Function) Load() error {
	if err := fn.Validate(); err != nil {
		return err
	}
	if err := fn.Pull(); err != nil {
		return err
	}
	if err := fn.Watch(); err != nil {
		return err
	}
	if err := fn.Build(); err != nil {
		return err
	}
	return nil
}

// Unload unloads the function fn.
func (fn *Function) Unload() error {
	// Stop watching
	if fn.watcher != nil {
		fn.watcher.Close()
	}

	// Cleanup artifacts from Docker
	return fn.CleanupArtifacts()
}

// Pull the fnuration of this function from its source location, e.g. GitHub or S3.
func (fn *Function) Pull() error {
	return nil
}

// Validate validates the function fn.
func (fn *Function) Validate() error {

	if fn.Name == "" {
		return perr.BadRequestWithMessage("name required for function")
	}

	if fn.Runtime == "" {
		return perr.BadRequestWithMessage("runtime required for function: " + fn.Name)
	}
	validRuntime := false
	validRuntimes, err := fn.RuntimesAvailable()
	if err != nil {
		return err
	}
	for _, r := range validRuntimes {
		if fn.Runtime == r {
			validRuntime = true
			break
		}
	}
	if !validRuntime {
		return perr.BadRequestWithMessage(fmt.Sprintf("invalid runtime `%s` requested for function: %s", fn.Runtime, fn.Name))
	}

	// Validate the source
	if fn.Source == "" {
		return perr.BadRequestWithMessage("'source' required for function: " + fn.Name)
	}
	// Convert source to an absolute path
	workspacePath := viper.GetString(constants.ArgModLocation)

	path := filepath.Join(workspacePath, fn.Source)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return perr.BadRequestWithMessage("failed to get absolute path to 'source 'for function: " + fn.Name)
	}
	srcStat, err := os.Stat(absPath)
	if err != nil {
		return perr.BadRequestWithMessage("'source' not found for function: " + fn.Name)
	}
	if !srcStat.IsDir() {
		return perr.BadRequestWithMessage("'source' must be a directory for function: " + fn.Name)
	}
	fn.AbsolutePath = absPath

	// Validate the PullParentImagePeriod
	if _, err := time.ParseDuration(fn.PullParentImagePeriod); err != nil {
		slog.Info("invalid pull parent image period", "PullParentImagePeriod", fn.PullParentImagePeriod, "function", fn.Name, "error", err)
		fn.PullParentImagePeriod = DefaultPullParentImagePeriod
	}

	return nil
}

func (fn *Function) Watch() error {

	fn.watcher = watcher.New()

	// Only get one event per watching period. Avoids unneccessary handler runs.
	fn.watcher.SetMaxEvents(1)

	// Watch all the function directories
	if err := fn.watcher.AddRecursive(fn.AbsolutePath); err != nil {
		return perr.BadRequestWithMessage("failed to add watch for function: " + fn.Name)
	}

	// Watch for changes and react to them
	go func() {
		for {
			select {
			case event := <-fn.watcher.Event:
				go func() {
					slog.Info("function watch event", "event", event)
					if err := fn.Build(); err != nil {
						slog.Error(fmt.Sprintf("failed to build function %s, got error: %v", fn.Name, err), "error", err, "functionName", fn.Name)
					}
				}()
			case err := <-fn.watcher.Error:
				slog.Error("file watcher error", "error", err)
			case <-fn.watcher.Closed:
				return
			}
		}
	}()

	// Start the watcher in the background, it'll check for changes every 100ms.
	go func() {
		// TODO - what do we do if this returns an error?
		err := fn.watcher.Start(time.Millisecond * 100)
		if err != nil {
			slog.Error("failed to start file watcher", "function", fn.Name, "error", err)
		}
	}()

	return nil
}

func (fn *Function) Start(imageName string) (string, error) {

	// Only allow the local machine to connect
	hostIP := "127.0.0.1"
	// But allow any port to be allocated
	hostPort := "0"

	containerfn := container.Config{
		Image: imageName,
		Cmd:   []string{fn.GetHandler()},
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
		Labels: map[string]string{
			// TODO - Is this standard for containers?
			"org.opencontainers.container.created": time.Now().Format(putils.RFC3339WithMS),
		},
		Env: fn.GetEnv(),
	}

	containerHostfn := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{{HostIP: hostIP, HostPort: hostPort}},
		},
	}

	if fn.Timeout != nil {
		timeout := int(*fn.Timeout)
		containerfn.StopTimeout = &timeout
	}

	// Create a container using the specified image
	resp, err := fn.dockerClient.CLI.ContainerCreate(fn.ctx, &containerfn, containerHostfn, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return "", err
	}

	// Start the container
	if err := fn.dockerClient.CLI.ContainerStart(fn.ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	// Get the allocated port for the Lambda function
	info, err := fn.dockerClient.CLI.ContainerInspect(fn.ctx, resp.ID)
	if err != nil {
		return "", err
	}
	port := info.NetworkSettings.Ports["8080/tcp"][0].HostPort

	// TODO - gross way to set the version
	v := fn.Versions[imageName]
	v.Port = port
	fn.Versions[imageName] = v

	slog.Info("Docker container started successfully. Lambda function exposed on port", "port", port, "functionName", fn.Name, "imageName", imageName, "containerID", resp.ID, "containerName", resp.ID[:12])
	return resp.ID, nil
}

func (fn *Function) IsStarted(imageName string) bool {
	return fn.Versions[imageName].Port != ""
}

func (fn *Function) StartIfNotStarted(imageName string) (string, error) {
	if fn.IsStarted(imageName) {
		return fn.Versions[imageName].Port, nil
	}
	ret, err := fn.Start(imageName)
	if err != nil {
		return ret, err
	}

	// Wait for the Lambda function to be ready, StartTimeoutInSeconds should be set with a value when we instantiated this struct
	// or the timeout will be 0
	timeout := fn.StartTimeoutInSeconds

	// Lambda endpoint
	v := fn.Versions[fn.CurrentVersionName]

	started := false
	for i := 0; i < timeout; i++ {
		resp, err := http.Get(v.LambdaEndpoint())
		if err != nil {
			slog.Debug("Lambda service may not be ready yet, retrying after 1 second", "error", err)

			// Wait before retrying
			time.Sleep(time.Second)
			continue
		}

		defer func() {
			if resp != nil && resp.Body != nil {
				err := resp.Body.Close()
				if err != nil {
					slog.Warn("Error closing response body", "error", err)
				}
			}
		}()

		if resp.StatusCode == http.StatusOK {
			started = true
			break
		}

		// Since the Lambda function is not ready yet, we wait for a second before retrying
		time.Sleep(time.Second)
	}

	if !started {
		slog.Error("Timeout waiting for Lambda function", "error", err)
		return ret, perr.TimeoutWithMessage("Timed out waiting for Lambda endpoint to be ready")
	}

	return ret, nil
}

func (fn *Function) Invoke(input []byte) (int, []byte, error) {
	output := []byte{}

	// Ensure the function has been started
	_, err := fn.StartIfNotStarted(fn.CurrentVersionName)
	if err != nil {
		return 0, output, err
	}

	// Forward request to lambda endpoint
	v := fn.Versions[fn.CurrentVersionName]

	slog.Debug("Executing Lambda function", "LambdaEndpoint", v.LambdaEndpoint(), "CurrentVersionName", fn.CurrentVersionName)

	// Invoke the Lambda function
	resp, err := http.Post(v.LambdaEndpoint(), "application/json", bytes.NewReader(input))
	if err != nil {
		slog.Error("Error invoking Lambda function", "error", err)
		return 0, output, err
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				slog.Warn("Error closing response body", "error", err)
			}
		}
	}()

	// Response handling
	output, err = io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading Lambda function response", "error", err)
	}

	return resp.StatusCode, output, err
}

func (fn *Function) Restart(containerId string) (string, error) {

	newContainerId := ""

	slog.Info("restartDockerContainer", "imageTag", fn.GetImageTag(), "containerId", containerId)

	// Stop the container
	err := fn.dockerClient.CLI.ContainerStop(fn.ctx, containerId, container.StopOptions{})
	if err != nil {
		slog.Error("Container stop failed", "error", err)
		return newContainerId, err
	}

	// Remove the container
	err = fn.dockerClient.CLI.ContainerRemove(fn.ctx, containerId, container.RemoveOptions{})
	if err != nil {
		slog.Error("Container remove failed", "error", err)
		return newContainerId, err
	}

	// Run the Docker container again
	newContainerId, err = fn.Start(fn.CurrentVersionName)
	if err != nil {
		slog.Error("Container run failed", "error", err)
		return newContainerId, err
	}

	return newContainerId, nil
}

func (fn *Function) Build() error {
	// if we want to wait for the result, we can do so like this
	receiveChannel := make(chan error)
	fn.fqueue.RegisterCallback(receiveChannel)

	fn.fqueue.Enqueue(fn.buildOne)

	// execute returns immediately
	fn.fqueue.Execute()

	err := <-receiveChannel
	return err
}

func (fn *Function) buildOne() error {

	// The UpdatedAt time is used as the build tag, ensuring unique
	// versions.
	fn.SetUpdatedAt()

	// Do the build!
	err := fn.buildImage()
	if err != nil {
		return err
	}

	// Add this version to the list for the function
	imageName := fn.GetImageTag()
	fn.Versions[imageName] = Version{}
	slog.Info("Function versions", "versions", fn.Versions)

	// The latest built version is the current version used for new invocations
	fn.CurrentVersionName = imageName

	return fn.CleanupOldArtifacts()
}

// RuntimesAvailable returns a list of available runtimes based on those defined
// in the runtimes directory.
func (fn *Function) RuntimesAvailable() ([]string, error) {
	dirNames, err := runtime.RuntimesAvailable()
	if err != nil {
		return nil, perr.InternalWithMessage("unable to read runtimes directory")
	}
	return dirNames, nil
}

func (fn *Function) PullParentImageDuration() time.Duration {
	// Cannot error since we validate during load
	d, _ := time.ParseDuration(fn.PullParentImagePeriod)
	return d
}

func (fn *Function) PullParentImageDueNow() bool {
	if fn.ParentImageLastPulledAt == nil {
		return true
	}
	return fn.ParentImageLastPulledAt.Add(fn.PullParentImageDuration()).Before(time.Now())
}

// buildImage builds the function image. Should only be called by Build().
func (fn *Function) buildImage() error {

	// Tar up the function code for use in the build
	buildCtx, err := archive.TarWithOptions(fn.AbsolutePath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	// Our Dockerfile is runtime specific and stored outside the user-defined function
	// code.

	dockerfileCtx, err := runtime.RuntimeDockerfile(fn.Runtime)
	if err != nil {
		return perr.InternalWithMessage("unable to open Dockerfile: " + err.Error())
	}
	defer dockerfileCtx.Close()

	// Add our Dockerfile to the build context (tar stream) that contains the user-defined
	// function code. The dockerfile gets a unique name, e.g. .dockerfile.64cf467fe12e4c96de83
	buildCtx, relDockerfile, err := build.AddDockerfileToBuildContext(dockerfileCtx, buildCtx)
	if err != nil {
		return err
	}

	buildOptions := types.ImageBuildOptions{
		// The image name is specific to every build, ensuring we're always running
		// an exact version.
		Tags: []string{fn.GetImageTag(), fn.GetImageLatestTag()},
		// The Dockerfile is relative to the build context. Basically, it's the
		// unique name for the file that we added to the build context above.
		Dockerfile: relDockerfile,
		// We want to see the output of the build process.
		SuppressOutput: false,
		// Remove the build container after the build is complete.
		Remove: true,
		// This will update the FROM image in the Dockerfile to the latest
		// version.
		// TODO - only do this occasionally, e.g. once a day, for faster
		// performance during development.
		PullParent: fn.PullParentImageDueNow(),
		// Add standard and identifying labels to the image.
		Labels: map[string]string{
			"io.flowpipe.type":                 "function",
			"io.flowpipe.runtime":              fn.Runtime,
			"io.flowpipe.name":                 fn.Name,
			"org.opencontainers.image.created": time.Now().Format(putils.RFC3339WithMS),
		},
	}

	slog.Info("Building image ...", "PullParent", buildOptions.PullParent, "Dockerfile", buildOptions.Dockerfile, "functionName", fn.Name)

	resp, err := fn.dockerClient.CLI.ImageBuild(fn.ctx, buildCtx, buildOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Build succeeded, so update the parent image pull time
	if buildOptions.PullParent {
		fn.SetParentImageLastPulledAt()
	}

	decoder := json.NewDecoder(resp.Body)

	var buildOutput []interface{}

	for {
		var message struct {
			Stream      string `json:"stream"`
			Error       string `json:"error"`
			ErrorDetail struct {
				Message string `json:"message"`
			} `json:"errorDetail"`
		}

		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				break
			}
			// Handle other errors (e.g., JSON decoding errors)
			slog.Error("Error decoding JSON from docker build response", "error", err)
			return perr.InternalWithMessage("Error decoding JSON from docker build response: " + err.Error())
		}

		buildOutput = append(buildOutput, message)
		if message.Error != "" {
			// Handle the build error
			slog.Error("Error building image", "error", message.Error, "buildOutput", buildOutput)
			return perr.InternalWithMessage("Error building image: " + message.Error)
		}

	}

	slog.Info("Docker image built successfully.", "functionName", fn.Name, "output", buildOutput)
	return nil
}

// Cleanup all docker containers and images for all versions of the given
// function.
func (fn *Function) CleanupArtifacts() error {
	return fn.dockerClient.CleanupArtifactsForLabel("io.flowpipe.name", fn.Name)
}

func (fn *Function) CleanupOldArtifacts() error {
	return fn.dockerClient.CleanupArtifactsForLabel("io.flowpipe.name", fn.Name, docker.WithSkipLatest(true))
}
