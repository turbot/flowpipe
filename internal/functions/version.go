package function

import (
	"fmt"
	"strings"
	"time"
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
