//nolint:forbidigo //TODO: initial import
package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/turbot/pipe-fittings/perr"
)

var GlobalDockerClient *DockerClient

var dockerInitializeMutex sync.Mutex

func Initialize(ctx context.Context) error {
	slog.Info("Initializing Docker client, attempting to lock")
	dockerInitializeMutex.Lock()
	defer func() {
		slog.Info("Docker client initialization complete, releasing lock")
		dockerInitializeMutex.Unlock()
	}()

	if GlobalDockerClient != nil {
		slog.Info("Docker client already initialized")
		return nil
	}
	slog.Info("Lock acquired, initializing Docker client")

	dc, err := New(WithContext(ctx), WithPingTest())
	if err != nil {
		slog.Error("Failed to initialize Docker client", "error", err)
		return err
	}

	GlobalDockerClient = dc

	slog.Info("Docker client initialized")
	return nil
}

// Client represents a connection to Docker.
type DockerClient struct {
	CLI *client.Client

	// If true, intermediate images will be removed when cleaning up
	// images. This keeps the environment clean, but increases build
	// times when Flowpipe is first launched. Default is true.
	PruneImages bool

	ctx context.Context
}

// Option defines a function signature for configuring the Docker client.
type Option func(*DockerClient) error

// WithContext configures the Docker client with a specific context.
func WithContext(ctx context.Context) Option {
	return func(c *DockerClient) error {
		c.ctx = ctx
		return nil
	}
}

func WithPruneImages() Option {
	return func(c *DockerClient) error {
		c.PruneImages = true
		return nil
	}
}

// WithPingTest configures the Docker client to perform a ping test to ensure the Docker service is running and available.
func WithPingTest() Option {
	return func(c *DockerClient) error {
		pingCtx, cancel := context.WithTimeout(c.ctx, time.Second*5)
		defer cancel()
		_, err := c.CLI.Ping(pingCtx)
		if err != nil {
			return err
		}
		return nil
	}
}

// New creates a new Docker client with the provided options.
func New(options ...Option) (*DockerClient, error) {

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	dc := &DockerClient{
		CLI: cli,

		// By default, leave intermediate images around to speed up launch time.
		PruneImages: true,
	}

	for _, option := range options {
		if err := option(dc); err != nil {
			return nil, err
		}
	}

	if dc.ctx == nil {
		dc.ctx = context.Background()
	}

	return dc, nil
}

func (dc *DockerClient) ImageExists(imageName string) (bool, error) {
	// Inspect the image to check if it exists
	_, _, err := dc.CLI.ImageInspectWithRaw(dc.ctx, imageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, perr.InternalWithMessage(fmt.Sprintf("error checking for image %s: %v", imageName, err.Error()))
	}
	return true, nil
}

func (dc *DockerClient) ImagePull(imageName string) error {
	resp, err := dc.CLI.ImagePull(dc.ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer resp.Close()

	// TODO - what do we do with the output? Or are we just checking for errors?
	_, err = io.ReadAll(resp)
	if err != nil {
		return err
	}

	return nil
}

// CleanupArtifacts deletes all containers and images related to flowpipe.
func (dc *DockerClient) CleanupArtifacts() error {
	// Delete any containers & images related to flowpipe
	err := dc.deleteContainersWithLabelKey("io.flowpipe.type")
	if err != nil {
		return fmt.Errorf("failed to cleanup flowpipe containers: %v", err)
	}
	err = dc.deleteImagesWithLabelKey("io.flowpipe.type")
	if err != nil {
		return fmt.Errorf("failed to cleanup flowpipe images: %v", err)
	}
	return nil
}

// deleteContainersWithLabel deletes all containers with the specified label.
func (dc *DockerClient) deleteContainersWithLabelKey(labelKey string) error {
	containers, err := dc.CLI.ContainerList(dc.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %s", err)
	}

	for _, container := range containers {
		if container.Labels[labelKey] != "" {
			err = dc.CLI.ContainerRemove(dc.ctx, container.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				slog.Error("failed to remove container", "containerID", container.ID, "error", err)
			} else {
				slog.Info("container deleted", "containerID", container.ID)
			}
		}
	}

	return nil
}

// deleteImagesWithLabel deletes all images with the specified label.
func (dc *DockerClient) deleteImagesWithLabelKey(labelKey string) error {

	images, err := dc.CLI.ImageList(dc.ctx, types.ImageListOptions{})
	if err != nil {
		slog.Error("failed to list images", "error", err)
		return perr.InternalWithMessage("failed to list images: " + err.Error())
	}

	for _, image := range images {
		if image.Labels[labelKey] != "" {
			imgRemoveOpts := types.ImageRemoveOptions{
				Force: true,
				// Prevent dangling images from being left around, but this means we have
				// to rebuild parts of the basic image on each startup (e.g. pip
				// install, npm install).
				// TODO - find some way to support this, but also to keep it
				// fast(er) by default
				PruneChildren: true,
			}
			_, err = dc.CLI.ImageRemove(dc.ctx, image.ID, imgRemoveOpts)
			if err != nil {
				slog.Error("failed to remove image", "imageID", image.ID, "error", err)
			} else {
				slog.Info("image deleted", "imageID", image.ID)
			}
		}
	}

	return nil
}

type CleanupArtifactsOptions struct {
	SkipLatest bool
}

type CleanupArtifactsOption func(*CleanupArtifactsOptions)

func WithSkipLatest(skipLatest bool) CleanupArtifactsOption {
	return func(options *CleanupArtifactsOptions) {
		options.SkipLatest = skipLatest
	}
}

// CleanupArtifacts deletes all containers and images related to flowpipe.
func (dc *DockerClient) CleanupArtifactsForLabel(key string, value string, opts ...CleanupArtifactsOption) error {
	err := dc.deleteContainersWithLabel(key, value, opts...)
	if err != nil {
		return fmt.Errorf("failed to cleanup flowpipe containers: %v", err)
	}
	err = dc.deleteImagesWithLabel(key, value, opts...)
	if err != nil {
		return fmt.Errorf("failed to cleanup flowpipe images: %v", err)
	}
	return nil
}

// deleteContainersWithLabel deletes all containers with the specified label.
func (dc *DockerClient) deleteContainersWithLabel(key string, value string, opts ...CleanupArtifactsOption) error {

	// Options
	cleanupOptions := &CleanupArtifactsOptions{
		SkipLatest: false,
	}
	for _, opt := range opts {
		opt(cleanupOptions)
	}

	// Convenience
	cli := dc.CLI

	// Prepare filters to match containers by label key and value
	labelFilter := filters.NewArgs()
	labelFilter.Add("label", fmt.Sprintf("%s=%s", key, value))
	listOptions := types.ContainerListOptions{
		// Include both running and stopped containers
		All:     true,
		Filters: labelFilter,
	}

	containers, err := cli.ContainerList(dc.ctx, listOptions)
	if err != nil {
		return fmt.Errorf("failed to list containers: %s", err)
	}

	// Iterate through the containers and stop/remove them
	for _, c := range containers {
		fmt.Println("delete container?", c.Image)
		if cleanupOptions.SkipLatest && strings.HasSuffix(c.Image, ":latest") {
			continue
		}
		// Gracefully stop the container if it's running
		if c.State == "running" {
			err = cli.ContainerStop(dc.ctx, c.ID, container.StopOptions{})
			if err != nil {
				slog.Warn(fmt.Sprintf("failed to stop container %s: %s", c.ID, err))
			} else {
				slog.Info(fmt.Sprintf("container %s stopped\n", c.ID))
			}
		}
		// Remove the container
		err = cli.ContainerRemove(dc.ctx, c.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to remove container %s: %s\n", c.ID, err))
		} else {
			slog.Info(fmt.Sprintf("container %s deleted\n", c.ID))
		}
	}

	return nil
}

// deleteImagesWithLabel deletes all images with the specified label.
func (dc *DockerClient) deleteImagesWithLabel(key string, value string, opts ...CleanupArtifactsOption) error {

	// Options
	cleanupOptions := &CleanupArtifactsOptions{
		SkipLatest: false,
	}
	for _, opt := range opts {
		opt(cleanupOptions)
	}

	// Convenience
	cli := dc.CLI

	// Prepare filters to match containers by label key and value
	labelFilter := filters.NewArgs()
	labelFilter.Add("label", fmt.Sprintf("%s=%s", key, value))
	listOptions := types.ImageListOptions{
		// Do not include intermediate images in the results, since
		// they are removed through the PruneChildren option below.
		All:     false,
		Filters: labelFilter,
	}

	images, err := cli.ImageList(dc.ctx, listOptions)
	if err != nil {
		return fmt.Errorf("failed to list images: %s", err)
	}

	for _, image := range images {
		fmt.Println("delete image?", image.RepoTags)
		if cleanupOptions.SkipLatest {
			isLatest := false
			for _, tag := range image.RepoTags {
				if strings.HasSuffix(tag, ":latest") {
					isLatest = true
				}
			}
			if isLatest {
				continue
			}
		}
		imgRemoveOpts := types.ImageRemoveOptions{
			// Just in case, since we should only be deleting images that
			// are not in use.
			Force: true,
			// Prevent dangling images from being left around, but this means we have
			// to rebuild parts of the basic image on each startup (e.g. pip
			// install, npm install).
			// TODO - We may want to make this an option for those who want faster
			// performance on startup, but don't mind having dangling images.
			PruneChildren: dc.PruneImages,
		}
		_, err = dc.CLI.ImageRemove(dc.ctx, image.ID, imgRemoveOpts)
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to remove image %s: %s\n", image.ID, err))
		} else {
			slog.Info(fmt.Sprintf("image %s deleted\n", image.ID))
		}
	}

	return nil
}
