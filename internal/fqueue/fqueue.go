package fqueue

import (
	"log/slog"
	"sync"
)

type FunctionCall func()

type FunctionQueue struct {
	Name      string
	Callback  chan string
	DropCount int

	queue      chan FunctionCall
	isRunning  bool
	runLock    sync.Mutex
	queueCount int
}

func NewFunctionQueueWithSize(name string, size int) *FunctionQueue {
	return &FunctionQueue{
		Name:       name,
		DropCount:  0,
		queue:      make(chan FunctionCall, size),
		isRunning:  false,
		queueCount: 0,
	}
}

func NewFunctionQueue(name string) *FunctionQueue {
	return NewFunctionQueueWithSize(name, 2)
}

func (fq *FunctionQueue) Enqueue(fn FunctionCall) {
	fq.runLock.Lock()
	defer fq.runLock.Unlock()

	select {
	case fq.queue <- fn:
		// Function added to queue
		fq.queueCount++
		slog.Debug("Added to queue", "queue", fq.Name, "queue_count", fq.queueCount)

	default:
		// Queue is full, function call is dropped or handled otherwise
		slog.Debug("Dropped from queue", "queue", fq.Name, "queue_count", fq.queueCount)
		fq.DropCount++
	}
}

func (fq *FunctionQueue) Execute() {
	fq.runLock.Lock()
	if fq.isRunning {
		fq.runLock.Unlock()
		return
	}

	fq.runLock.Unlock()

	go func() {
		fq.runLock.Lock()
		fq.isRunning = true
		fq.runLock.Unlock()

		for fn := range fq.queue {
			slog.Debug("Before execute", "queue", fq.Name, "queue_count", fq.queueCount)
			fn() // Execute the function call
			slog.Debug("After execute", "queue", fq.Name, "queue_count", fq.queueCount)

			fq.runLock.Lock()
			fq.queueCount--

			if fq.queueCount == 0 {
				slog.Debug("No item in the queue .. returning", "queue", fq.Name, "queue_count", fq.queueCount)

				defer func() {
					fq.isRunning = false

					if fq.Callback != nil {
						fq.Callback <- "done"
					}
					fq.runLock.Unlock()
				}()

				return
			}
			fq.runLock.Unlock()
		}
	}()
}
