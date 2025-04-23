package resources

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type PipelineStepInput struct {
	PipelineStepBase

	InputType  string  `json:"type" cty:"type"`
	Prompt     *string `json:"prompt" cty:"prompt"`
	OptionList []PipelineStepInputOption

	// Notifier cty.Value `json:"-" cty:"notify"`
	Notifier NotifierImpl `json:"notify" cty:"-"`

	// overrides
	Cc      []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc     []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel *string  `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Subject *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	To      []string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}

func (p *PipelineStepInput) Equals(other PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(other) {
		return true
	}

	if p == nil && !helpers.IsNil(other) || p != nil && helpers.IsNil(other) {
		return false
	}

	pOther, ok := other.(*PipelineStepInput)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&pOther.PipelineStepBase) {
		return false
	}

	if len(p.OptionList) != len(pOther.OptionList) {
		return false
	}

	for i, opt := range p.OptionList {
		if !opt.Equals(&pOther.OptionList[i]) {
			return false
		}
	}

	return p.Name == other.GetName() &&
		p.InputType == pOther.InputType &&
		utils.PtrEqual(p.Prompt, pOther.Prompt) &&
		helpers.StringSliceEqualIgnoreOrder(p.Cc, pOther.Cc) &&
		helpers.StringSliceEqualIgnoreOrder(p.Bcc, pOther.Bcc) &&
		utils.PtrEqual(p.Channel, pOther.Channel) &&
		utils.PtrEqual(p.Description, pOther.Description) &&
		utils.PtrEqual(p.Subject, pOther.Subject) &&
		utils.PtrEqual(p.Title, pOther.Title) &&
		helpers.StringSliceEqualIgnoreOrder(p.To, pOther.To) &&
		p.Notifier.Equals(&pOther.Notifier)

}

func (p *PipelineStepInput) GetInputs2(evalContext *hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {
	var connectionDependenciesAll []ConnectionDependency

	results := map[string]interface{}{}
	results[schema.AttributeTypeType] = p.InputType

	var diags hcl.Diagnostics

	// prompt
	promptValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypePrompt, p.Prompt)
	if diags.HasErrors() {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypePrompt] = promptValue
	connectionDependenciesAll = append(connectionDependenciesAll, connectionDependencies...)

	// channel
	channelValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeChannel, p.Channel)
	if diags.HasErrors() {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeChannel] = channelValue
	connectionDependenciesAll = append(connectionDependenciesAll, connectionDependencies...)

	// subject
	subjectValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeSubject, p.Subject)
	if diags.HasErrors() {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeSubject] = subjectValue
	connectionDependenciesAll = append(connectionDependenciesAll, connectionDependencies...)

	// to
	toValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeTo, p.To)
	if diags.HasErrors() {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeTo] = toValue
	connectionDependenciesAll = append(connectionDependenciesAll, connectionDependencies...)

	// cc
	ccValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeCc, p.Cc)
	if diags.HasErrors() {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeCc] = ccValue
	connectionDependenciesAll = append(connectionDependenciesAll, connectionDependencies...)

	// bcc
	bccValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeBcc, p.Bcc)
	if diags.HasErrors() {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeBcc] = bccValue
	connectionDependenciesAll = append(connectionDependenciesAll, connectionDependencies...)

	// options
	var err error
	var resolvedOpts []PipelineStepInputOption

	if p.UnresolvedAttributes[schema.AttributeTypeOptions] != nil {
		// attribute needs resolving, this case may happen if we specify the entire option as an attribute
		var opts cty.Value
		diags := gohcl.DecodeExpression(p.UnresolvedAttributes[schema.AttributeTypeOptions], evalContext, &opts)
		if diags.HasErrors() {
			return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
		}
		resolvedOpts, err = CtyValueToPipelineStepInputOptionList(opts)
		if err != nil {
			return nil, nil, perr.BadRequestWithMessage(p.Name + ": unable to parse options attribute: " + err.Error())
		}
	} else if len(p.OptionList) > 0 {
		// This may happen if we specify the options as blocks
		resolvedOpts = make([]PipelineStepInputOption, len(p.OptionList))

		for i, opt := range p.OptionList {
			var diags hcl.Diagnostics
			newOpt, diags := opt.Resolve(evalContext)
			if diags.HasErrors() {
				return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
			}
			resolvedOpts[i] = *newOpt
		}
	}

	results[schema.AttributeTypeOptions] = resolvedOpts

	// notifier
	if attr, ok := p.UnresolvedAttributes[schema.AttributeTypeNotifier]; !ok {
		results[schema.AttributeTypeNotifier] = p.Notifier
	} else {
		notifierCtyVal, moreDiags := attr.Value(evalContext)
		if moreDiags.HasErrors() {
			return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, moreDiags)
		}

		notifier, err := ctyValueToPipelineStepNotifierValueMap(notifierCtyVal)
		if err != nil {
			return nil, nil, perr.BadRequestWithMessage(p.Name + ": unable to parse notifier attribute: " + err.Error())
		}
		results[schema.AttributeTypeNotifier] = notifier
	}

	return results, connectionDependenciesAll, nil

}

func (p *PipelineStepInput) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	res, _, err := p.GetInputs2(evalContext)
	return res, err
}

func (p *PipelineStepInput) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeType:
			stepDiags := setStringAttribute(attr, evalContext, p, "InputType", false)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypePrompt, schema.AttributeTypeChannel, schema.AttributeTypeSubject:

			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, structFieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeCc, schema.AttributeTypeBcc, schema.AttributeTypeTo:
			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringSliceAttribute(attr, evalContext, p, structFieldName, false)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeOptions:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				opts, ctyErr := CtyValueToPipelineStepInputOptionList(val)
				if ctyErr != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeOptions + " attribute to InputOption slice",
						Detail:   ctyErr.Error(),
						Subject:  &attr.Range,
					})
					continue
				}
				p.OptionList = append(p.OptionList, opts...)
			}
		case schema.AttributeTypeNotifier:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.Notifier, err = ctyValueToPipelineStepNotifierValueMap(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeNotifier + " attribute to InputNotifier",
						Detail:   err.Error(),
						Subject:  &attr.Range,
					})
				}
			}

		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for Input Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepInput) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.PipelineStepBase.SetBlockConfig(blocks, evalContext)

	hasAttrOptions := len(p.OptionList) > 0 || p.UnresolvedAttributes["options"] != nil
	optionIndex := 0
	for _, b := range blocks {
		switch b.Type {
		case schema.BlockTypeOption:
			opt := PipelineStepInputOption{
				PipelineStepBase:     &p.PipelineStepBase,
				UnresolvedAttributes: make(map[string]hcl.Expression),
			}
			if hasAttrOptions {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Option blocks and options attribute are mutually exclusive",
					Subject:  &b.DefRange,
				})
				continue
			}

			opt.OptionLabel = utils.ToPointer(b.Labels[0])

			optAttributes, moreDiags := b.Body.JustAttributes()

			// This error is not a "handling" error, if we fail to get the attributes it's a fatal error don't call "HandleDecodeDiags"
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			moreDiags = opt.SetAttributes(optAttributes, evalContext)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			p.OptionList = append(p.OptionList, opt)
			optionIndex++
		}
	}

	return diags
}

func (p *PipelineStepInput) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	// validate type
	if !constants.IsValidInputType(p.InputType) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Attribute " + schema.AttributeTypeType + " specified with invalid value " + p.InputType,
		})
	}

	// check for and validate style on options
	for _, o := range p.OptionList {
		if !helpers.IsNil(o.Style) && !constants.IsValidInputStyleType(*o.Style) {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute " + schema.AttributeTypeStyle + " specified with invalid value " + *o.Style,
			})
		}
	}

	return diags
}

func ctyValueToPipelineStepNotifierValueMap(value cty.Value) (NotifierImpl, error) {
	notifier := NotifierImpl{}

	if value == cty.NilVal {
		return notifier, perr.BadRequestWithMessage("notifier value is nil")
	}

	if !value.Type().IsMapType() && !value.Type().IsObjectType() {
		return notifier, perr.BadRequestWithMessage("notifier value must be a reference to a notifier resource")
	}

	valueMap := value.AsValueMap()
	notifiesCty := valueMap[schema.AttributeTypeNotifies]

	if notifiesCty == cty.NilVal {
		return notifier, nil
	}

	notifiesCtySlice := notifiesCty.AsValueSlice()

	for _, notifyCty := range notifiesCtySlice {
		n, err := ctyValueToNotify(notifyCty)
		if err != nil {
			return notifier, err
		}
		notifier.Notifies = append(notifier.Notifies, n)
	}

	if valueMap["full_name"] != cty.NilVal {
		notifier.FullName = valueMap["full_name"].AsString()
	}
	if valueMap["short_name"] != cty.NilVal {
		notifier.ShortName = valueMap["short_name"].AsString()
	}
	if valueMap["notifier_name"] != cty.NilVal {
		notifier.NotifierName = valueMap["notifier_name"].AsString()
	}

	return notifier, nil
}

func ctyValueToNotify(val cty.Value) (Notify, error) {

	n := Notify{}

	if val.IsNull() {
		return n, nil
	}

	valMap := val.AsValueMap()

	cc := valMap[schema.AttributeTypeCc]
	if cc != cty.NilVal {
		ccSlice := cc.AsValueSlice()
		for _, c := range ccSlice {
			n.Cc = append(n.Cc, c.AsString())
		}
	}

	bcc := valMap["bcc"]
	if bcc != cty.NilVal {
		bccSlice := bcc.AsValueSlice()
		for _, b := range bccSlice {
			n.Bcc = append(n.Bcc, b.AsString())
		}
	}

	channel := valMap["channel"]
	if channel != cty.NilVal {
		channel := channel.AsString()
		n.Channel = &channel
	}

	description := valMap["description"]
	if description != cty.NilVal {
		description := description.AsString()
		n.Description = &description
	}

	subject := valMap["subject"]
	if subject != cty.NilVal {
		subject := subject.AsString()
		n.Subject = &subject
	}

	title := valMap["title"]
	if title != cty.NilVal {
		title := title.AsString()
		n.Title = &title
	}

	to := valMap["to"]
	if to != cty.NilVal {
		toSlice := to.AsValueSlice()
		for _, t := range toSlice {
			n.To = append(n.To, t.AsString())
		}
	}

	integration := valMap["integration"]

	if integration != cty.NilVal {
		integration, err := integrationFromCtyValue(integration)
		if err != nil {
			return n, err
		}
		n.Integration = integration
	}

	return n, nil
}

type PipelineStepInputOption struct {
	// circular link to its "parent"
	PipelineStepBase *PipelineStepBase `json:"-"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`

	OptionLabel *string `json:"-"` // the label on the option block
	Label       *string `json:"label,omitempty" hcl:"label,optional"`
	Value       *string `json:"value,omitempty" hcl:"value,optional"`
	Selected    *bool   `json:"selected,omitempty" hcl:"selected,optional"`
	Style       *string `json:"style,omitempty" hcl:"style,optional"`
}

func (p *PipelineStepInputOption) AppendDependsOn(dependsOn ...string) {
	p.PipelineStepBase.AppendDependsOn(dependsOn...)
}

func (p *PipelineStepInputOption) AppendCredentialDependsOn(...string) {
	// not implemented
}

func (p *PipelineStepInputOption) AppendConnectionDependsOn(...string) {
	// not implemented
}

func (p *PipelineStepInputOption) GetPipeline() *Pipeline {
	return p.PipelineStepBase.GetPipeline()
}

func (p *PipelineStepInputOption) AddUnresolvedAttribute(name string, expr hcl.Expression) {
	p.UnresolvedAttributes[name] = expr
}

func (p *PipelineStepInputOption) Resolve(evalContext *hcl.EvalContext) (*PipelineStepInputOption, hcl.Diagnostics) {

	newOpt := &PipelineStepInputOption{}

	// make a copy, don't point to the same memory
	if p.Label != nil {
		newOpt.Label = utils.ToPointer(*p.Label)
	} else if p.UnresolvedAttributes[schema.AttributeTypeLabel] != nil {
		attr := p.UnresolvedAttributes[schema.AttributeTypeLabel]
		val, diags := attr.Value(evalContext)
		if diags.HasErrors() {
			return nil, diags
		}

		if val != cty.NilVal {
			valString, err := hclhelpers.CtyToString(val)
			if err != nil {
				return nil, hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeLabel + " attribute to string",
						Detail:   "Unable to parse " + schema.AttributeTypeLabel + " attribute to string",
						Subject:  attr.Range().Ptr(),
					},
				}
			}

			newOpt.Label = utils.ToPointer(valString)
		}
	}

	if p.Value != nil {
		newOpt.Value = utils.ToPointer(*p.Value)
	} else if p.UnresolvedAttributes[schema.AttributeTypeValue] != nil {
		val, diags := p.UnresolvedAttributes[schema.AttributeTypeValue].Value(evalContext)
		if diags.HasErrors() {
			return nil, diags
		}

		if val != cty.NilVal && val.Type() == cty.String {
			newOpt.Value = utils.ToPointer(val.AsString())
		}
	}

	if p.Selected != nil {
		newOpt.Selected = utils.ToPointer(*p.Selected)
	} else if p.UnresolvedAttributes[schema.AttributeTypeSelected] != nil {
		val, diags := p.UnresolvedAttributes[schema.AttributeTypeLabel].Value(evalContext)
		if diags.HasErrors() {
			return nil, diags
		}

		if val != cty.NilVal && val.Type() == cty.Bool {
			newOpt.Selected = utils.ToPointer(val.True())
		}
	}

	if p.Style != nil {
		newOpt.Style = utils.ToPointer(*p.Style)
	} else if p.UnresolvedAttributes[schema.AttributeTypeStyle] != nil {
		val, diags := p.UnresolvedAttributes[schema.AttributeTypeStyle].Value(evalContext)
		if diags.HasErrors() {
			return nil, diags
		}

		if val != cty.NilVal && val.Type() == cty.String {
			newOpt.Style = utils.ToPointer(val.AsString())
		}
	}

	// If the Value is nil, get the value from the option label (if it's a block option)
	if newOpt.Value == nil {
		newOpt.Value = p.OptionLabel
	}

	return newOpt, hcl.Diagnostics{}
}

func (p *PipelineStepInputOption) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := hcl.Diagnostics{}

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeLabel, schema.AttributeTypeValue, schema.AttributeTypeStyle:
			structFieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, structFieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

		case schema.AttributeTypeSelected:
			stepDiags := setBoolAttribute(attr, evalContext, p, "Selected", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
			}

		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unsupported attribute for Input Option: " + attr.Name,
				Subject:  &attr.Range,
			})
		}
	}

	return diags
}

func (p *PipelineStepInputOption) Equals(other *PipelineStepInputOption) bool {
	if p == nil && other == nil {
		return true
	}

	if p == nil && other != nil || p != nil && other == nil {
		return false
	}

	for key, expr := range p.UnresolvedAttributes {
		otherExpr, ok := other.UnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	// reverse
	for key := range other.UnresolvedAttributes {
		if _, ok := p.UnresolvedAttributes[key]; !ok {
			return false
		}
	}

	return utils.PtrEqual(p.Label, other.Label) &&
		utils.PtrEqual(p.Value, other.Value) &&
		utils.BoolPtrEqual(p.Selected, other.Selected) &&
		utils.PtrEqual(p.Style, other.Style) &&
		utils.PtrEqual(p.OptionLabel, other.OptionLabel)
}

func CtyValueToPipelineStepInputOptionList(value cty.Value) ([]PipelineStepInputOption, error) {
	var output []PipelineStepInputOption

	opts := value.AsValueSlice()

	for _, opt := range opts {
		valueMap := opt.AsValueMap()

		isValid := false
		option := PipelineStepInputOption{
			UnresolvedAttributes: make(map[string]hcl.Expression),
		}

		for k, v := range valueMap {
			switch k {
			case schema.AttributeTypeValue:
				if !v.IsNull() {
					isValid = true
					val := v.AsString()
					option.Value = &val
				}
			case schema.AttributeTypeLabel:
				if !v.IsNull() {
					label := v.AsString()
					option.Label = &label
				}
			case schema.AttributeTypeSelected:
				if !v.IsNull() && v.Type() == cty.Bool {
					isSelected := v.True()
					option.Selected = &isSelected
				}
			case schema.AttributeTypeStyle:
				if !v.IsNull() {
					s := v.AsString()
					option.Style = &s
				}
			default:
				return nil, perr.BadRequestWithMessage(k + " is not a valid attribute for input options")
			}
		}

		if isValid {
			output = append(output, option)
		} else {
			return nil, perr.BadRequestWithMessage("input options must declare a value")
		}
	}

	return output, nil
}

func (p *PipelineStepInputOption) Validate() hcl.Diagnostics {
	var diags hcl.Diagnostics

	// TODO: Figure out validation(s)

	return diags
}
