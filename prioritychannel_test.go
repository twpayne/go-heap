package heap_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-heap"
)

func TestPriorityOneValue(t *testing.T) {
	inCh := make(chan int)
	outCh := heap.PriorityChannel(inCh, func(a, b int) bool {
		return a < b
	})

	inCh <- 1
	assert.Equal(t, 1, <-outCh)

	close(inCh)
	_, ok := <-outCh
	assert.False(t, ok)
}

func TestPriorityReorderValues(t *testing.T) {
	inCh := make(chan int)
	outCh := heap.PriorityChannel(inCh, func(a, b int) bool {
		return a < b
	})

	for i := 9; i >= 0; i-- {
		inCh <- i
	}

	for i := range 10 {
		assert.Equal(t, i, <-outCh)
	}

	close(inCh)

	_, ok := <-outCh
	assert.False(t, ok)
}

func TestPriorityMixedValueOrder(t *testing.T) {
	inCh := make(chan int)
	outCh := heap.PriorityChannel(inCh, func(a, b int) bool {
		return a < b
	})

	inCh <- 8
	inCh <- 6
	inCh <- 4
	inCh <- 2
	inCh <- 0
	assert.Equal(t, 0, <-outCh)
	inCh <- 7
	assert.Equal(t, 2, <-outCh)
	inCh <- 5
	assert.Equal(t, 4, <-outCh)
	inCh <- 3
	assert.Equal(t, 3, <-outCh)
	inCh <- 1
	assert.Equal(t, 1, <-outCh)
	assert.Equal(t, 5, <-outCh)
	assert.Equal(t, 6, <-outCh)
	assert.Equal(t, 7, <-outCh)
	assert.Equal(t, 8, <-outCh)

	close(inCh)
	_, ok := <-outCh
	assert.False(t, ok)
}

func TestPriorityChannelCloseSourceBeforeReading(t *testing.T) {
	inCh := make(chan int)
	outCh := heap.PriorityChannel(inCh, func(a, b int) bool {
		return a < b
	})

	for i := 9; i >= 0; i-- {
		inCh <- i
	}
	close(inCh)

	for i := range 10 {
		assert.Equal(t, i, <-outCh)
	}

	_, ok := <-outCh
	assert.False(t, ok)
}

func TestPriorityChannelCloseSourceDuringRead(t *testing.T) {
	inCh := make(chan int)
	outCh := heap.PriorityChannel(inCh, func(a, b int) bool {
		return a < b
	})

	for i := 9; i > 4; i-- {
		inCh <- i
	}
	assert.Equal(t, 5, <-outCh)
	assert.Equal(t, 6, <-outCh)
	for i := 4; i >= 0; i-- {
		inCh <- i
	}
	close(inCh)
	assert.Equal(t, 0, <-outCh)
	assert.Equal(t, 1, <-outCh)
	assert.Equal(t, 2, <-outCh)
	assert.Equal(t, 3, <-outCh)
	assert.Equal(t, 4, <-outCh)
	assert.Equal(t, 7, <-outCh)
	assert.Equal(t, 8, <-outCh)
	assert.Equal(t, 9, <-outCh)

	_, ok := <-outCh
	assert.False(t, ok)
}
