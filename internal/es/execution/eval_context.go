package execution

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

func AddEachForEach(stepForEach *modconfig.StepForEach, evalContext *hcl.EvalContext) *hcl.EvalContext {
	eachValue := map[string]cty.Value{}
	eachValue[schema.AttributeTypeValue] = stepForEach.Each.Value
	eachValue[schema.AttributeKey] = cty.StringVal(stepForEach.Key)
	evalContext.Variables[schema.AttributeEach] = cty.ObjectVal(eachValue)

	return evalContext
}

func AddStepOutputAsResults(stepName string, output *modconfig.Output, stepOutput map[string]interface{}, evalContext *hcl.EvalContext) *hcl.EvalContext {
	return evalContext
}
