//nolint:forbidigo // test file
package fqueue

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFunctionQueue(t *testing.T) {

	assert := assert.New(t)

	// Create a new function queue
	// fq := NewFunctionQueueWithSize(4)
	fq := NewFunctionQueue()

	fq.Callback = make(chan string)

	// Add a function call to the queue
	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	// Start the function queue
	fq.Execute()

	fq.Enqueue(func() {
		fmt.Println("start 3 second sleep")
		time.Sleep(3 * time.Second)
		fmt.Println("** end 3 second sleep")
	})

	time.Sleep(1 * time.Second)

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep A")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep B")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep C")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep D")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep E")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep F")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep G")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep G")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	fq.Enqueue(func() {
		fmt.Println("start 1 second sleep")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
	})

	res := <-fq.Callback
	assert.Equal("done", res)
	fmt.Println("res is " + res)
}
