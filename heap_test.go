package heap_test

import (
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-heap"
)

func TestEmpty(t *testing.T) {
	t.Parallel()

	h := heap.NewOrderedHeap[int]()

	assert.True(t, h.Empty())
	assert.Zero(t, h.Len())

	v, ok := h.Peek()
	assert.False(t, ok)
	assert.Zero(t, v)

	v, ok = h.Pop()
	assert.False(t, ok)
	assert.Zero(t, v)
}

func TestOneElement(t *testing.T) {
	t.Parallel()

	h := heap.NewOrderedHeap[string]()

	h.Push("a")

	assert.False(t, h.Empty())
	assert.Equal(t, 1, h.Len())

	v, ok := h.Peek()
	assert.True(t, ok)
	assert.Equal(t, "a", v)

	v, ok = h.Pop()
	assert.True(t, ok)
	assert.Equal(t, "a", v)

	assert.Zero(t, h.Len())
}

func TestTwoElements(t *testing.T) {
	t.Parallel()

	h := heap.NewOrderedHeap[float64]()

	h.Push(2)
	h.Push(1)

	assert.False(t, h.Empty())
	assert.Equal(t, 2, h.Len())

	for _, expected := range []float64{1, 2} {
		v, ok := h.Pop()
		assert.True(t, ok)
		assert.Equal(t, expected, v)
	}
}

func TestThreeElements(t *testing.T) {
	t.Parallel()

	h := heap.NewOrderedHeap[uint]()

	h.PushMany(3, 2, 1)

	assert.False(t, h.Empty())
	assert.Equal(t, 3, h.Len())

	for _, expected := range []uint{1, 2, 3} {
		v, ok := h.Pop()
		assert.True(t, ok)
		assert.Equal(t, expected, v)
	}
}

func TestRandomPermutations(t *testing.T) {
	t.Parallel()

	const N = 1024
	r := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < N; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			n := r.IntN(N)
			values := make([]int, 0, n)
			for i := 0; i < n; i++ {
				values = append(values, r.IntN(N))
			}
			expected := slices.Clone(values)
			slices.Sort(expected)

			t.Run("forward", func(t *testing.T) {
				h := heap.NewOrderedHeap[int]().Grow(n)
				for _, value := range values {
					h.Push(value)
				}
				assert.Equal(t, expected, h.PopAll())
			})

			slices.Reverse(expected)
			t.Run("reverse", func(t *testing.T) {
				h := heap.NewReverseOrderedHeap[int]().Grow(n)
				for _, value := range values {
					h.Push(value)
				}
				assert.Equal(t, expected, h.PopAll())
			})
		})
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	const N = 1024
	r := rand.New(rand.NewPCG(1, 2))
	for i := 0; i < N; i++ {
		n := r.IntN(32)
		h := heap.NewOrderedHeap[int]().Grow(n)
		values := make([]int, 0, n)
		for i := 0; i < n; i++ {
			value := r.IntN(N)
			values = append(values, value)
		}
		h.Set(values)
		expected := slices.Clone(values)
		slices.Sort(expected)
		assert.Equal(t, expected, h.PopAll())
	}
}

func TestPushPop(t *testing.T) {
	h := heap.NewOrderedHeap[int]()
	assert.Equal(t, 1, h.PushPop(1))
	h.Push(2)
	assert.Equal(t, 1, h.PushPop(1))
	assert.Equal(t, 2, h.PushPop(3))
	h.PushMany(5, 4, 6)
	assert.Equal(t, []int{3, 4, 5, 6}, h.PopAll())
}

func TestCapAndClip(t *testing.T) {
	h := heap.NewOrderedHeap[byte]()
	assert.Zero(t, h.Cap())
	h.Grow(16)
	assert.Equal(t, 16, h.Cap())
	h.Push(1)
	h.Clip()
	assert.Equal(t, 1, h.Cap())
}

func TestMustPop(t *testing.T) {
	h := heap.NewOrderedHeap[int]()
	h.Push(1)
	assert.NotPanics(t, func() {
		assert.Equal(t, 1, h.MustPop())
	})
	assert.Panics(t, func() {
		_ = h.MustPop()
	})
}

func TestHeapFloat64s(t *testing.T) {
	h := heap.NewOrderedHeap[float64]()
	h.PushMany(2, math.NaN(), 1, math.Copysign(0, -1), math.Copysign(0, 1), math.Inf(1), math.Inf(-1))
	assert.True(t, math.IsNaN(h.MustPop()))
	assert.Equal(t, math.Inf(-1), h.MustPop())
	assert.Equal(t, math.Copysign(0, 1), h.MustPop())
	assert.Equal(t, math.Copysign(0, -1), h.MustPop())
	assert.Equal(t, 1, h.MustPop())
	assert.Equal(t, 2, h.MustPop())
	assert.Equal(t, math.Inf(1), h.MustPop())
}
