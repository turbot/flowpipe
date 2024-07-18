package output

import (
	"context"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"golang.org/x/sync/semaphore"
)

var PipelineProgress *Progress

type Progress struct {
	spinner   *spinner.Spinner
	mu        sync.Mutex
	status    string
	Semaphore *semaphore.Weighted
}

func NewProgress(initialText string) *Progress {
	return &Progress{
		status:    initialText,
		Semaphore: semaphore.NewWeighted(1),
		spinner:   spinner.New(),
	}
}

func (p *Progress) Run(action func()) error {
	if err := p.Semaphore.Acquire(context.Background(), 1); err != nil {
		return err
	}
	defer p.Semaphore.Release(1)

	if p.spinner == nil {
		p.spinner = spinner.New()
	}

	p.spinner.Title(p.status).Action(action)
	return p.spinner.Run()
}

func (p *Progress) Update(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = msg
	if p.spinner != nil {
		p.spinner.Title(p.status)
	}
}
