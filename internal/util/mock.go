package util

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type HclExpressionMock struct {
	EvalContext *hcl.EvalContext
}

func (h *HclExpressionMock) Value(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	h.EvalContext = evalCtx
	return cty.StringVal("test"), nil
}

func (*HclExpressionMock) Variables() []hcl.Traversal {
	return nil
}

func (*HclExpressionMock) Range() hcl.Range {
	return hcl.Range{}
}

func (*HclExpressionMock) StartRange() hcl.Range {
	return hcl.Range{}
}
