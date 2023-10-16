package function

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/runtime"
	"github.com/turbot/pipe-fittings/perr"
)

type Version struct {

	// Configuration
	Tag          string    `json:"tag"`
	Port         string    `json:"port"`
	ContainerIDs []string  `json:"container_ids"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`

	// Runtime information
	Function    *Function `json:"-"`
	BuildQueued bool      `json:"build_queued"`

	// Internal
	buildMutex *sync.Mutex
}

func (v *Version) LambdaEndpoint() string {
	return fmt.Sprintf("http://localhost:%s/2015-03-31/functions/function/invocations", v.Port)
}

// GetImageName returns the docker image name for the function.
func (v *Version) GetImageTag() string {
	tag := v.CreatedAt.Format("20060102150405.000")
	tag = strings.ReplaceAll(tag, ".", "")
	return fmt.Sprintf("flowpipe/%s:%s", v.Function.Name, tag)
}

func (v *Version) Build() error {

	// Ensure only one build is running at a time. I feel there is probably
	// a better way to do this with channels, but this works for now.
	if !v.buildMutex.TryLock() {
		// Already building
		v.BuildQueued = true
		return nil
	}

	// Do the build!
	err := v.buildImage()
	if err != nil {
		return err
	}

	// Need to clear the lock before re-running the Build function.
	v.buildMutex.Unlock()

	// If a build was queued, run it now.
	if v.BuildQueued {
		v.BuildQueued = false
		err := v.Build()
		if err != nil {
			return err
		}
	}

	return nil
}

// buildImage builds the function image. Should only be called by Build().
func (v *Version) buildImage() error {

	logger := fplog.Logger(v.Function.runCtx)

	// Tar up the function code for use in the build
	buildCtx, err := archive.TarWithOptions(v.Function.AbsolutePath, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer buildCtx.Close()

	// Our Dockerfile is runtime specific and stored outside the user-defined function
	// code.
	dockerfileCtx, err := runtime.RuntimeDockerfile(v.Function.Runtime)
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
		Tags: []string{v.GetImageTag()},
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
		PullParent: true,
		// Add standard and identifying labels to the image.
		Labels: map[string]string{
			"io.flowpipe.image.type":           "function",
			"io.flowpipe.image.runtime":        v.Function.Runtime,
			"org.opencontainers.image.created": time.Now().Format(time.RFC3339),
		},
	}

	resp, err := v.Function.dockerClient.CLI.ImageBuild(v.Function.ctx, buildCtx, buildOptions)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logger.Info("Building Docker image...")

	// Output the build progress
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		logger.Error("Error reading build output: "+err.Error(), "error", err)
		return err
	}

	logger.Info("Docker image built successfully.")

	return nil
}
