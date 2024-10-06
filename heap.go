// Package heaps implements a generic heap data structure.
package heap

import (
	"cmp"
	"iter"
	"slices"
)

// A Heap is a heap of Ts.
type Heap[T any] struct {
	lessFunc func(T, T) bool
	values   []T
}

// NewHeap returns a new heap that uses lessFunc to compare elements.
func NewHeap[T any](lessFunc func(T, T) bool) *Heap[T] {
	return &Heap[T]{
		lessFunc: lessFunc,
	}
}

// NewOrderedHeap returns a new heap that operates on [cmp.Ordered] elements.
func NewOrderedHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeap(cmp.Less[T])
}

// NewReverseOrderedHeap returns a new heap that operates on [cmp.Ordered]
// elements in reverse order.
func NewReverseOrderedHeap[T cmp.Ordered]() *Heap[T] {
	return NewHeap(func(a, b T) bool {
		return cmp.Less[T](b, a)
	})
}

// Cap returns the underlying capacity of h.
func (h *Heap[T]) Cap() int {
	return cap(h.values)
}

// Clip removes unused capacity from h.
func (h *Heap[T]) Clip() *Heap[T] {
	h.values = slices.Clip(h.values)
	return h
}

// Empty returns whether h is empty in O(1) time and memory.
func (h *Heap[T]) Empty() bool {
	return len(h.values) == 0
}

// Grow increases h's capacity by at least n.
func (h *Heap[T]) Grow(n int) *Heap[T] {
	h.values = slices.Grow(h.values, n)
	return h
}

// Len returns the size of h in O(1) time and memory.
func (h *Heap[T]) Len() int {
	return len(h.values)
}

// MustPop returns the lowest value in h. It panics if h is empty.
func (h *Heap[T]) MustPop() T {
	value, ok := h.Pop()
	if !ok {
		panic("empty heap")
	}
	return value
}

// Peek returns the lowest value in h in O(1) time and memory, without removing
// it, and whether it exists.
func (h *Heap[T]) Peek() (T, bool) {
	if h.Empty() {
		var zero T
		return zero, false
	}
	return h.values[0], true
}

// Pop returns the lowest value in h, removing it, and whether it exists in O(N)
// time and O(1) memory.
func (h *Heap[T]) Pop() (T, bool) {
	switch n := len(h.values); n {
	case 0:
		var zero T
		return zero, false
	case 1:
		value := h.values[0]
		h.values = h.values[:0] // Truncate values instead of setting values to nil to reduce GC pressure.
		return value, true
	default:
		value := h.values[0]
		h.values[0] = h.values[n-1]
		h.values = h.values[:n-1]
		h.siftDown(0)
		return value, true
	}
}

// PopAll returns an iterator that pops all values.
func (h *Heap[T]) PopAll() iter.Seq[T] {
	return func(yield func(T) bool) {
		for value, ok := h.Pop(); ok && yield(value); value, ok = h.Pop() { //nolint:revive
		}
	}
}

// Push adds value to h in amortized O(N) time.
func (h *Heap[T]) Push(value T) *Heap[T] {
	h.values = append(h.values, value)
	h.siftUp(len(h.values) - 1)
	return h
}

// PushMany pushes multiple values onto the heap.
func (h *Heap[T]) PushMany(values ...T) *Heap[T] {
	n := len(h.values)
	h.values = append(h.values, values...)
	for i := range values {
		h.siftUp(n + i)
	}
	return h
}

// PushPop pushes value onto the heap and then pops the lowest value off the
// heap and returns it in O(N) time. It is slightly more efficient than separate
// calls to [Heap.Push] and [Heap.Pop].
func (h *Heap[T]) PushPop(value T) T {
	if len(h.values) == 0 || h.lessFunc(value, h.values[0]) {
		return value
	}
	value, h.values[0] = h.values[0], value
	h.siftDown(0)
	return value
}

// Set sets the values on h to be values in amortized O(N) time. h takes
// ownership of values.
func (h *Heap[T]) Set(values []T) *Heap[T] {
	h.values = values
	for index := len(values) / 2; index >= 0; index-- {
		h.siftDown(index)
	}
	return h
}

// siftDown implements the sift down operation, moving the element at index
// towards the leaves.
func (h *Heap[T]) siftDown(index int) {
	for n := len(h.values); index < n/2; {
		leftChildIndex, rightChildIndex := 2*index+1, 2*index+2
		smallestChildIndex := leftChildIndex
		if rightChildIndex < n && h.lessFunc(h.values[rightChildIndex], h.values[leftChildIndex]) {
			smallestChildIndex = rightChildIndex
		}
		if h.lessFunc(h.values[index], h.values[smallestChildIndex]) {
			return
		}
		h.values[index], h.values[smallestChildIndex] = h.values[smallestChildIndex], h.values[index]
		index = smallestChildIndex
	}
}

// siftUp implements the sift up operation, moving the element at index towards
// the root.
func (h *Heap[T]) siftUp(index int) {
	for index > 0 {
		parentIndex := (index - 1) / 2
		if h.lessFunc(h.values[parentIndex], h.values[index]) {
			return
		}
		h.values[index], h.values[parentIndex] = h.values[parentIndex], h.values[index]
		index = parentIndex
	}
}
