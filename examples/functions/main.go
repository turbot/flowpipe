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
	"github.com/radovskyb/watcher"
)

func main() {
	// Specify the directory containing the Node.js files
	nodeJSFilesDir := "./lambda-nodejs"

	absPath, err := filepath.Abs(nodeJSFilesDir)
	if err != nil {
		log.Fatalf("Failed to resolve absolute path: %v", err)
	}

	// Build the initial Docker image
	imageName := "flowpipe-lambda-nodejs"
	if err := buildDockerImage(imageName, nodeJSFilesDir); err != nil {
		log.Fatalf("Failed to build Docker image: %v", err)
	}

	w := watcher.New()
	w.SetMaxEvents(1)
	defer w.Close()

	// Run the Docker container initially
	containerID, err := runDockerContainer(imageName)
	if err != nil {
		log.Fatalf("Failed to run Docker container: %v", err)
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Printf("Detected file change: %v\n", event)
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
			case err := <-w.Error:
				log.Printf("File watcher error: %v", err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(absPath); err != nil {
		log.Fatalln(err)
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
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
