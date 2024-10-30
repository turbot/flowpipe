package resources

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

type Notifier interface {
	modconfig.HclResource
	modconfig.ResourceWithMetadata

	CtyValue() (cty.Value, error)
	GetNotifierImpl() *NotifierImpl
	GetNotifies() []Notify
	SetFileReference(fileName string, startLineNumber int, endLineNumber int)

	Equals(Notifier) bool
}

type NotifierImpl struct {
	modconfig.HclResourceImpl          `json:"-"`
	modconfig.ResourceWithMetadataImpl `json:"-"`

	NotifierName string   `json:"notifier_name" cty:"notifier_name" hcl:"notifier_name"`
	Notifies     []Notify `json:"notifies" cty:"notifies" hcl:"notifies"`

	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	FileName        string `json:"-" cty:"-" hcl:"-"`
	StartLineNumber int    `json:"-" cty:"-" hcl:"-"`
	EndLineNumber   int    `json:"-" cty:"-" hcl:"-"`
}

func (n *NotifierImpl) Equals(other Notifier) bool {

	if n == nil && helpers.IsNil(other) {
		return true
	}

	if n == nil && !helpers.IsNil(other) || !helpers.IsNil(other) && n == nil {
		return false
	}

	if len(n.Notifies) != len(other.GetNotifierImpl().Notifies) {
		return false
	}

	for i, notify := range n.Notifies {
		if !notify.Equals(&other.GetNotifierImpl().Notifies[i]) {
			return false
		}
	}

	return n.FileName == other.GetNotifierImpl().FileName &&
		n.StartLineNumber == other.GetNotifierImpl().StartLineNumber &&
		n.EndLineNumber == other.GetNotifierImpl().EndLineNumber
}

func (n *NotifierImpl) SetFileReference(fileName string, startLineNumber int, endLineNumber int) {
	n.FileName = fileName
	n.StartLineNumber = startLineNumber
	n.EndLineNumber = endLineNumber
}

func (n *NotifierImpl) GetNotifies() []Notify {
	return n.Notifies
}

func (n *NotifierImpl) GetNotifierImpl() *NotifierImpl {
	return n
}

func DefaultNotifiers(defaultHttpIntegration Integration) (map[string]Notifier, error) {
	notifiers := make(map[string]Notifier)

	description := "Default notifier"

	notifier := NotifierImpl{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName:        "default",
			ShortName:       "default",
			UnqualifiedName: "default",
			Description:     &description,
		},
		NotifierName: "default",
	}

	notify := Notify{
		Integration: defaultHttpIntegration,
	}
	notifier.Notifies = []Notify{notify}

	notifiers["default"] = &notifier

	return notifiers, nil
}

func (n *NotifierImpl) CtyValue() (cty.Value, error) {
	notifies := []any{}

	for _, notify := range n.Notifies {
		mapInterface, err := notify.MapInterface()
		if err != nil {
			return cty.NilVal, err
		}

		notifies = append(notifies, mapInterface)
	}

	notifierMap := make(map[string]interface{}, 1)
	notifierMap["notifies"] = notifies
	notifierMap["title"] = n.Title
	notifierMap["full_name"] = n.FullName
	notifierMap["short_name"] = n.ShortName
	notifierMap["name"] = n.ShortName
	notifierMap["unqualified_name"] = n.UnqualifiedName
	if n.Description != nil {
		notifierMap["description"] = *n.Description
	}
	notifierMap["resource_type"] = "notifier"
	notifierMap["notifier_name"] = n.NotifierName

	notifierCtyVal, err := hclhelpers.ConvertInterfaceToCtyValue(notifierMap)
	return notifierCtyVal, err
}

func (n *NotifierImpl) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	notifies := n.GetNotifies()
	if len(notifies) < 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  schema.BlockTypeNotifier + " must have at least one " + schema.BlockTypeNotify + " block to send the request to: " + n.Name(),
		})
	}
	return diags
}

// CustomType implements custom_type.CustomType interface
func (n *NotifierImpl) CustomType() {
}

type Notify struct {
	// required to allow partial decoding
	Remain hcl.Body `hcl:",remain" json:"-"`

	Integration Integration `json:"integration"`

	Cc          []string `json:"cc,omitempty" cty:"cc" hcl:"cc,optional"`
	Bcc         []string `json:"bcc,omitempty" cty:"bcc" hcl:"bcc,optional"`
	Channel     *string  `json:"channel,omitempty" cty:"channel" hcl:"channel,optional"`
	Description *string  `json:"description,omitempty" cty:"description" hcl:"description,optional"`
	Subject     *string  `json:"subject,omitempty" cty:"subject" hcl:"subject,optional"`
	Title       *string  `json:"title,omitempty" cty:"title" hcl:"title,optional"`
	To          []string `json:"to,omitempty" cty:"to" hcl:"to,optional"`
}

func (n *Notify) Equals(other *Notify) bool {

	if n == nil && other == nil {
		return true
	}

	if n == nil && other != nil || n != nil && other == nil {
		return false
	}

	return helpers.StringSliceEqualIgnoreOrder(n.Cc, other.Cc) &&
		helpers.StringSliceEqualIgnoreOrder(n.Bcc, other.Bcc) &&
		helpers.StringSliceEqualIgnoreOrder(n.To, other.To) &&
		utils.PtrEqual(n.Channel, other.Channel) &&
		utils.PtrEqual(n.Description, other.Description) &&
		utils.PtrEqual(n.Subject, other.Subject) &&
		utils.PtrEqual(n.Title, other.Title) &&
		n.Integration.Equals(other.Integration)
}

// UnmarshalJSON custom unmarshaller for Notify
func (n *Notify) UnmarshalJSON(data []byte) error {
	// Define a struct that mirrors Notify but with Integration as json.RawMessage
	// to defer its unmarshalling
	type Alias Notify
	temp := &struct {
		Integration json.RawMessage `json:"integration"`
		*Alias
	}{
		Alias: (*Alias)(n), // Cast n to Alias type to unmarshal other fields
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Now, determine the type of Integration and unmarshal accordingly
	// This assumes your JSON contains some type identifier for the integration.
	// You might need a temporary struct to peek into the raw JSON to read that identifier
	var typeIndicator struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(temp.Integration, &typeIndicator); err != nil {
		return err
	}

	switch typeIndicator.Type {
	case schema.IntegrationTypeSlack:
		var slackIntegration SlackIntegration
		if err := json.Unmarshal(temp.Integration, &slackIntegration); err != nil {
			return err
		}
		n.Integration = &slackIntegration
	case schema.IntegrationTypeEmail:
		var emailIntegration EmailIntegration
		if err := json.Unmarshal(temp.Integration, &emailIntegration); err != nil {
			return err
		}
		n.Integration = &emailIntegration
	case schema.IntegrationTypeHttp:
		var httpIntegration HttpIntegration
		if err := json.Unmarshal(temp.Integration, &httpIntegration); err != nil {
			return err
		}
		n.Integration = &httpIntegration
	case schema.IntegrationTypeMsTeams:
		var teamsIntegration MsTeamsIntegration
		if err := json.Unmarshal(temp.Integration, &teamsIntegration); err != nil {
			return err
		}
		n.Integration = &teamsIntegration
	default:
		return perr.InternalWithMessage(fmt.Sprintf("unknown integration type: %s", typeIndicator.Type))
	}

	return nil
}
func (n *Notify) MapInterface() (map[string]interface{}, error) {
	notifyMap := make(map[string]interface{})

	if n.Cc != nil {
		notifyMap[schema.AttributeTypeCc] = n.Cc
	}

	if n.Bcc != nil {
		notifyMap["bcc"] = n.Bcc
	}

	if n.Channel != nil {
		notifyMap["channel"] = *n.Channel
	}

	if n.Description != nil {
		notifyMap["description"] = *n.Description
	}

	if n.Subject != nil {
		notifyMap["subject"] = *n.Subject
	}

	if n.Title != nil {
		notifyMap["title"] = *n.Title
	}

	if n.To != nil {
		notifyMap["to"] = n.To
	}

	var err error
	notifyMap["integration"], err = n.Integration.MapInterface()
	if err != nil {
		return nil, err
	}

	return notifyMap, nil
}

func (n *Notify) CtyValue() (cty.Value, error) {
	notifyMap := make(map[string]interface{})

	var err error

	if n.Cc != nil {
		notifyMap[schema.AttributeTypeCc] = n.Cc
	}

	if n.Bcc != nil {
		notifyMap["bcc"] = n.Bcc
	}

	if n.Channel != nil {
		notifyMap["channel"] = *n.Channel
	}

	if n.Description != nil {
		notifyMap["description"] = *n.Description
	}

	if n.Subject != nil {
		notifyMap["subject"] = *n.Subject
	}

	if n.Title != nil {
		notifyMap["title"] = *n.Title
	}

	if n.To != nil {
		notifyMap["to"] = n.To
	}

	notifyMap["integration"], err = n.Integration.MapInterface()
	if err != nil {
		return cty.NilVal, err
	}

	ctyVal, err := hclhelpers.ConvertInterfaceToCtyValue(notifyMap)
	if err != nil {
		return cty.NilVal, err
	}

	return ctyVal, nil
}

func (n *Notify) Validate() hcl.Diagnostics {
	diags := hcl.Diagnostics{}

	if n.Integration != nil {
		integrationType := n.Integration.GetIntegrationType()

		if integrationType != schema.IntegrationTypeEmail && len(n.Cc) > 0 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute '" + schema.AttributeTypeCc + "' is not a valid attribute for " + integrationType + " type integration",
			})
		}

		if integrationType != schema.IntegrationTypeEmail && len(n.Bcc) > 0 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute '" + schema.AttributeTypeBcc + "' is not a valid attribute for " + integrationType + " type integration",
			})
		}

		if integrationType != schema.IntegrationTypeEmail && len(n.To) > 0 {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute '" + schema.AttributeTypeTo + "' is not a valid attribute for " + integrationType + " type integration",
			})
		}

		if integrationType != schema.IntegrationTypeEmail && n.Subject != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute '" + schema.AttributeTypeSubject + "' is not a valid attribute for " + integrationType + " type integration",
			})
		}

		if integrationType != schema.IntegrationTypeSlack && n.Channel != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Attribute '" + schema.AttributeTypeChannel + "' is not a valid attribute for " + integrationType + " type integration",
			})
		}
	}

	return diags
}

func (n *Notify) SetAttributes(body hcl.Body, evalCtx *hcl.EvalContext) hcl.Diagnostics {
	attribs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return diags
	}

	attr := attribs[schema.AttributeTypeIntegration]
	if attr == nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Missing required attribute: " + schema.AttributeTypeIntegration,
				Subject:  body.MissingItemRange().Ptr(),
			},
		}
	}

	integrationCtys, diags := attr.Expr.Value(evalCtx)
	if diags.HasErrors() {
		return diags
	}

	integration, err := integrationFromCtyValue(integrationCtys)
	if err != nil {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Error parsing " + schema.AttributeTypeIntegration + " attribute",
				Detail:   err.Error(),
				Subject:  body.MissingItemRange().Ptr(),
			},
		}
	}

	n.Integration = integration
	return diags
}
