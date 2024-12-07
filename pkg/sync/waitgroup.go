package sync

import (
	"sync"
	"time"
)

func NewWaitGroup(size int) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(size)
	return &wg
}

func WaitFor(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-done:
		return false
	case <-time.After(timeout):
		return true
	}
}
