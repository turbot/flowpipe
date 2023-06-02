package docker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Client represents a connection to Docker.
type DockerClient struct {
	cli *client.Client
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

// WithPingTest configures the Docker client to perform a ping test to ensure the Docker service is running and available.
func WithPingTest() Option {
	return func(c *DockerClient) error {
		pingCtx, cancel := context.WithTimeout(c.ctx, time.Second*5)
		defer cancel()
		_, err := c.cli.Ping(pingCtx)
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
		cli: cli,
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

// CleanupArtifacts deletes all containers and images related to flowpipe.
func (dc *DockerClient) CleanupArtifacts() error {
	// Delete any containers & images related to flowpipe
	err := dc.deleteContainersWithLabel("io.flowpipe.image.type")
	if err != nil {
		return fmt.Errorf("failed to cleanup flowpipe containers: %v", err)
	}
	err = dc.deleteImagesWithLabel("io.flowpipe.image.type")
	if err != nil {
		return fmt.Errorf("failed to cleanup flowpipe images: %v", err)
	}
	return nil
}

// deleteContainersWithLabel deletes all containers with the specified label.
func (dc *DockerClient) deleteContainersWithLabel(labelKey string) error {

	containers, err := dc.cli.ContainerList(dc.ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %s", err)
	}

	for _, container := range containers {
		if container.Labels[labelKey] != "" {
			err = dc.cli.ContainerRemove(dc.ctx, container.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				log.Printf("failed to remove container %s: %s\n", container.ID, err)
			} else {
				log.Printf("container %s deleted\n", container.ID)
			}
		}
	}

	return nil
}

// deleteImagesWithLabel deletes all images with the specified label.
func (dc *DockerClient) deleteImagesWithLabel(labelKey string) error {

	images, err := dc.cli.ImageList(dc.ctx, types.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %s", err)
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
				// PruneChildren: true,
			}
			_, err = dc.cli.ImageRemove(dc.ctx, image.ID, imgRemoveOpts)
			if err != nil {
				log.Printf("failed to remove image %s: %s\n", image.ID, err)
			} else {
				log.Printf("image %s deleted\n", image.ID)
			}
		}
	}

	return nil
}
