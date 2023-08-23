package pipeparser

import (
	"context"
	"fmt"
	"path"

	pcconstants "github.com/turbot/flowpipe/pipeparser/constants"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/parse"
	"github.com/turbot/flowpipe/pipeparser/pcerr"
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

func LoadFlowpipeConfig(ctx context.Context, configPath string) (*parse.ModParseContext, error) {
	parseCtx := parse.NewModParseContext(ctx, nil, configPath, parse.CreateDefaultMod,
		&filehelpers.ListOptions{
			// listFlag specifies whether to load files recursively
			Flags: filehelpers.Files | filehelpers.Recursive,
			// Exclude: w.exclusions,
			Include: filehelpers.InclusionsFromExtensions([]string{pcconstants.ModDataExtension, pcconstants.PipelineExtension}),
		})

	// check whether sourcePath is a glob with a root location which exists in the file system
	localSourcePath, globPattern, err := filehelpers.GlobRoot(configPath)
	if err != nil {
		return nil, err
	}

	if localSourcePath == globPattern {
		// if the path is a folder,
		// append '*' to the glob explicitly, to match all files in that folder.
		globPattern = path.Join(globPattern, fmt.Sprintf("*%s", pcconstants.PipelineExtension))
	}

	flowpipeConfigFilePaths, err := filehelpers.ListFiles(localSourcePath, &filehelpers.ListOptions{
		Flags:   filehelpers.AllRecursive,
		Include: []string{globPattern},
	})
	if err != nil {
		return nil, err
	}

	// pipelineFilePaths is the list of all pipeline files found in the pipelinePath
	if len(flowpipeConfigFilePaths) == 0 {
		return parseCtx, nil
	}

	fileData, diags := parse.LoadFileData(flowpipeConfigFilePaths...)
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("Failed to load workspace profiles", diags)
	}

	if len(fileData) != len(flowpipeConfigFilePaths) {
		return nil, pcerr.InternalWithMessage("Failed to load all pipeline files")

	}

	// Each file in the pipelineFilePaths is parsed and the result is stored in the bodies variable
	// bodies.data length should be the same with pipelineFilePaths length
	bodies, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("Failed to load workspace profiles", diags)
	}

	// do a partial decode
	content, diags := bodies.Content(modconfig.FlowpipeConfigBlockSchema)
	if diags.HasErrors() {
		return nil, error_helpers.HclDiagsToError("Failed to load workspace profiles", diags)
	}

	parseCtx.SetDecodeContent(content, fileData)

	// build parse context
	err = parse.ParseAllFlowipeConfig(parseCtx)
	if err != nil {
		return parseCtx, err
	}

	return parseCtx, nil
}

func LoadPipelines(ctx context.Context, configPath string) (map[string]*modconfig.Pipeline, error) {
	fpParseContext, err := LoadFlowpipeConfig(ctx, configPath)
	if err != nil {
		return nil, err
	}
	return fpParseContext.PipelineHcls, nil
}
