package resources

import (
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/pipe-fittings/error_helpers"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/schema"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/zclconf/go-cty/cty"
)

const (
	HttpMethodGet    = "get"
	HttpMethodPost   = "post"
	HttpMethodPut    = "put"
	HttpMethodDelete = "delete"
	HttpMethodPatch  = "patch"
)

var ValidHttpMethods = []string{
	HttpMethodGet,
	HttpMethodPost,
	HttpMethodPut,
	HttpMethodDelete,
	HttpMethodPatch,
}

type PipelineStepHttp struct {
	PipelineStepBase

	Url             *string                `json:"url" binding:"required"`
	Method          *string                `json:"method,omitempty"`
	CaCertPem       *string                `json:"ca_cert_pem,omitempty"`
	Insecure        *bool                  `json:"insecure,omitempty"`
	RequestBody     *string                `json:"request_body,omitempty"`
	RequestHeaders  map[string]interface{} `json:"request_headers,omitempty"`
	BasicAuthConfig *BasicAuthConfig       `json:"basic_auth,omitempty"`
}

func (p *PipelineStepHttp) Equals(iOther PipelineStep) bool {
	// If both pointers are nil, they are considered equal
	if p == nil && helpers.IsNil(iOther) {
		return true
	}

	if p == nil && !helpers.IsNil(iOther) || p != nil && helpers.IsNil(iOther) {
		return false
	}

	other, ok := iOther.(*PipelineStepHttp)
	if !ok {
		return false
	}

	if !p.PipelineStepBase.Equals(&other.PipelineStepBase) {
		return false
	}

	if p.BasicAuthConfig != nil && !p.BasicAuthConfig.Equals(other.BasicAuthConfig) {
		return false
	} else if p.BasicAuthConfig == nil && other.BasicAuthConfig != nil {
		return false
	}

	return utils.PtrEqual(p.Url, other.Url) &&
		utils.PtrEqual(p.Method, other.Method) &&
		utils.PtrEqual(p.CaCertPem, other.CaCertPem) &&
		utils.BoolPtrEqual(p.Insecure, other.Insecure) &&
		utils.PtrEqual(p.RequestBody, other.RequestBody) &&
		reflect.DeepEqual(p.RequestHeaders, other.RequestHeaders)
}

func (p *PipelineStepHttp) GetInputs(evalContext *hcl.EvalContext) (map[string]interface{}, error) {
	res, _, err := p.GetInputs2(evalContext)
	return res, err
}

func (p *PipelineStepHttp) GetInputs2(evalContext *hcl.EvalContext) (map[string]interface{}, []ConnectionDependency, error) {

	var diags hcl.Diagnostics
	var allConnectionDependencies []ConnectionDependency

	results, err := p.GetBaseInputs(evalContext)
	if err != nil {
		return nil, nil, err
	}

	// url
	urlValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeUrl, p.Url)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeUrl] = urlValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// method
	methodValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeMethod, p.Method)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeMethod] = methodValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// ca_cert_pem
	caCertPemValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeCaCertPem, p.CaCertPem)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeCaCertPem] = caCertPemValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// insecure
	insecureValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeInsecure, p.Insecure)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeInsecure] = insecureValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// request_body
	requestBodyValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeRequestBody, p.RequestBody)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeRequestBody] = requestBodyValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	// request_headers
	requestHeadersValue, connectionDependencies, diags := decodeStepAttribute(p.UnresolvedAttributes, evalContext, p.Name, schema.AttributeTypeRequestHeaders, p.RequestHeaders)
	if len(diags) > 0 {
		return nil, nil, error_helpers.BetterHclDiagsToError(p.Name, diags)
	}
	results[schema.AttributeTypeRequestHeaders] = requestHeadersValue
	allConnectionDependencies = append(allConnectionDependencies, connectionDependencies...)

	if p.BasicAuthConfig != nil {
		basicAuth, diags := p.BasicAuthConfig.GetInputs(evalContext, p.UnresolvedAttributes)
		if diags.HasErrors() {
			return nil, nil, error_helpers.BetterHclDiagsToError(schema.BlockTypePipelineStep, diags)
		}
		basicAuthMap := make(map[string]interface{})
		basicAuthMap["Username"] = basicAuth.Username
		basicAuthMap["Password"] = basicAuth.Password
		results[schema.BlockTypePipelineBasicAuth] = basicAuthMap
	}
	results[schema.AttributeTypeStepName] = p.Name

	return results, allConnectionDependencies, nil
}

func (p *PipelineStepHttp) SetAttributes(hclAttributes hcl.Attributes, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diags := p.SetBaseAttributes(hclAttributes, evalContext)

	for name, attr := range hclAttributes {
		switch name {
		case schema.AttributeTypeUrl:
			fieldName := utils.CapitalizeFirst(name)
			stepDiags := setStringAttribute(attr, evalContext, p, fieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeMethod:
			fieldName := utils.CapitalizeFirst(name)

			stepDiags := setStringAttribute(attr, evalContext, p, fieldName, true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if types.SafeString(p.Method) != "" {
				if !helpers.StringSliceContains(ValidHttpMethods, strings.ToLower(types.SafeString(p.Method))) {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid HTTP method: " + types.SafeString(p.Method),
						Subject:  &attr.Range,
					})
					continue
				}
			}

		case schema.AttributeTypeCaCertPem:
			stepDiags := setStringAttribute(attr, evalContext, p, "CaCertPem", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeInsecure:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				if val.Type() != cty.Bool {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid value for insecure attribute",
						Subject:  &attr.Range,
					})
					continue
				}
				insecure := val.True()
				p.Insecure = &insecure
			}

		case schema.AttributeTypeRequestBody:
			stepDiags := setStringAttribute(attr, evalContext, p, "RequestBody", true)
			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

		case schema.AttributeTypeRequestHeaders:
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)

			if stepDiags.HasErrors() {
				diags = append(diags, stepDiags...)
				continue
			}

			if val != cty.NilVal {
				var err error
				p.RequestHeaders, err = hclhelpers.CtyToGoMapInterface(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse request_headers attribute",
						Subject:  &attr.Range,
					})
					continue
				}
			}
		default:
			if !p.IsBaseAttribute(name) {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Unsupported attribute for HTTP Step: " + attr.Name,
					Subject:  &attr.Range,
				})
			}
		}
	}

	return diags
}

func (p *PipelineStepHttp) SetBlockConfig(blocks hcl.Blocks, evalContext *hcl.EvalContext) hcl.Diagnostics {

	diags := p.PipelineStepBase.SetBlockConfig(blocks, evalContext)

	basicAuthConfig := &BasicAuthConfig{}

	if basicAuthBlocks := blocks.ByType()[schema.BlockTypePipelineBasicAuth]; len(basicAuthBlocks) > 0 {
		if len(basicAuthBlocks) > 1 {
			return hcl.Diagnostics{&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Multiple basic_auth blocks found for step http",
			}}
		}
		basicAuthBlock := basicAuthBlocks[0]

		var attributes hcl.Attributes

		attributes, diags = basicAuthBlock.Body.JustAttributes()
		if len(diags) > 0 {
			return diags
		}

		if attr, exists := attributes[schema.AttributeTypeUsername]; exists {
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				username, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypeUsername + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				basicAuthConfig.Username = username
			}

		}

		if attr, exists := attributes[schema.AttributeTypePassword]; exists {
			val, stepDiags := dependsOnFromExpressions(attr, evalContext, p)
			if len(stepDiags) > 0 {
				diags = append(diags, stepDiags...)
				return diags
			}

			if val != cty.NilVal {
				password, err := hclhelpers.CtyToString(val)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Unable to parse " + schema.AttributeTypePassword + " attribute to string",
						Subject:  &attr.Range,
					})
					return diags
				}
				basicAuthConfig.Password = password
			}

		}
		p.BasicAuthConfig = basicAuthConfig
	}

	return diags
}

func (p *PipelineStepHttp) Validate() hcl.Diagnostics {
	diags := p.ValidateBaseAttributes()
	return diags
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`

	UnresolvedAttributes map[string]hcl.Expression `json:"-"`
}

func (b *BasicAuthConfig) GetInputs(evalContext *hcl.EvalContext, unresolvedAttributes map[string]hcl.Expression) (*BasicAuthConfig, hcl.Diagnostics) {

	newBasicAuthConfig := &BasicAuthConfig{}

	var username, password string
	if unresolvedAttributes[schema.AttributeTypeUsername] != nil {
		diags := gohcl.DecodeExpression(unresolvedAttributes[schema.AttributeTypeUsername], evalContext, &username)
		if diags.HasErrors() {
			return nil, diags
		}
		newBasicAuthConfig.Username = username
	} else {
		newBasicAuthConfig.Username = b.Username
	}

	if unresolvedAttributes[schema.AttributeTypePassword] != nil {
		diags := gohcl.DecodeExpression(unresolvedAttributes[schema.AttributeTypePassword], evalContext, &password)
		if diags.HasErrors() {
			return nil, diags
		}
		newBasicAuthConfig.Password = password
	} else {
		newBasicAuthConfig.Password = b.Password
	}

	return newBasicAuthConfig, nil
}

func (b *BasicAuthConfig) Equals(other *BasicAuthConfig) bool {

	if b == nil && other == nil {
		return false
	}

	if b == nil && other != nil || b != nil && other == nil {
		return false
	}

	// Compare UnresolvedAttributes (map comparison)
	if len(b.UnresolvedAttributes) != len(other.UnresolvedAttributes) {
		return false
	}

	for key, expr := range b.UnresolvedAttributes {
		otherExpr, ok := other.UnresolvedAttributes[key]
		if !ok || !hclhelpers.ExpressionsEqual(expr, otherExpr) {
			return false
		}
	}

	// and reverse
	for key := range other.UnresolvedAttributes {
		if _, ok := b.UnresolvedAttributes[key]; !ok {
			return false
		}
	}

	return b.Username == other.Username &&
		b.Password == other.Password
}
