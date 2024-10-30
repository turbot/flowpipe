package parse

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/schema"
)

var notifierBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     schema.AttributeTypeDescription,
			Required: false,
		},
		{
			Name:     schema.AttributeTypeTitle,
			Required: false,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: schema.BlockTypeNotify,
		},
	},
}

func DecodeNotifier(configPath string, block *hcl.Block, evalCtx *hcl.EvalContext) (*resources.NotifierImpl, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(block.Labels) != 1 {
		diags = hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "invalid notifier block - expected 1 label",
				Subject:  &block.DefRange,
			},
		}
		return nil, diags
	}

	notifierName := block.Labels[0]

	notifier := resources.NotifierImpl{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName:        notifierName,
			ShortName:       notifierName,
			UnqualifiedName: notifierName,
			DeclRange:       block.DefRange,
		},
		NotifierName: notifierName,
	}

	content, diags := block.Body.Content(notifierBlockSchema)
	if len(diags) > 0 {
		return nil, diags
	}

	for _, b := range content.Blocks {
		switch b.Type {
		case schema.BlockTypeNotify:
			notify := resources.Notify{}
			moreDiags := gohcl.DecodeBody(b.Body, evalCtx, &notify)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			moreDiags = notify.SetAttributes(b.Body, evalCtx)
			if len(moreDiags) > 0 {
				diags = append(diags, moreDiags...)
				continue
			}

			notifier.Notifies = append(notifier.Notifies, notify)

			validationDiags := notify.Validate()
			if len(validationDiags) > 0 {
				diags = append(diags, validationDiags...)
				continue
			}
		default:
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' in notifier", b.Type),
				Subject:  &b.DefRange,
			})
		}
	}

	moreDiags := resources.HclImplFromAttributes(&notifier.HclResourceImpl, content.Attributes, evalCtx)
	if len(moreDiags) > 0 {
		diags = append(diags, moreDiags...)
	}

	validationDiags := notifier.Validate()
	if len(validationDiags) > 0 {
		diags = append(diags, validationDiags...)
	}

	notifier.SetFileReference(block.DefRange.Filename, block.DefRange.Start.Line, block.DefRange.End.Line)

	return &notifier, diags
}
