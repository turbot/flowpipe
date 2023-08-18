package parse

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/flowpipe/pipeparser/modconfig"
	"github.com/zclconf/go-cty/cty"
)

type FlowpipeConfigParseContext struct {
	ParseContext
	PipelineHcls map[string]*modconfig.Pipeline
	TriggerHcls  map[string]modconfig.ITrigger
}

func (c *FlowpipeConfigParseContext) BuildEvalContext() {
	vars := map[string]cty.Value{}
	pipelineVars := map[string]cty.Value{}

	for _, pipeline := range c.PipelineHcls {
		pipelineVars[pipeline.Name()] = pipeline.AsCtyValue()
	}

	vars["pipeline"] = cty.ObjectVal(pipelineVars)

	c.ParseContext.BuildEvalContext(vars)
}

// AddPipeline stores this resource as a variable to be added to the eval context. It alse
func (c *FlowpipeConfigParseContext) AddPipeline(pipelineHcl *modconfig.Pipeline) hcl.Diagnostics {
	c.PipelineHcls[pipelineHcl.Name()] = pipelineHcl

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, pipelineHcl.Name())

	c.BuildEvalContext()
	return nil
}

func (c *FlowpipeConfigParseContext) AddTrigger(trigger modconfig.ITrigger) hcl.Diagnostics {

	c.TriggerHcls[trigger.GetName()] = trigger

	// remove this resource from unparsed blocks
	delete(c.UnresolvedBlocks, trigger.GetName())

	c.BuildEvalContext()
	return nil
}

func NewFlowpipeConfigParseContext(ctx context.Context, rootEvalPath string) *FlowpipeConfigParseContext {
	parseContext := NewParseContext(ctx, rootEvalPath)
	// TODO uncomment once https://github.com/turbot/steampipe/issues/2640 is done

	c := &FlowpipeConfigParseContext{
		ParseContext: parseContext,
		PipelineHcls: make(map[string]*modconfig.Pipeline),
		TriggerHcls:  make(map[string]modconfig.ITrigger),
	}

	c.BuildEvalContext()

	return c
}
