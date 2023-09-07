package parse

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/flowpipe/pipeparser/error_helpers"
	"github.com/turbot/flowpipe/pipeparser/filepaths"
	"github.com/turbot/flowpipe/pipeparser/funcs"
	"github.com/turbot/flowpipe/pipeparser/hclhelpers"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/turbot/flowpipe/pipeparser/perr"
	"github.com/turbot/flowpipe/pipeparser/schema"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/zclconf/go-cty/cty"
)

func LoadModfile(modPath string) (*modconfig.Mod, error) {
	if !ModfileExists(modPath) {
		return nil, nil
	}

	// build an eval context just containing functions
	evalCtx := &hcl.EvalContext{
		Functions: funcs.ContextFunctions(modPath),
		Variables: make(map[string]cty.Value),
	}

	mod, res := ParseModDefinition(modPath, evalCtx)
	if res.Diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load mod", res.Diags)
	}

	return mod, nil
}

func ParseModDefinitionWithFileName(modPath string, modFileName string, evalCtx *hcl.EvalContext) (*modconfig.Mod, *DecodeResult) {
	res := newDecodeResult()

	// if there is no mod at this location, return error
	modFilePath := filepath.Join(modPath, modFileName)

	modFileFound := true
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		modFileFound = false
		for _, file := range filepaths.PipesComponentValidModFiles {
			modFilePath = filepath.Join(modPath, file)
			if _, err := os.Stat(modFilePath); os.IsNotExist(err) {

				continue
			} else {
				modFileFound = true
				break
			}
		}
	}

	if !modFileFound {
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "no mod file found in " + modPath,
		})
		return nil, res
	}

	fileData, diags := LoadFileData(modFilePath)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	body, diags := ParseHclFiles(fileData)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	workspaceContent, diags := body.Content(WorkspaceBlockSchema)
	res.addDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}

	block := hclhelpers.GetFirstBlockOfType(workspaceContent.Blocks, schema.BlockTypeMod)
	if block == nil {
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("no mod definition found in %s", modPath),
		})
		return nil, res
	}
	var defRange = block.DefRange
	if hclBody, ok := block.Body.(*hclsyntax.Body); ok {
		defRange = hclBody.SrcRange
	}
	mod := modconfig.NewMod(block.Labels[0], path.Dir(modFilePath), defRange)
	// set modFilePath
	mod.SetFilePath(modFilePath)

	mod, res = decodeMod(block, evalCtx, mod)
	if res.Diags.HasErrors() {
		return nil, res
	}

	// NOTE: IGNORE DEPENDENCY ERRORS

	// call decode callback
	diags = mod.OnDecoded(block, nil)
	res.addDiags(diags)

	return mod, res
}

// ParseModDefinition parses the modfile only
// it is expected the calling code will have verified the existence of the modfile by calling ModfileExists
// this is called before parsing the workspace to, for example, identify dependency mods
//
// This function only parse the "mod" block, and does not parse any resources in the mod file
func ParseModDefinition(modPath string, evalCtx *hcl.EvalContext) (*modconfig.Mod, *DecodeResult) {
	return ParseModDefinitionWithFileName(modPath, filepaths.PipesComponentModsFileName, evalCtx)
}

// ParseMod parses all source hcl files for the mod path and associated resources, and returns the mod object
// NOTE: the mod definition has already been parsed (or a default created) and is in opts.RunCtx.RootMod
func ParseMod(fileData map[string][]byte, pseudoResources []modconfig.MappableResource, parseCtx *ModParseContext) (*modconfig.Mod, *error_helpers.ErrorAndWarnings) {
	body, diags := ParseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to load all mod source files", diags))
	}

	content, moreDiags := body.Content(WorkspaceBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to load mod", diags))
	}

	mod := parseCtx.CurrentMod
	if mod == nil {
		return nil, error_helpers.NewErrorsAndWarning(fmt.Errorf("ParseMod called with no Current Mod set in ModParseContext"))
	}
	// get names of all resources defined in hcl which may also be created as pseudo resources
	hclResources, err := loadMappableResourceNames(content)
	if err != nil {
		return nil, error_helpers.NewErrorsAndWarning(err)
	}

	// if variables were passed in parsecontext, add to the mod
	if parseCtx.Variables != nil {
		for _, v := range parseCtx.Variables.RootVariables {
			if diags = mod.AddResource(v); diags.HasErrors() {
				return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to add resource to mod", diags))
			}
		}
	}

	// add pseudo resources to the mod
	addPseudoResourcesToMod(pseudoResources, hclResources, mod)

	// add the parsed content to the run context
	parseCtx.SetDecodeContent(content, fileData)

	// add the mod to the run context
	// - this it to ensure all pseudo resources get added and build the eval context with the variables we just added

	// ! This is the place where the child mods (dependent mods) resources are "pulled up" into this current evaluation
	// ! context.
	// !
	// ! Step through the code to find the place where the child mod resources are added to the "referencesValue"
	// !
	// ! Note that this resource MUST implement ModItem interface, otherwise it will look "flat", i.e. it will be added
	// ! to the current mod
	// !
	// ! There's also a bug where we test for ModTreeItem, we added a new interface ModItem for resources that are mod
	// ! resources but not necessarily need to be in the mod tree
	// !
	if diags = parseCtx.AddModResources(mod); diags.HasErrors() {
		return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to add mod to run context", diags))
	}

	// collect warnings as we parse
	var res = &error_helpers.ErrorAndWarnings{}

	// we may need to decode more than once as we gather dependencies as we go
	// continue decoding as long as the number of unresolved blocks decreases
	prevUnresolvedBlocks := 0
	for attempts := 0; ; attempts++ {
		diags = decode(parseCtx)
		if diags.HasErrors() {
			return nil, error_helpers.NewErrorsAndWarning(plugin.DiagsToError("Failed to decode all mod hcl files", diags))
		}
		// now retrieve the warning strings
		res.AddWarning(plugin.DiagsToWarnings(diags)...)

		// if there are no unresolved blocks, we are done
		unresolvedBlocks := len(parseCtx.UnresolvedBlocks)
		if unresolvedBlocks == 0 {
			log.Printf("[TRACE] parse complete after %d decode passes", attempts+1)
			break
		}
		// if the number of unresolved blocks has NOT reduced, fail
		if prevUnresolvedBlocks != 0 && unresolvedBlocks >= prevUnresolvedBlocks {
			str := parseCtx.FormatDependencies()
			msg := fmt.Sprintf("Failed to resolve dependencies after %d passes. Unresolved blocks:\n%s", attempts+1, str)
			return nil, error_helpers.NewErrorsAndWarning(perr.BadRequestWithTypeAndMessage(perr.ErrorCodeDependencyFailure, msg))
		}
		// update prevUnresolvedBlocks
		prevUnresolvedBlocks = unresolvedBlocks
	}

	// now tell mod to build tree of resources
	res.Error = mod.BuildResourceTree(parseCtx.GetTopLevelDependencyMods())

	return mod, res
}
