package execution

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
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

func AddStepOutputAsResults(stepName string, output *modconfig.Output, stepOutput map[string]interface{}, evalContext *hcl.EvalContext) (*hcl.EvalContext, error) {
	var err error
	stepNativeOutputMap := map[string]cty.Value{}

	if output != nil {
		stepNativeOutputMap, err = output.AsCtyMap()
		if err != nil {
			return evalContext, perr.InternalWithMessage("unable to convert step output to cty map: " + err.Error())
		}
	}

	stepOutputCtyMap := map[string]cty.Value{}

	for k, v := range stepOutput {
		stepOutputCtyMap[k], err = hclhelpers.ConvertInterfaceToCtyValue(v)
		if err != nil {
			return evalContext, perr.InternalWithMessage("unable to convert step output to cty map: " + err.Error())
		}
	}
	stepNativeOutputMap["output"] = cty.ObjectVal(stepOutputCtyMap)

	evalContext.Variables["results"] = cty.ObjectVal(stepNativeOutputMap)

	return evalContext, nil
}
