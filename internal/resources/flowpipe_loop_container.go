package resources

import (
	"reflect"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/iancoleman/strcase"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type LoopContainerStep struct {
	LoopStep

	Image             *string            `json:"image,omitempty" hcl:"image,optional" cty:"image"`
	Source            *string            `json:"source,omitempty" hcl:"source,optional" cty:"source"`
	Cmd               *[]string          `json:"cmd,omitempty" hcl:"cmd,optional" cty:"cmd"`
	Env               *map[string]string `json:"env,omitempty" hcl:"env,optional" cty:"env"`
	Entrypoint        *[]string          `json:"entrypoint,omitempty" hcl:"entrypoint,optional" cty:"entrypoint"`
	CpuShares         *int64             `json:"cpu_shares,omitempty" hcl:"cpu_shares,optional" cty:"cpu_shares"`
	Memory            *int64             `json:"memory,omitempty" hcl:"memory,optional" cty:"memory"`
	MemoryReservation *int64             `json:"memory_reservation,omitempty" hcl:"memory_reservation,optional" cty:"memory_reservation"`
	MemorySwap        *int64             `json:"memory_swap,omitempty" hcl:"memory_swap,optional" cty:"memory_swap"`
	MemorySwappiness  *int64             `json:"memory_swappiness,omitempty" hcl:"memory_swappiness,optional" cty:"memory_swappiness"`
	ReadOnly          *bool              `json:"read_only,omitempty" hcl:"read_only,optional" cty:"read_only"`
	User              *string            `json:"user,omitempty" hcl:"user,optional" cty:"user"`
	Workdir           *string            `json:"workdir,omitempty" hcl:"workdir,optional" cty:"workdir"`
}

func (l *LoopContainerStep) Equals(other LoopDefn) bool {
	if l == nil && helpers.IsNil(other) {
		return true
	}

	if l == nil && !helpers.IsNil(other) || l != nil && helpers.IsNil(other) {
		return false
	}

	otherLoopContainerStep, ok := other.(*LoopContainerStep)
	if !ok {
		return false
	}

	if !l.LoopStep.Equals(otherLoopContainerStep.LoopStep) {
		return false
	}

	// compare env using reflection
	if !reflect.DeepEqual(l.Env, otherLoopContainerStep.Env) {
		return false
	}

	if l.Cmd == nil && otherLoopContainerStep.Cmd != nil || l.Cmd != nil && otherLoopContainerStep.Cmd == nil {
		return false
	} else if l.Cmd != nil {
		if slices.Compare(*l.Cmd, *otherLoopContainerStep.Cmd) != 0 {
			return false
		}
	}

	if l.Entrypoint == nil && otherLoopContainerStep.Entrypoint != nil || l.Entrypoint != nil && otherLoopContainerStep.Entrypoint == nil {
		return false
	} else if l.Entrypoint != nil {
		if slices.Compare(*l.Entrypoint, *otherLoopContainerStep.Entrypoint) != 0 {
			return false
		}
	}

	if l.Env == nil && otherLoopContainerStep.Env != nil || l.Env != nil && otherLoopContainerStep.Env == nil {
		return false
	} else if l.Env != nil {
		if !reflect.DeepEqual(*l.Env, *otherLoopContainerStep.Env) {
			return false
		}
	}

	return utils.BoolPtrEqual(l.Until, otherLoopContainerStep.Until) &&
		utils.PtrEqual(l.Image, otherLoopContainerStep.Image) &&
		utils.PtrEqual(l.Source, otherLoopContainerStep.Source) &&
		utils.PtrEqual(l.CpuShares, otherLoopContainerStep.CpuShares) &&
		utils.PtrEqual(l.Memory, otherLoopContainerStep.Memory) &&
		utils.PtrEqual(l.MemoryReservation, otherLoopContainerStep.MemoryReservation) &&
		utils.PtrEqual(l.MemorySwap, otherLoopContainerStep.MemorySwap) &&
		utils.PtrEqual(l.MemorySwappiness, otherLoopContainerStep.MemorySwappiness) &&
		utils.BoolPtrEqual(l.ReadOnly, otherLoopContainerStep.ReadOnly) &&
		utils.PtrEqual(l.User, otherLoopContainerStep.User) &&
		utils.PtrEqual(l.Workdir, otherLoopContainerStep.Workdir)
}

func (*LoopContainerStep) GetType() string {
	return schema.BlockTypePipelineStepContainer
}

func (l *LoopContainerStep) UpdateInput(input Input, evalContext *hcl.EvalContext) (Input, error) {

	result, diags := simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), input, evalContext, schema.AttributeTypeImage, l.Image)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeSource, l.Source)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeCpuShares, l.CpuShares)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeMemory, l.Memory)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeMemoryReservation, l.MemoryReservation)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeMemorySwap, l.MemorySwap)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeMemorySwappiness, l.MemorySwappiness)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeUser, l.User)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = simpleTypeInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeWorkdir, l.Workdir)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = stringSliceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeCmd, l.Cmd)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = stringSliceInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeEntrypoint, l.Entrypoint)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	result, diags = stringMapInputFromAttribute(l.GetUnresolvedAttributes(), result, evalContext, schema.AttributeTypeEnv, l.Env)
	if len(diags) > 0 {
		return nil, error_helpers.BetterHclDiagsToError("container", diags)
	}

	return result, nil
}

func (l *LoopContainerStep) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := l.LoopStep.SetAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeImage, schema.AttributeTypeSource, schema.AttributeTypeUser, schema.AttributeTypeWorkdir:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeCmd, schema.AttributeTypeEntrypoint:
			fieldName := strcase.ToCamel(name)
			stepDiags := setStringSliceAttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}
		case schema.AttributeTypeEnv:
			val, stepDiags := dependsOnFromExpressionsWithResultControl(attr, evalContext, l, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

			if val == cty.NilVal {
				continue
			}

			env, err := hclhelpers.CtyToGoMapString(val)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid env",
					Detail:   "Invalid env in the step loop block",
					Subject:  &attr.Range,
				})
				continue
			}

			l.Env = &env

		case schema.AttributeTypeCpuShares, schema.AttributeTypeMemory, schema.AttributeTypeMemoryReservation, schema.AttributeTypeMemorySwap, schema.AttributeTypeMemorySwappiness:
			fieldName := strcase.ToCamel(name)
			stepDiags := setInt64AttributeWithResultReference(attr, evalContext, l, fieldName, true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

		case schema.AttributeTypeReadOnly:
			stepDiags := setBoolAttributeWithResultReference(attr, evalContext, l, "ReadOnly", true, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

		case schema.AttributeTypeUntil:
			// already handled in SetAttributes
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid attribute",
				Detail:   "Invalid attribute '" + name + "' in the step loop block",
				Subject:  &attr.Range,
			})
		}

	}
	return diags
}
