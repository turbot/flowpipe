package modinstaller

import (
	"errors"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GitUrlMode string

// String is used both by fmt.Print and by Cobra in help text
func (e *GitUrlMode) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *GitUrlMode) Set(v string) error {
	switch v {
	case "https", "ssh":
		*e = GitUrlMode(v)
		return nil
	default:
		return errors.New(`must be one of "https" or "ssh"`)
	}
}

// Type is only used in help text
func (e *GitUrlMode) Type() string {
	return "GitUrlMode"
}

const (
	GitUrlModeHTTPS GitUrlMode = "https"
	GitUrlModeSSH   GitUrlMode = "ssh"
)

func getGitUrl(modName string, urlMode GitUrlMode) string {
	return transformToGitURL(modName, urlMode)
}

func transformToGitURL(input string, urlMode GitUrlMode) string {

	if urlMode == GitUrlModeHTTPS {
		if !strings.HasPrefix(input, "https://") {
			input = "https://" + input
		}
		return input
	}

	if !strings.HasPrefix(input, "github.com") {
		return input
	}

	if !strings.HasPrefix(input, "git@") {
		input = "git@" + input
	}

	if !strings.HasSuffix(input, ".git") {
		input += ".git"
	}

	// Add a colon after the "git@github.com" part, so it replaces the first / with :
	if !strings.Contains(input, ":") {
		index := strings.Index(input, "/")
		input = input[:index] + ":" + input[index+1:]
	}

	return input
}

func getTags(repo string) ([]string, error) {
	// Create the remote with repository URL
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})

	// load remote references
	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return nil, err
	}

	// filters the references list and only keeps tags
	var tags []string
	for _, ref := range refs {
		if ref.Name().IsTag() {
			tags = append(tags, ref.Name().Short())
		}
	}

	return tags, nil
}

func getTagVersionsFromGit(repo string, includePrerelease bool) (semver.Collection, error) {
	tags, err := getTags(repo)
	if err != nil {
		return nil, err
	}

	versions := make(semver.Collection, len(tags))
	// handle index manually as we may not add all tags - if we cannot parse them as a version
	idx := 0
	for _, raw := range tags {
		v, err := semver.NewVersion(raw)
		if err != nil {
			continue
		}

		if !includePrerelease && v.Metadata() != "" || v.Prerelease() != "" {
			continue
		}
		versions[idx] = v
		idx++
	}
	// shrink slice
	versions = versions[:idx]

	// sort the versions in REVERSE order
	sort.Sort(sort.Reverse(versions))
	return versions, nil
}
