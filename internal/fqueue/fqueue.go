package fqueue

import (
	"log/slog"
	"sync"
)

type FunctionCall func() error

// FunctionQueue is a queue of function calls (without parameter) that only the head and tail will be executed.
//
// For example: we are calling the same function 5 times in rapid sequence: f1, f2, f3, f4, f5.
//
// f1 will be executed, f2, f3, f4 are discarded and f5 will be executed after f1 has been completed. All 5 callers will be notified after f5 has been
// completed.
//
// This is a specific use case that we use to build Docker images. If there are multiple goroutines that are trying to build the same image, we only want to build just one image. While
// building the underlying Dockerfile might have changed, so we need to build one last time for the final request (f5). f2, f3, f4 are discarded as the Dockerfile has changed and those request
// to build is no longer relevant.
type FunctionQueue struct {
	Name             string
	CallbackChannels []chan error
	DropCount        int

	queue      chan FunctionCall
	isRunning  bool
	runLock    sync.Mutex
	queueCount int
}

func NewFunctionQueueWithSize(name string, size int) *FunctionQueue {
	return &FunctionQueue{
		Name:             name,
		DropCount:        0,
		queue:            make(chan FunctionCall, size),
		isRunning:        false,
		queueCount:       0,
		CallbackChannels: []chan error{},
	}
}

func NewFunctionQueue(name string) *FunctionQueue {
	return NewFunctionQueueWithSize(name, 2)
}

func (fq *FunctionQueue) RegisterCallback(callback chan error) {
	fq.CallbackChannels = append(fq.CallbackChannels, callback)
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
			err := fn() // Execute the function call
			slog.Debug("After execute", "queue", fq.Name, "queue_count", fq.queueCount)

			fq.runLock.Lock()
			fq.queueCount--

			if fq.queueCount == 0 {
				slog.Debug("No item in the queue .. returning", "queue", fq.Name, "queue_count", fq.queueCount)

				fq.isRunning = false

				for _, ch := range fq.CallbackChannels {
					ch <- err
				}

				for _, ch := range fq.CallbackChannels {
					close(ch)
				}
				fq.CallbackChannels = []chan error{}

				fq.runLock.Unlock()

				return
			}
			fq.runLock.Unlock()
		}
	}()
}
