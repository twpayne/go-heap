package heap_test

import (
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-heap"
)

func TestPriorityOneValue(t *testing.T) {
	destCh := make(chan int)
	sourceCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		heap.PriorityChannel(destCh, sourceCh, func(a, b int) bool {
			return a < b
		})
	}()

	sourceCh <- 1
	assert.Equal(t, 1, <-destCh)

	close(sourceCh)
	wg.Wait()
}

func TestPriorityReorderValues(t *testing.T) {
	destCh := make(chan int)
	sourceCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		heap.PriorityChannel(destCh, sourceCh, func(a, b int) bool {
			return a < b
		})
	}()

	for i := 9; i >= 0; i-- {
		sourceCh <- i
	}

	for i := range 10 {
		assert.Equal(t, i, <-destCh)
	}

	close(sourceCh)
	wg.Wait()
}

func TestPriorityMixedValueOrder(t *testing.T) {
	destCh := make(chan int)
	sourceCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		heap.PriorityChannel(destCh, sourceCh, func(a, b int) bool {
			return a < b
		})
	}()

	sourceCh <- 8
	sourceCh <- 6
	sourceCh <- 4
	sourceCh <- 2
	sourceCh <- 0
	assert.Equal(t, 0, <-destCh)
	sourceCh <- 7
	assert.Equal(t, 2, <-destCh)
	sourceCh <- 5
	assert.Equal(t, 4, <-destCh)
	sourceCh <- 3
	assert.Equal(t, 3, <-destCh)
	sourceCh <- 1
	assert.Equal(t, 1, <-destCh)
	assert.Equal(t, 5, <-destCh)
	assert.Equal(t, 6, <-destCh)
	assert.Equal(t, 7, <-destCh)
	assert.Equal(t, 8, <-destCh)

	close(sourceCh)
	wg.Wait()
}

func TestPriorityChannelCloseSourceBeforeReading(t *testing.T) {
	destCh := make(chan int)
	sourceCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		heap.PriorityChannel(destCh, sourceCh, func(a, b int) bool {
			return a < b
		})
	}()

	for i := 9; i >= 0; i-- {
		sourceCh <- i
	}
	close(sourceCh)

	for i := range 10 {
		assert.Equal(t, i, <-destCh)
	}

	wg.Wait()
}

func TestPriorityChannelCloseSourceDuringRead(t *testing.T) {
	destCh := make(chan int)
	sourceCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		heap.PriorityChannel(destCh, sourceCh, func(a, b int) bool {
			return a < b
		})
	}()

	for i := 9; i > 4; i-- {
		sourceCh <- i
	}
	assert.Equal(t, 5, <-destCh)
	assert.Equal(t, 6, <-destCh)
	for i := 4; i >= 0; i-- {
		sourceCh <- i
	}
	close(sourceCh)
	assert.Equal(t, 0, <-destCh)
	assert.Equal(t, 1, <-destCh)
	assert.Equal(t, 2, <-destCh)
	assert.Equal(t, 3, <-destCh)
	assert.Equal(t, 4, <-destCh)
	assert.Equal(t, 7, <-destCh)
	assert.Equal(t, 8, <-destCh)
	assert.Equal(t, 9, <-destCh)

	wg.Wait()
}
