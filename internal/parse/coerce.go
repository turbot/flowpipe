package parse

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/connection"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func ValidateParams(p resources.ResourceWithParam, inputParams map[string]interface{}, evalCtx *hcl.EvalContext) []error {
	errors := []error{}

	// Lists out all the pipeline params that don't have a default value
	pipelineParamsWithNoDefaultValue := map[string]bool{}
	for _, v := range p.GetParams() {
		if v.Default.IsNull() && !v.Optional {
			pipelineParamsWithNoDefaultValue[v.Name] = true
		}
	}

	for k, v := range inputParams {
		param := p.GetParam(k)
		if param == nil {
			errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("unknown parameter specified '%s'", k)))
			continue
		}

		errorExist := false

		// if the param is a custom type, check the resource type
		if mapParam, ok := v.(map[string]any); ok {
			switch {
			case param.IsConnectionType():
				if !connection.ConnectionTypeMeetsRequiredType(param.TypeString, mapParam["resource_type"].(string), mapParam["type"].(string)) {
					errorExist = true
					errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted connection but received %s", k, param.TypeString)))
				} else {
					delete(pipelineParamsWithNoDefaultValue, k)
				}
				continue
			case param.IsNotifierType():
				if mapParam["resource_type"] != param.TypeString {
					errorExist = true
					errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received %s", k, param.TypeString, mapParam["resource_type"])))
				} else {
					delete(pipelineParamsWithNoDefaultValue, k)
				}
				continue
			}
		}

		if !hclhelpers.GoTypeMatchesCtyType(v, param.Type) {
			wanted := param.Type.FriendlyName()
			typeOfInterface := reflect.TypeOf(v)
			if typeOfInterface == nil {
				errorExist = true
				errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received null", k, wanted)))
			} else {
				received := typeOfInterface.String()
				errorExist = true
				errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("invalid data type for parameter '%s' wanted %s but received %s", k, wanted, received)))
			}
		} else {
			delete(pipelineParamsWithNoDefaultValue, k)
		}

		if !errorExist {
			errValidation := validateParam(param, v, evalCtx)
			if errValidation != nil {
				errors = append(errors, errValidation)
			}
		}

	}

	var missingParams []string
	for k := range pipelineParamsWithNoDefaultValue {
		missingParams = append(missingParams, k)
	}

	// Return error if there is no arguments provided for the pipeline params that don't have a default value
	if len(missingParams) > 0 {
		errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("missing parameter: %s", strings.Join(missingParams, ", "))))
	}

	return errors
}

func validateParam(param *resources.PipelineParam, inputParam interface{}, evalCtx *hcl.EvalContext) error {
	var valToValidate cty.Value
	var err error
	if !hclhelpers.IsComplexType(param.Type) && !param.Type.HasDynamicTypes() {
		valToValidate, err = gocty.ToCtyValue(inputParam, param.Type)
		if err != nil {
			return err
		}
	} else {
		// we'll do our best here
		valToValidate, err = hclhelpers.ConvertInterfaceToCtyValue(inputParam)
		if err != nil {
			return err
		}
	}
	validParam, diags, err := param.ValidateSetting(valToValidate, evalCtx)
	if err != nil {
		return err
	} else if !validParam {
		if len(diags) > 0 {
			return error_helpers.BetterHclDiagsToError(param.Name, diags)
		}
		return perr.BadRequestWithMessage("invalid value for param " + param.Name)
	}
	return nil
}

// This is inefficient because we are coercing the value from string -> Go using Cty (because that's how the pipeline is defined)
// and again we convert from Go -> Cty when we're executing the pipeline to build EvalContext when we're evaluating
// data are not resolved during parse time.
func CoerceParams(p resources.ResourceWithParam, inputParams map[string]string, evalCtx *hcl.EvalContext) (map[string]interface{}, []error) {
	errors := []error{}

	// Lists out all the pipeline params that don't have a default value
	pipelineParamsWithNoDefaultValue := map[string]bool{}
	for _, p := range p.GetParams() {
		if p.Default.IsNull() && !p.Optional {
			pipelineParamsWithNoDefaultValue[p.Name] = true
		}
	}

	res := map[string]interface{}{}

	for k, v := range inputParams {
		param := p.GetParam(k)
		if param == nil {
			errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("unknown parameter specified '%s'", k)))
			continue
		}

		var val interface{}
		if param.IsConnectionType() {

			if hclhelpers.IsComplexType(param.Type) {
				fakeFilename := fmt.Sprintf("<value for var.%s>", v)
				expr, diags := hclsyntax.ParseExpression([]byte(v), fakeFilename, hcl.Pos{Line: 1, Column: 1})
				if diags.HasErrors() {
					errors = append(errors, error_helpers.BetterHclDiagsToError(k, diags))
					continue
				}

				ctyVal, diags := expr.Value(evalCtx)
				if diags.HasErrors() {
					errors = append(errors, error_helpers.BetterHclDiagsToError(k, diags))
					continue
				}

				var err error
				val, err = hclhelpers.CtyToGo(ctyVal)
				if err != nil {
					errors = append(errors, err)
					continue
				}
			} else {
				dottedStringParts := strings.Split(v, ".")
				if len(dottedStringParts) != 3 {
					errors = append(errors, perr.BadRequestWithMessage("invalid connection string format"))
					continue
				}

				val = map[string]interface{}{
					"name":          dottedStringParts[2],
					"type":          dottedStringParts[1],
					"resource_type": schema.BlockTypeConnection,
					"temporary":     true,
				}
			}
		} else if param.IsNotifierType() {
			if hclhelpers.IsComplexType(param.Type) {
				fakeFilename := fmt.Sprintf("<value for var.%s>", v)
				expr, diags := hclsyntax.ParseExpression([]byte(v), fakeFilename, hcl.Pos{Line: 1, Column: 1})
				if diags.HasErrors() {
					errors = append(errors, error_helpers.BetterHclDiagsToError(k, diags))
					continue
				}

				ctyVal, diags := expr.Value(evalCtx)
				if diags.HasErrors() {
					errors = append(errors, error_helpers.BetterHclDiagsToError(k, diags))
					continue
				}

				var err error
				val, err = hclhelpers.CtyToGo(ctyVal)
				if err != nil {
					errors = append(errors, err)
					continue
				}
			} else {
				dottedStringParts := strings.Split(v, ".")
				if len(dottedStringParts) != 2 {
					errors = append(errors, perr.BadRequestWithMessage("invalid notifier string format"))
					continue
				}

				val = map[string]interface{}{
					"name":          dottedStringParts[1],
					"resource_type": schema.BlockTypeNotifier,
				}
			}
		} else {
			var moreErr error
			val, moreErr = hclhelpers.CoerceStringToGoBasedOnCtyType(v, param.Type)
			if moreErr != nil {
				errors = append(errors, moreErr)
				continue
			}
		}
		res[k] = val

		delete(pipelineParamsWithNoDefaultValue, k)

		if evalCtx != nil {
			errValidation := validateParam(param, val, evalCtx)
			if errValidation != nil {
				errors = append(errors, errValidation)
			}
		}
	}

	var missingParams []string
	for k := range pipelineParamsWithNoDefaultValue {
		missingParams = append(missingParams, k)
	}

	// Return error if there is no arguments provided for the pipeline params that don't have a default value
	if len(missingParams) > 0 {
		errors = append(errors, perr.BadRequestWithMessage(fmt.Sprintf("missing parameter: %s", strings.Join(missingParams, ", "))))
	}

	return res, errors
}
