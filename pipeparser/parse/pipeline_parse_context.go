package parse

import (
	"context"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type FlowpipeConfigParseContext struct {
	ParseContext

	// TODO: temporary mapping until we sort out merging Flowpipe and Steampipe
	PipelineHcls map[string]*modconfig.Pipeline
	TriggerHcls  map[string]*modconfig.Trigger
}

// TODO: we need to push this up to the Mod to build the eval context, we should not have
// TODO: multiple places where we "build the eval context"
func (c *FlowpipeConfigParseContext) BuildEvalContext() {
	vars := map[string]cty.Value{}
	// pipelineVars := map[string]cty.Value{}

	// TODO: this logic can be improved if we know that there's only 1 mod (?)
	for _, pipeline := range c.PipelineHcls {
		// Split and get the last part for pipeline name
		parts := strings.Split(pipeline.Name(), ".")
		pipelineNameOnly := parts[len(parts)-1]
		modNameOnly := parts[0]

		modVars := vars[modNameOnly]
		if modVars == cty.NilVal {
			modVars = cty.ObjectVal(map[string]cty.Value{
				"pipeline": cty.ObjectVal(map[string]cty.Value{}),
			})
		}

		modVarsValueMap := modVars.AsValueMap()
		pipelineVars := modVarsValueMap["pipeline"]
		valueMaps := pipelineVars.AsValueMap()
		if valueMaps == nil {
			valueMaps = map[string]cty.Value{}
		}
		valueMaps[pipelineNameOnly] = pipeline.AsCtyValue()

		modVarsValueMap["pipeline"] = cty.ObjectVal(valueMaps)

		vars[modNameOnly] = cty.ObjectVal(modVarsValueMap)
	}

	c.ParseContext.BuildEvalContext(vars)
}

// AddPipeline stores this resource as a variable to be added to the eval context. It alse
func (c *FlowpipeConfigParseContext) AddPipeline(pipelineHcl *modconfig.Pipeline) hcl.Diagnostics {

	// Split and get the last part for pipeline name
	parts := strings.Split(pipelineHcl.Name(), ".")
	pipelineNameOnly := parts[len(parts)-1]

	c.PipelineHcls[pipelineNameOnly] = pipelineHcl

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, pipelineHcl.Name())

	c.BuildEvalContext()
	return nil
}

func (c *FlowpipeConfigParseContext) AddTrigger(trigger *modconfig.Trigger) hcl.Diagnostics {

	// Split and get the last part for pipeline name
	parts := strings.Split(trigger.Name(), ".")
	triggerNameOnly := parts[len(parts)-1]

	c.TriggerHcls[triggerNameOnly] = trigger

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, trigger.Name())

	c.BuildEvalContext()
	return nil
}

func NewFlowpipeConfigParseContext(ctx context.Context, rootEvalPath string) *FlowpipeConfigParseContext {
	parseContext := NewParseContext(ctx, rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done

	c := &FlowpipeConfigParseContext{
		ParseContext: parseContext,
		PipelineHcls: make(map[string]*modconfig.Pipeline),
		TriggerHcls:  make(map[string]*modconfig.Trigger),
	}

	c.BuildEvalContext()

	return c
}
