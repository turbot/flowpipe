package parse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/turbot/flowpipe/internal/resources"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/app_specific"
	"github.com/turbot/pipe-fittings/load_mod"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/parse"
	"github.com/turbot/pipe-fittings/perr"
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
func LoadPipelines(ctx context.Context, configPath string) (map[string]*resources.Pipeline, map[string]*resources.Trigger, error) {

	mod, err := LoadPipelinesReturningItsMod(ctx, configPath)
	if err != nil {
		return nil, nil, err
	}

	var pipelines map[string]*resources.Pipeline
	var triggers map[string]*resources.Trigger

	if mod != nil && mod.GetModResources() != nil {
		modResources := resources.GetModResources(mod)
		pipelines = modResources.Pipelines
		triggers = modResources.Triggers
	}

	return pipelines, triggers, err
}

// TODO update this to NOT use deprecated LoadModWithFileName
func LoadPipelinesReturningItsMod(ctx context.Context, configPath string) (*modconfig.Mod, error) {
	var modDir string
	var fileName string
	var modFileNameToLoad string

	//Get information about the path
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	// Check if it's a regular file
	if info.Mode().IsRegular() {
		fileName = filepath.Base(configPath)
		modDir = filepath.Dir(configPath)

		// TODO: this is a hack (ish) to let the existing automated test to pass
		if filepath.Ext(fileName) == ".fp" {
			modFileNameToLoad = "ignore.fp"
		} else {
			modFileNameToLoad = fileName
		}
	} else if info.IsDir() { // Check if it's a directory

		defaultModSp := filepath.Join(configPath, app_specific.ModFileNameDeprecated)

		_, err := os.Stat(defaultModSp)
		if err == nil {
			// default mod.hcl exist
			fileName = app_specific.ModFileNameDeprecated
			modDir = configPath
		} else {
			fileName = "*.fp"
			modDir = configPath
		}
		modFileNameToLoad = fileName
	} else {
		return nil, perr.BadRequestWithMessage("invalid path")
	}

	parseCtx, err := parse.NewModParseContext(
		nil,
		modDir,
		parse.WithParseFlags(parse.CreateDefaultMod),
		parse.WithListOptions(filehelpers.ListOptions{
			Flags:   filehelpers.Files | filehelpers.Recursive,
			Include: []string{"**/" + fileName},
		}))

	if err != nil {
		return nil, err
	}
	mod, errorsAndWarnings := load_mod.LoadModWithFileName(ctx, modDir, modFileNameToLoad, parseCtx)

	if errorsAndWarnings.Error != nil {
		return nil, errorsAndWarnings.Error
	}

	return mod, nil
}
