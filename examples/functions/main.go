package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/fsnotify/fsnotify"
)

func main() {
	// Specify the directory containing the Node.js files
	nodeJSFilesDir := "./lambda-nodejs"

	// Build the initial Docker image
	imageName := "flowpipe-lambda-nodejs"
	if err := buildDockerImage(imageName, nodeJSFilesDir); err != nil {
		log.Fatalf("Failed to build Docker image: %v", err)
	}

	// Create a file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	// Watch the directory for file changes
	go watchFiles(watcher, nodeJSFilesDir)

	// Run the Docker container initially
	containerID, err := runDockerContainer(imageName)
	if err != nil {
		log.Fatalf("Failed to run Docker container: %v", err)
	}

	// Monitor for file changes and rebuild/restart the container
	for {
		fmt.Printf("Wait for file change event...\n")
		select {
		case event := <-watcher.Events:
			fmt.Printf("Detected file change: %v\n", event)
			time.Sleep(100 * time.Millisecond)
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("Detected file change. Rebuilding Docker image...")

				if err := buildDockerImage(imageName, nodeJSFilesDir); err != nil {
					log.Printf("Failed to rebuild Docker image: %v", err)
				} else {
					fmt.Println("Docker image rebuilt successfully. Restarting container...")
					if containerID, err = restartDockerContainer(imageName, containerID); err != nil {
						log.Printf("Failed to restart Docker container: %v", err)
					} else {
						fmt.Println("Docker container restarted successfully.")
					}
				}
			}
		case err := <-watcher.Errors:
			log.Printf("File watcher error: %v", err)
		}
	}
}

func watchFiles(watcher *fsnotify.Watcher, nodeJSFilesDir string) {
	err := filepath.Walk(nodeJSFilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Printf("Failed to add file to watcher: %v", err)
			} else {
				log.Printf("Watching file: %s", path)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Failed to watch files: %v", err)
	}
}

func buildDockerImage(imageName, nodeJSFilesDir string) error {

	ctx := context.Background()

	log.Printf("buildDockerImage: %s, %s\n", imageName, nodeJSFilesDir)

	// Resolve the absolute path of the nodeJSFilesDir
	absPath, err := filepath.Abs(nodeJSFilesDir)
	if err != nil {
		return err
	}

	log.Printf("absPath: %s\n", absPath)

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	/*
		// Build the Docker image using the Dockerfile
		buildCtx, err := os.Open(absPath)
		if err != nil {
			return err
		}
		defer buildCtx.Close()
	*/

	tar, err := archive.TarWithOptions(absPath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer tar.Close()

	buildOptions := types.ImageBuildOptions{
		Tags:           []string{imageName},
		Dockerfile:     "Dockerfile",
		SuppressOutput: false,
		Remove:         true,
	}

	resp, err := cli.ImageBuild(ctx, tar, buildOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("Building Docker image...")
	// Output the build progress
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Docker image built successfully.")
	return nil
}

// ...

/*

func runDockerContainer(imageName string) error {
	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// Create a container using the specified image
	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: imageName,
	}, &container.HostConfig{}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return err
	}

	// Start the container
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	fmt.Println("Docker container started successfully.")
	return nil
}

*/

func runDockerContainer(imageName string) (string, error) {

	ctx := context.Background()

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	cli.NegotiateAPIVersion(ctx)

	// Create a container using the specified image
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
		},
	}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		return "", err
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	// Get the allocated port for the Lambda function
	info, err := cli.ContainerInspect(ctx, resp.ID)
	if err != nil {
		return "", err
	}
	port := info.NetworkSettings.Ports["8080/tcp"][0].HostPort

	fmt.Printf("Docker container started successfully. Lambda function exposed on port %s\n", port)
	return resp.ID, nil
}

func restartDockerContainer(imageName, containerID string) (string, error) {

	ctx := context.Background()

	newContainerID := ""

	fmt.Printf("restartDockerContainer: %s, %s\n", imageName, containerID)

	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return newContainerID, err
	}
	cli.NegotiateAPIVersion(ctx)

	// Stop the container
	err = cli.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		fmt.Printf("Container stop failed: %v\n", err)
		return newContainerID, err
	}

	// Remove the container
	err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		fmt.Printf("Container remove failed: %v\n", err)
		return newContainerID, err
	}

	// Run the Docker container again
	newContainerID, err = runDockerContainer(imageName)
	if err != nil {
		fmt.Printf("Container run failed: %v\n", err)
		return newContainerID, err
	}

	return newContainerID, nil
}

/*

func restartDockerContainer(imageName string) error {
	// Create Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// Stop the container
	err = cli.ContainerStop(context.Background(), "container-id", container.StopOptions{})
	if err != nil {
		return err
	}

	// Remove the container
	err = cli.ContainerRemove(context.Background(), "container-id", types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	// Run the Docker container again
	_, err = runDockerContainer(imageName)
	if err != nil {
		return err
	}

	return nil
}

*/
