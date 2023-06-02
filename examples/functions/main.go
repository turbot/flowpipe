package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
	"github.com/turbot/flowpipe-functions/docker"
	"gopkg.in/yaml.v2"
)

type FunctionConfig struct {
	Name    string            `mapstructure:"name"`
	Runtime string            `mapstructure:"runtime"`
	Handler string            `mapstructure:"handler"`
	Src     string            `mapstructure:"src"`
	Env     map[string]string `mapstructure:"env"`

	// Runtime information
	AbsolutePath string     `mapstructure:"-"`
	CreatedAt    *time.Time `mapstructure:"-"`
	UpdatedAt    *time.Time `mapstructure:"-"`
	ContainerIDs []string   `mapstructure:"-"`
}

func (fnConfig *FunctionConfig) GetImageName() string {
	tag := fnConfig.CreatedAt.Format("20060102150405.000")
	if fnConfig.UpdatedAt != nil {
		tag = fnConfig.UpdatedAt.Format("20060102150405.000")
	}
	tag = strings.ReplaceAll(tag, ".", "")
	return fmt.Sprintf("flowpipe/%s:%s", fnConfig.Name, tag)
}

func (fnConfig *FunctionConfig) SetUpdatedAt() {
	now := time.Now()
	fnConfig.UpdatedAt = &now
}

func (fnConfig *FunctionConfig) GetEnv() []string {
	env := []string{}
	for k, v := range fnConfig.Env {
		env = append(env, fmt.Sprintf("%s=%s", quoteEnvVar(k), quoteEnvVar(v)))
	}
	return env
}

type ContainerConfig struct {
	Name  string            `mapstructure:"name"`
	Image string            `mapstructure:"image"`
	Cmd   []string          `mapstructure:"cmd"`
	Env   map[string]string `mapstructure:"env"`
}

func (containerConfig *ContainerConfig) GetEnv() []string {
	env := []string{}
	for k, v := range containerConfig.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

type AppConfig struct {
	Functions  map[string]FunctionConfig  `mapstructure:"functions"`
	Containers map[string]ContainerConfig `mapstructure:"containers"`
}

var config AppConfig

func main() {

	ctx := context.Background()

	dc, err := docker.New(docker.WithContext(ctx), docker.WithPingTest())
	if err != nil {
		log.Fatalf("Failed to connect to Docker: %v", err)
	}

	// Create a channel to receive OS signals
	sigCh := make(chan os.Signal, 1)

	// Notify the signal channel on SIGINT (Ctrl+C) and SIGTERM
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to handle the signal
	go func() {
		// Wait for the signal
		<-sigCh

		// Cleanup docker artifacts
		err := dc.CleanupArtifacts()
		if err != nil {
			log.Fatalf("Failed to cleanup flowpipe docker artifacts: %v", err)
		}

		// Exit the program
		os.Exit(0)
	}()

	// Read the YAML file
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Parse the YAML file into the config struct
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	now := time.Now()

	for fnName, fnConfig := range config.Functions {

		fmt.Println(fnConfig)

		// Set the CreatedAt time
		fnConfig.CreatedAt = &now

		// Set the name in the config for convenience
		if fnConfig.Name == "" {
			fnConfig.Name = fnName
		}

		// Convert Src to AbsolutePath
		absPath, err := filepath.Abs(fnConfig.Src)
		if err != nil {
			fmt.Println("Failed to resolve absolute path:", err)
			return
		}

		// Get file/directory information
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			fmt.Println("Absolute path not found:", err)
			return
		}

		// Check if the path is a directory
		if !fileInfo.IsDir() {
			fmt.Println("Absolute path must be a directory:", err)
			return
		}

		// Path looks good, let's use it
		fnConfig.AbsolutePath = absPath

		// Save it back to the main config
		config.Functions[fnName] = fnConfig
	}

	w := watcher.New()
	defer w.Close()

	// We can receive multiple events per cycle since they might be for
	// different function locations
	// TODO - the negative effect of this is when there are multiple changes
	// inside a given function directory, we will rebuild the Docker image
	// multiple times. We should probably debounce the events so that we
	// only rebuild the Docker image once per cycle.
	w.SetMaxEvents(0)

	// Block the watcher until it is started
	//w.Wait()

	// Watch all the function directories
	for _, fnConfig := range config.Functions {
		if err := w.AddRecursive(fnConfig.AbsolutePath); err != nil {
			fmt.Println("Failed to add path to watcher:", err)
		}
	}

	// Build the initial Docker image for each function
	// TODO - do this in parallel?
	for _, fnConfig := range config.Functions {
		if err := buildDockerImage(fnConfig.GetImageName(), fnConfig); err != nil {
			log.Fatalf("Failed to build Docker image: %v", err)
		}
	}

	// Build the initial Docker image for each function
	// TODO - do this in parallel?
	for fnName, fnConfig := range config.Functions {
		// Run the Docker container initially
		containerID, err := runDockerContainer(fnConfig)
		if err != nil {
			log.Fatalf("Failed to run Docker container: %v", err)
		}
		fnConfig.ContainerIDs = append(fnConfig.ContainerIDs, containerID)

		// Save back to the main config
		config.Functions[fnName] = fnConfig
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Printf("Detected file change: %v\n", event)
				if event.IsDir() && event.Op == watcher.Write {
					// Directory write events happen when there is any change to the files
					// or directories inside a directory. They are just noise and can be
					// safely ignored - especially since the event for the actual change
					// will be raised anyway.
					continue
				}
				for fnName, fnConfig := range config.Functions {
					fmt.Printf("Did this fn change? %s\n", fnName)
					eventMatchesFn, err := isSubPath(fnConfig.AbsolutePath, event.Path)
					if err != nil {
						log.Printf("Failed to check if event path is a subpath of function path: %v", err)
						continue
					}
					if eventMatchesFn {
						fmt.Printf("Yes, fn changed: %s\n", fnName)
						fnConfig.SetUpdatedAt()
						if err := buildDockerImage(fnConfig.GetImageName(), fnConfig); err != nil {
							log.Printf("Failed to rebuild Docker image: %v", err)
						} else {
							fmt.Println("Docker image rebuilt successfully. Restarting containers...")
							newIDs := []string{}
							for _, containerID := range fnConfig.ContainerIDs {
								if newID, err := restartDockerContainer(fnConfig, containerID); err != nil {
									log.Printf("Failed to restart Docker container: %v", err)
								} else {
									fmt.Println("Docker container restarted successfully.")
									newIDs = append(newIDs, newID)
								}
							}
							fnConfig.ContainerIDs = newIDs
							config.Functions[fnName] = fnConfig
						}
					}
				}
			case err := <-w.Error:
				log.Printf("File watcher error: %v", err)
			case <-w.Closed:
				return
			}
		}
	}()

	/*
		// TEST - convenient way to trigger SIGINT for testing
		// Sleep for 5 seconds
		time.Sleep(5 * time.Second)
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			p.Signal(os.Interrupt)
		}
	*/

	go startWebServer()

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}

}

func buildDockerImage(imageName string, fnConfig FunctionConfig) error {

	ctx := context.Background()

	log.Printf("buildDockerImage: %s, %s\n", imageName, fnConfig.AbsolutePath)

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}

	// Tar up the function code for use in the build
	buildCtx, err := archive.TarWithOptions(fnConfig.AbsolutePath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	// Our Dockerfile is runtime specific and stored outside the user-defined function
	// code.
	dockerfilePath := fmt.Sprintf("./runtimes/%s/Dockerfile", fnConfig.Runtime)
	dockerfileCtx, err := os.Open(dockerfilePath)
	if err != nil {
		return errors.Errorf("unable to open Dockerfile: %v", err)
	}
	defer dockerfileCtx.Close()

	// Add our Dockerfile to the build context (tar stream) that contains the user-defined
	// function code. The dockerfile gets a unique name, e.g. .dockerfile.64cf467fe12e4c96de83
	buildCtx, relDockerfile, err := build.AddDockerfileToBuildContext(dockerfileCtx, buildCtx)
	if err != nil {
		return err
	}

	buildOptions := types.ImageBuildOptions{
		Tags: []string{imageName},
		//Dockerfile:     fmt.Sprintf("/Users/nathan/src/flowpipe/examples/functions/runtimes/%s/Dockerfile", fnConfig.Runtime),
		// The Dockerfile is relative to the build context. Basically, it's the
		// unique name for the file that we added to the build context above.
		Dockerfile:     relDockerfile,
		SuppressOutput: false,
		Remove:         true,
		// This will update the FROM image in the Dockerfile to the latest
		// version.
		// TODO - only do this occasionally, e.g. once a day, for faster
		// performance during development.
		PullParent: true,
		Labels: map[string]string{
			"io.flowpipe.image.type":           "function",
			"io.flowpipe.image.runtime":        fnConfig.Runtime,
			"org.opencontainers.image.created": time.Now().Format(time.RFC3339),
		},
	}

	fmt.Println(buildOptions.Dockerfile)

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOptions)
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

func runDockerContainer(fnConfig FunctionConfig) (string, error) {

	ctx := context.Background()

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", err
	}

	// Only allow the local machine to connect
	hostIP := "127.0.0.1"
	// But allow any port to be allocated
	hostPort := "0"

	containerConfig := &container.Config{
		Image: fnConfig.GetImageName(),
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
		Labels: map[string]string{
			// Is this standard for containers?
			"org.opencontainers.container.created": time.Now().Format(time.RFC3339),
		},
		Env: fnConfig.GetEnv(),
	}
	if fnConfig.Handler != "" {
		// Override the Cmd if they have specified a custom handler location
		containerConfig.Cmd = []string{fnConfig.Handler}
	}

	containerHostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{{HostIP: hostIP, HostPort: hostPort}},
		},
	}

	// Create a container using the specified image
	resp, err := cli.ContainerCreate(ctx, containerConfig, containerHostConfig, &network.NetworkingConfig{}, nil, "")
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

	// Register the function with our API Gateway
	hookToLambdaEndpoint[fnConfig.Name] = fmt.Sprintf("http://localhost:%s/2015-03-31/functions/function/invocations", port)

	fmt.Printf("Docker container started successfully. Lambda function exposed on port %s\n", port)
	return resp.ID, nil
}

func restartDockerContainer(fnConfig FunctionConfig, containerID string) (string, error) {

	ctx := context.Background()

	newContainerID := ""

	fmt.Printf("restartDockerContainer: %s, %s\n", fnConfig.GetImageName(), containerID)

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return newContainerID, err
	}

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
	newContainerID, err = runDockerContainer(fnConfig)
	if err != nil {
		fmt.Printf("Container run failed: %v\n", err)
		return newContainerID, err
	}

	return newContainerID, nil
}

func isSubPath(basePath, subPath string) (bool, error) {
	relPath, err := filepath.Rel(basePath, subPath)
	if err != nil {
		return false, err
	}
	fmt.Printf("relPath: %s + %s = %s\n", basePath, subPath, relPath)
	return len(relPath) > 0 && !strings.HasPrefix(relPath, ".."), nil
}

func quoteEnvVar(s string) string {
	return strconv.Quote(s)
}
