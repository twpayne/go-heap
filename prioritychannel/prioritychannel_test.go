package prioritychannel_test

import (
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-heap/prioritychannel"
)

func TestSimple(t *testing.T) {
	c := prioritychannel.NewPriorityChannel(
		prioritychannel.WithLessFunc(func(a, b int) bool {
			return a < b
		}),
	)
	destCh := make(chan int)
	sourceCh := make(chan int)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Run(destCh, sourceCh)
	}()

	sourceCh <- 1
	assert.Equal(t, 1, <-destCh)

	close(sourceCh)
	wg.Wait()
}
