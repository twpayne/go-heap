//go:build rangefunc

package heap_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-heap"
)

func TestHeap_All(t *testing.T) {
	h := heap.NewOrderedHeap[int]().PushMany(4, 2, 1, 0)
	var actual []int
	for value := range h.All() {
		switch value {
		case 1:
			h.Push(3)
		case 2:
			h.Push(5)
		}
		actual = append(actual, value)
	}
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, actual)
}
