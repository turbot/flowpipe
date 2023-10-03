package pipeparser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/flowpipe/pipeparser/filepaths"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/perr"
	filehelpers "github.com/turbot/go-kit/files"
)

// ToError formats the supplied value as an error (or just returns it if already an error)
func ToError(val interface{}) error {
	if e, ok := val.(error); ok {
		return e
	} else {
		// return fperr.InternalWithMessage(fmt.Sprintf("%v", val))
		return fmt.Errorf("%v", val)
	}
}

// Convenient function to support testing
//
// # The automated tests were initially created before the concept of Mod is introduced in Flowpipe
//
// We can potentially remove this function, but we have to refactor all our test cases
func LoadPipelines(ctx context.Context, configPath string) (map[string]*modconfig.Pipeline, map[string]*modconfig.Trigger, error) {

	var modDir string
	var fileName string
	var modFileNameToLoad string

	// Get information about the path
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*modconfig.Pipeline{}, map[string]*modconfig.Trigger{}, nil
		}
		return nil, nil, err
	}

	// Check if it's a regular file
	if info.Mode().IsRegular() {
		fileName = filepath.Base(configPath)
		modDir = filepath.Dir(configPath)

		// TODO: this is a hack (ish) to let the existing automated test to pass
		if filepath.Ext(fileName) == ".fp" {
			modFileNameToLoad = "ignore.sp"
		} else {
			modFileNameToLoad = fileName
		}
	} else if info.IsDir() { // Check if it's a directory

		defaultModSp := filepath.Join(configPath, filepaths.PipesComponentModsFileName)

		_, err := os.Stat(defaultModSp)
		if err == nil {
			// default mod.sp exist
			fileName = filepaths.PipesComponentModsFileName
			modDir = configPath
		} else {
			fileName = "*.fp"
			modDir = configPath
		}
		modFileNameToLoad = fileName
	} else {
		return nil, nil, perr.BadRequestWithMessage("invalid path")
	}

	parseCtx := parse.NewModParseContext(
		ctx,
		nil,
		modDir,
		parse.CreateTransientLocalMod,
		&filehelpers.ListOptions{
			Flags:   filehelpers.Files | filehelpers.Recursive,
			Include: []string{"**/" + fileName},
		})

	mod, errorsAndWarnings := LoadModWithFileName(modDir, modFileNameToLoad, parseCtx)

	var pipelines map[string]*modconfig.Pipeline
	var triggers map[string]*modconfig.Trigger

	if mod != nil && mod.ResourceMaps != nil {
		pipelines = mod.ResourceMaps.Pipelines
		triggers = mod.ResourceMaps.Triggers
	}
	return pipelines, triggers, errorsAndWarnings.Error
}
