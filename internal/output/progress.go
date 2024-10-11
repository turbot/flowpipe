package output

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"

	spinner "github.com/turbot/pipe-fittings/statushooks"
)

var PipelineProgress *Progress

type Progress struct {
	spinner   *spinner.StatusSpinner
	mu        sync.Mutex
	status    string
	Semaphore *semaphore.Weighted
}

func NewProgress(initialText string) *Progress {
	return &Progress{
		status:    initialText,
		Semaphore: semaphore.NewWeighted(1),
		spinner:   spinner.NewStatusSpinnerHook(),
	}
}

func (p *Progress) Run(action func()) error {
	if err := p.Semaphore.Acquire(context.Background(), 1); err != nil {
		return err
	}
	defer p.Semaphore.Release(1)

	if p.spinner == nil {
		p.spinner = spinner.NewStatusSpinnerHook()
	}

	p.spinner.UpdateSpinnerMessage(p.status)
	p.spinner.Show()
	defer p.spinner.Hide()
	action()
	return nil
}

func (p *Progress) Update(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.spinner != nil && msg != p.status {
		p.status = msg
		p.spinner.UpdateSpinnerMessage(p.status)
	}
}
