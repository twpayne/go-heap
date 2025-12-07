package heap_test

import (
	"context"
	"strconv"
	"testing"
	"testing/synctest"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-heap"
)

func TestPriorityOneValue(t *testing.T) {
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
		return a < b
	})

	inCh <- 1
	assert.Equal(t, 1, <-outCh)

	close(inCh)
	_, ok := <-outCh
	assert.False(t, ok)
}

func TestPriorityReorderValues(t *testing.T) {
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
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
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
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
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
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
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
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

func TestBufferedPriorityChannel(t *testing.T) {
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.BufferedPriorityChannel(ctx, inCh, 4, func(a, b int) bool {
		return a < b
	})

	go func() {
		defer close(inCh)
		for i := 9; i >= 0; i-- {
			inCh <- i
		}
	}()

	result := make([]int, 0, 10)
	for value := range outCh {
		result = append(result, value)
	}

	assert.Equal(t, []int{6, 5, 4, 3, 2, 1, 0, 7, 8, 9}, result)
}

func TestBufferedPriorityChannelInvalidSize(t *testing.T) {
	assert.Panics(t, func() {
		_ = heap.BufferedPriorityChannel[int](context.Background(), nil, -1, nil)
	})
}

func TestBufferedPriorityChannelNeverFull(t *testing.T) {
	ctx := context.Background()

	inCh := make(chan int)
	outCh := heap.BufferedPriorityChannel(ctx, inCh, 4, func(a, b int) bool {
		return a < b
	})

	go func() {
		defer close(inCh)
		for i := 2; i >= 0; i-- {
			inCh <- i
		}
	}()

	result := make([]int, 0, 2)
	for value := range outCh {
		result = append(result, value)
	}

	assert.Equal(t, []int{0, 1, 2}, result)
}

func TestBufferedPriorityChannelCancelDuringFill(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	inCh := make(chan int)
	outCh := heap.BufferedPriorityChannel(ctx, inCh, 4, func(a, b int) bool {
		return a < b
	})

	inCh <- 2
	inCh <- 1
	inCh <- 0
	cancel()

	_, ok := <-outCh
	assert.False(t, ok)
}

func TestBufferedPriorityChannelCancelAfterFill(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	inCh := make(chan int)
	outCh := heap.BufferedPriorityChannel(ctx, inCh, 4, func(a, b int) bool {
		return a < b
	})

	inCh <- 3
	inCh <- 2
	inCh <- 1
	inCh <- 0
	close(inCh)

	value, ok := <-outCh
	assert.True(t, ok)
	assert.Equal(t, 0, value)

	cancel()
}

func TestBufferedPriorityChannelCancelDuringOperation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	inCh := make(chan int)
	outCh := heap.BufferedPriorityChannel(ctx, inCh, 4, func(a, b int) bool {
		return a < b
	})

	go func() {
		defer close(inCh)
		for i := 9; i >= 0; i-- {
			select {
			case <-ctx.Done():
				return
			case inCh <- i:
			}
		}
	}()

	assert.Equal(t, 6, <-outCh)
	assert.Equal(t, 5, <-outCh)

	cancel()
}

func TestPriorityChannelImmediateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	inCh := make(chan int)
	outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
		return a < b
	})

	cancel()

	_, ok := <-outCh
	assert.False(t, ok)
}

func TestPriorityChannelCancel(t *testing.T) {
	for n := range 5 {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())

				inCh := make(chan int)
				outCh := heap.PriorityChannel(ctx, inCh, func(a, b int) bool {
					return a < b
				})

				go func() {
					defer close(inCh)
					for i := range n {
						inCh <- i + 1
					}
				}()

				synctest.Wait()
				cancel()

				_, ok := <-outCh
				assert.False(t, ok)
			})
		})
	}
}

func TestBufferedPriorityChannelCancel(t *testing.T) {
	for n := range 5 {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())

				inCh := make(chan int)
				outCh := heap.BufferedPriorityChannel(ctx, inCh, 3, func(a, b int) bool {
					return a < b
				})

				go func() {
					defer close(inCh)
					for i := range n {
						select {
						case <-ctx.Done():
							return
						case inCh <- i + 1:
						}
					}
				}()

				synctest.Wait()
				cancel()

				_, ok := <-outCh
				assert.False(t, ok)
			})
		})
	}
}

func TestBufferedPriorityChannelDrain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	inCh := make(chan int)
	outCh := heap.BufferedPriorityChannel(ctx, inCh, 3, func(a, b int) bool {
		return a < b
	})

	inCh <- 3
	inCh <- 2
	inCh <- 1
	close(inCh)

	assert.Equal(t, 1, <-outCh)
	assert.Equal(t, 2, <-outCh)
	assert.Equal(t, 3, <-outCh)
	_, ok := <-outCh
	assert.False(t, ok)
}
