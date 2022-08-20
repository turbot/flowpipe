package primitive

import (
	"context"
	"errors"
)

type Input map[string]interface{}

type Output map[string]interface{}

type primitive struct {
	input Input
}

func (p *primitive) SetInput(i Input) error {
	p.input = i
	return nil
}

func (p *primitive) Input() Input {
	return p.input
}

func (p *primitive) Run(ctx context.Context) (Output, error) {
	return nil, errors.New("Run() not implemented")
}
