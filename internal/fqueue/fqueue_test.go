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
	fq := NewFunctionQueue("TestFunctionQueue")

	runMap := map[string]bool{}

	// Add a function call to the queue
	fq.Enqueue(func() error {
		fmt.Println("start sleep 0")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("** end sleep 0")
		runMap["0"] = true
		return nil
	})

	// Start the function queue
	fq.Execute()

	fq.Enqueue(func() error {
		fmt.Println("start sleep 1")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("** end sleep 1")
		runMap["1"] = true
		return nil
	})

	time.Sleep(25 * time.Millisecond)

	fq.Enqueue(func() error {
		fmt.Println("start sleep A")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("** end sleep A")
		runMap["A"] = true
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start sleep B")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("** end sleep B")
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start 1 second sleep C")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start 1 second sleep D")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start 1 second sleep E")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep F")
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start 1 second sleep G")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep G")
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start 1 second sleep")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
		return nil
	})

	fq.Enqueue(func() error {
		fmt.Println("start 1 second sleep")
		time.Sleep(1 * time.Second)
		fmt.Println("** end 1 second sleep")
		return nil
	})

	receiveChannel := make(chan error)
	fq.RegisterCallback(receiveChannel)

	err := <-receiveChannel
	assert.Nil(err)

	assert.Equal(3, len(runMap))
	assert.True(runMap["0"])
	assert.True(runMap["1"])
	assert.True(runMap["A"])
	assert.Equal(7, fq.DropCount)
}
