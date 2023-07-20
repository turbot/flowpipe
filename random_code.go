package main

import (
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Conditional variable implementation that uses channels for notifications.
// Only supports .Broadcast() method, however supports timeout based Wait() calls
// unlike regular sync.Cond.
type Cond struct {
	L sync.Locker
	n unsafe.Pointer
}

func NewCond(l sync.Locker) *Cond {
	c := &Cond{L: l}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

// Waits for Broadcast calls. Similar to regular sync.Cond, this unlocks the underlying
// locker first, waits on changes and re-locks it before returning.
func (c *Cond) Wait() {
	n := c.NotifyChan()
	c.L.Unlock()
	<-n
	c.L.Lock()
}

// Same as Wait() call, but will only wait up to a given timeout.
func (c *Cond) WaitWithTimeout(t time.Duration) {
	n := c.NotifyChan()
	c.L.Unlock()
	select {
	case <-n:
	case <-time.After(t):
	}
	c.L.Lock()
}

// Returns a channel that can be used to wait for next Broadcast() call.
func (c *Cond) NotifyChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

// Broadcast call notifies everyone that something has changed.
func (c *Cond) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	close(*(*chan struct{})(ptrOld))
}

func TestRace() {
	x := 0
	c := NewCond(&sync.Mutex{})
	done := make(chan bool)
	go func() {
		log.Println("go1")
		c.L.Lock()
		log.Println("go1 locked")
		x = 1
		c.Wait()
		if x != 2 {
			log.Fatal("want 2")
		}
		x = 3
		c.Broadcast()
		c.L.Unlock()
		log.Println("go1 ends")
		done <- true
	}()
	go func() {
		log.Println("go2")
		c.L.Lock()
		log.Println("go2 locked")
		for {
			if x == 1 {
				x = 2
				c.Broadcast()
				break
			}
			c.L.Unlock()
			runtime.Gosched()
			c.L.Lock()
		}
		c.L.Unlock()
		log.Println("go2 ends")
		done <- true
	}()
	go func() {
		log.Println("go3")
		c.L.Lock()
		log.Println("go3 locked")
		for {
			if x == 2 {
				c.Wait()
				if x != 3 {
					log.Fatal("want 3")
				}
				break
			}
			if x == 3 {
				break
			}
			c.L.Unlock()
			runtime.Gosched()
			c.L.Lock()
		}
		c.L.Unlock()
		log.Println("go3 ends")
		done <- true
	}()
	<-done
	<-done
	<-done
}

func main2() {
	log.Println("testing...")
	TestRace()
}
