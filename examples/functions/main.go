package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/turbot/flowpipe-functions/container"
	"github.com/turbot/flowpipe-functions/docker"
	"github.com/turbot/flowpipe-functions/function"
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

func quoteEnvVar(s string) string {
	return strconv.Quote(s)
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

type DockerClientContext struct{}

var config AppConfig

func main() {

	ctx := context.Background()

	dc, err := docker.New(docker.WithContext(ctx), docker.WithPingTest())
	if err != nil {
		log.Fatalf("Failed to connect to Docker: %v", err)
	}

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

	functions := map[string]*function.Function{}

	for fnName, fnConfig := range config.Functions {

		fmt.Println(fnConfig)

		fn, err := function.New(
			function.WithContext(ctx),
			function.WithDockerClient(dc),
		)
		if err != nil {
			panic(err)
		}
		fn.Name = fnName
		fn.Runtime = fnConfig.Runtime
		fn.Handler = fnConfig.Handler
		fn.Src = fnConfig.Src
		fmt.Println("Loading...")
		err = fn.Load()
		if err != nil {
			panic(err)
		}
		fmt.Println("Loaded")

		functions[fnName] = fn

	}

	containers := map[string]*container.Container{}

	for cName, cConfig := range config.Containers {

		fmt.Println(cConfig)

		c, err := container.NewContainer(
			container.WithContext(ctx),
			container.WithDockerClient(dc),
		)
		if err != nil {
			panic(err)
		}
		c.Name = cName
		c.Image = cConfig.Image
		c.Cmd = cConfig.Cmd
		c.Env = cConfig.Env
		fmt.Println("Loading...")
		err = c.Load()
		if err != nil {
			panic(err)
		}
		fmt.Println("Loaded")

		containers[cName] = c

	}

	// Create a channel to receive OS signals
	sigCh := make(chan os.Signal, 1)

	// Notify the signal channel on SIGINT (Ctrl+C) and SIGTERM
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle the signal
	go func() {

		// Wait for the signal
		<-sigCh

		// TODO - this should be dynamic, what if the functions change through
		// config changes?
		for _, fn := range functions {
			fmt.Printf("Stopping function: %s\n", fn.Name)
			err := fn.Unload()
			if err != nil {
				log.Fatalf("Failed to stop function: %v", err)
			}
		}

		// Cleanup docker artifacts
		// TODO - Can we remove this since we cleanup per function etc?
		err := dc.CleanupArtifacts()
		if err != nil {
			log.Fatalf("Failed to cleanup flowpipe docker artifacts: %v", err)
		}

		// Exit the program
		os.Exit(0)
	}()

	startWebServer(ctx, functions, containers)

}
