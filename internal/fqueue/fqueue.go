package fqueue

import (
	"fmt"
	"sync"
)

type FunctionCall func()

type FunctionQueue struct {
	Callback chan string

	queue     chan FunctionCall
	isRunning bool
	runLock   sync.Mutex
}

func NewFunctionQueueWithSize(size int) *FunctionQueue {
	return &FunctionQueue{
		queue:     make(chan FunctionCall, size),
		isRunning: false,
	}
}

func NewFunctionQueue() *FunctionQueue {
	return NewFunctionQueueWithSize(2)
}

func (fq *FunctionQueue) Enqueue(fn FunctionCall) {
	select {
	case fq.queue <- fn:
		// Function added to queue
	default:
		// Queue is full, function call is dropped or handled otherwise
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
			fmt.Println("Before execute")
			fn() // Execute the function call
			fmt.Println("After execute")
		}

		fmt.Println("Function queue finished trying to unlock isRuning")
		fq.runLock.Lock()
		fq.isRunning = false
		fq.runLock.Unlock()
		fmt.Println("Function queue finished mutex unlocked")

		if fq.Callback != nil {
			fq.Callback <- "done"
		}
	}()
}
