package execution

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
)

// This function mutates evalContext
func AddEachForEach(stepForEach *modconfig.StepForEach, evalContext *hcl.EvalContext) *hcl.EvalContext {
	eachValue := map[string]cty.Value{}
	eachValue[schema.AttributeTypeValue] = stepForEach.Each.Value
	eachValue[schema.AttributeKey] = cty.StringVal(stepForEach.Key)
	evalContext.Variables[schema.AttributeEach] = cty.ObjectVal(eachValue)
	return evalContext
}

// This function mutates evalContext
func AddLoop(stepLoop *modconfig.StepLoop, evalContext *hcl.EvalContext) *hcl.EvalContext {
	var loopValue cty.Value

	// Always override the loop variable, this function may be called in a loop
	// processing more than one step
	if stepLoop == nil {
		loopValue = cty.ObjectVal(map[string]cty.Value{
			"index": cty.NumberIntVal(int64(0)),
		})
	} else {
		loopValue = cty.ObjectVal(map[string]cty.Value{
			"index": cty.NumberIntVal(int64(stepLoop.Index)),
		})
	}

	evalContext.Variables["loop"] = loopValue
	return evalContext
}

func AddStepPrimitiveOutputAsResults(stepName string, output *modconfig.Output, evalContext *hcl.EvalContext) (*hcl.EvalContext, error) {

	var err error
	stepPrimitiveOutputMap := map[string]cty.Value{}

	if output != nil {
		stepPrimitiveOutputMap, err = output.AsCtyMap()
		if err != nil {
			return evalContext, perr.InternalWithMessage("unable to convert step output to cty map: " + err.Error())
		}
	}

	evalContext.Variables["result"] = cty.ObjectVal(stepPrimitiveOutputMap)

	return evalContext, nil
}

// This function *mutates* the evalContext passed in
func AddStepCalculatedOutputAsResults(stepName string, stepOutput map[string]interface{}, evalContext *hcl.EvalContext) (*hcl.EvalContext, error) {

	var err error

	var stepNativeOutputMap map[string]cty.Value
	if !evalContext.Variables["result"].IsNull() {
		stepNativeOutputMap = evalContext.Variables["result"].AsValueMap()
	} else {
		stepNativeOutputMap = map[string]cty.Value{}
	}

	stepOutputCtyMap := map[string]cty.Value{}

	for k, v := range stepOutput {
		stepOutputCtyMap[k], err = hclhelpers.ConvertInterfaceToCtyValue(v)
		if err != nil {
			return evalContext, perr.InternalWithMessage("unable to convert step output to cty map: " + err.Error())
		}
	}

	if stepNativeOutputMap["output"].IsNull() {
		stepNativeOutputMap["output"] = cty.ObjectVal(stepOutputCtyMap)
	} else {
		nestedOutputValueMap := stepNativeOutputMap["output"].AsValueMap()
		for k, v := range stepOutputCtyMap {
			if nestedOutputValueMap[k].IsNull() {
				nestedOutputValueMap[k] = v
			} else {
				return evalContext, perr.InternalWithMessage("output block '" + k + "' already exists in step '" + stepName + "'")
			}
		}

		stepNativeOutputMap["output"] = cty.ObjectVal(nestedOutputValueMap)
	}

	evalContext.Variables["result"] = cty.ObjectVal(stepNativeOutputMap)

	return evalContext, nil
}
