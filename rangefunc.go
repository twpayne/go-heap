//go:build rangefunc

package heap

import "iter"

func (h *Heap[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for value, ok := h.Pop(); ok && yield(value); value, ok = h.Pop() {
		}
	}
}
