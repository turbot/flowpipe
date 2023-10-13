package runtime

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed resources/*
var resourcesFs embed.FS

func RuntimesAvailable() ([]string, error) {
	dirEntry, err := resourcesFs.ReadDir("resources")
	if err != nil {
		return nil, err
	}

	var runtimes []string
	for _, f := range dirEntry {
		runtimes = append(runtimes, strings.Replace(f.Name(), "_", ":", 1))
	}
	return runtimes, nil
}

func RuntimeDockerfile(runtime string) (fs.File, error) {
	return resourcesFs.Open("resources/" + strings.Replace(runtime, ":", "_", 1) + "/Dockerfile")
}
