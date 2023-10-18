package primitive

import (
	"context"

	"github.com/turbot/flowpipe/internal/util"
	"github.com/turbot/pipe-fittings/modconfig"
)

type Email struct {
	Input modconfig.Input
}

func (h *Email) ValidateInput(ctx context.Context, i modconfig.Input) error {
	return util.ValidateEmailInput(ctx, i)
}

func (h *Email) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	// Validate the inputs
	if err := h.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	return util.RunSendEmail(ctx, input)
}
