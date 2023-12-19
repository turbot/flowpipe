package util

import (
	"context"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type HclExpressionMock struct {
	ValueFunc func(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics)
}

func (h *HclExpressionMock) Value(evalCtx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	if h.ValueFunc != nil {
		return h.ValueFunc(evalCtx)
	}

	res := map[string]cty.Value{
		"replace": cty.StringVal("me"),
	}
	return cty.ObjectVal(res), nil
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

type CommandBusMock struct {
	SendFunc         func(ctx context.Context, command interface{}) error
	SendWithLockFunc func(ctx context.Context, command interface{}, lock *sync.Mutex) error
}

func (c *CommandBusMock) Send(ctx context.Context, command interface{}) error {
	if c.SendFunc != nil {
		return c.SendFunc(ctx, command)
	}
	return nil
}

func (c *CommandBusMock) SendWithLock(ctx context.Context, command interface{}, lock *sync.Mutex) error {
	if c.SendWithLockFunc != nil {
		return c.SendFunc(ctx, command)
	}
	return nil
}
