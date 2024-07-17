package output

import (
	"context"
	"github.com/charmbracelet/huh/spinner"
	"sync"
)

var PipelineProgress *Progress

type Progress struct {
	spinner  *spinner.Spinner
	cancel   context.CancelFunc
	mu       sync.Mutex
	status   string
	isActive bool
}

func NewProgress(initialText string) *Progress {
	return &Progress{
		status: initialText,
	}
}

func (p *Progress) Start() {
	if p.spinner == nil {
		p.spinner = spinner.New()
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.isActive = true
	p.spinner.Title(p.status)
	go func() {
		p.spinner.Context(ctx).Run()
	}()
}

func (p *Progress) Stop() {
	p.isActive = false
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *Progress) Update(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = msg
	if p.isActive {
		p.spinner.Title(p.status)
	}
}

func (p *Progress) IsActive() bool {
	return p.isActive
}
