package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

var hookToLambdaEndpoint = map[string]string{}

//"hook_name": "http://lambda_endpoint",

func startWebServer() {
	router := gin.Default()

	router.POST("/api/v0/hooks/:hookName", handleHookRequest)
	router.POST("/api/v0/containers/:containerName", handleContainerRequest)

	err := router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

func handleHookRequest(c *gin.Context) {
	hookName := c.Param("hookName")
	lambdaEndpoint, ok := hookToLambdaEndpoint[hookName]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid hook name"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}

	// Forward request to lambda endpoint
	resp, err := http.Post(lambdaEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request to lambda"})
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read lambda response"})
		return
	}

	// Forward lambda response to client
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)
}

func handleContainerRequest(c *gin.Context) {
	containerName := c.Param("containerName")
	containerConfig, ok := config.Containers[containerName]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid container name"})
		return
	}

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Pull the Docker image if it's not already available
	imageExists, err := imageExists(cli, containerConfig.Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !imageExists {
		err = pullImage(cli, containerConfig.Image)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Create a container using the specified image
	createConfig := container.Config{
		Image: containerConfig.Image,
		Cmd:   containerConfig.Cmd,
		Labels: map[string]string{
			// Set on the container since it's not on the image
			"io.flowpipe.image.type": "container",
			// Is this standard for containers?
			"org.opencontainers.container.created": time.Now().Format(time.RFC3339),
		},
		Env: containerConfig.GetEnv(),
	}
	fmt.Println("createConfig", createConfig)
	containerResp, err := cli.ContainerCreate(context.Background(), &createConfig, &container.HostConfig{}, &network.NetworkingConfig{}, nil, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Start the container
	err = cli.ContainerStart(context.Background(), containerResp.ID, types.ContainerStartOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start container"})
		return
	}

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(context.Background(), containerResp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to wait for container"})
			return
		}
	case <-statusCh:
	}

	// Retrieve the container output
	outputBuf := new(bytes.Buffer)
	containerLogsOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       "all",
	}
	reader, err := cli.ContainerLogs(context.Background(), containerResp.ID, containerLogsOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve container logs"})
		return
	}
	defer reader.Close()

	_, err = outputBuf.ReadFrom(reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read container logs"})
		return
	}

	// Remove the container
	err = cli.ContainerRemove(context.Background(), containerResp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		log.Printf("Failed to remove container: %v", err)
	}

	// Return the container output as the result of the API call
	c.Data(http.StatusOK, "text/plain; charset=utf-8", outputBuf.Bytes())
}

func imageExists(cli *client.Client, imageName string) (bool, error) {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
			// Check for image ID match if tag is not present
			if strings.HasPrefix(tag, imageName+":") {
				return true, nil
			}
		}
	}

	return false, nil
}

func pullImage(cli *client.Client, imageName string) error {
	resp, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer resp.Close()
	_, err = io.ReadAll(resp)
	if err != nil {
		return err
	}

	return nil
}
