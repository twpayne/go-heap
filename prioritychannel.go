package heap

import "context"

// PriorityChannel greedily reads values from inCh and returns a channel that
// returns the same values prioritized according to lessFunc, with lesser values
// returned first.
//
// When inCh is closed, all remaining values are written to the returned
// channel, and then the returned channel is closed.
//
// If ctx is canceled then the returned channel is closed immediately.
func PriorityChannel[T any](ctx context.Context, inCh <-chan T, lessFunc func(T, T) bool) <-chan T {
	outCh := make(chan T)

	go func() {
		defer close(outCh)
		heap := NewHeap(lessFunc)

		var valueToSend T
		valueToSendValid := false

		for {
			// If we do not already have a value to send, get one. If the heap
			// is empty then read one from inCh, otherwise chose the highest
			// priority value from heap.
			if !valueToSendValid {
				var ok bool
				valueToSend, ok = heap.Pop()
				if !ok {
					select {
					case <-ctx.Done():
						return
					case valueToSend, ok = <-inCh:
						if !ok {
							// inCh was closed so we are done.
							return
						}
					}
				}
				valueToSendValid = true //nolint:wastedassign
			}

			// Either write valueToSend to outCh or read a new value from inCh
			// and update valueToSend.
			select {
			case <-ctx.Done():
				return
			case outCh <- valueToSend:
				// As valueToSend was sent, we need a new one.
				valueToSendValid = false
			case value, ok := <-inCh:
				// As valueToSend was not sent, push it back onto the heap.
				heap.Push(valueToSend)

				// If inCh was closed then send the remaining values to outCh
				// and return.
				if !ok {
					for value := range heap.PopAll() {
						select {
						case <-ctx.Done():
							return
						case outCh <- value:
						}
					}
					return
				}

				// Otherwise, add value to the heap and get the new value to
				// send.
				valueToSend = heap.PushPop(value)
				valueToSendValid = true
			}
		}
	}()

	return outCh
}

// BufferedPriorityChannel reads values from inCh and returns a channel that
// returns the same values prioritized according to lessFunc, with lesser values
// returned first.
//
// It maintains a buffer of size size, reading from inCh until the buffer is
// full, and then returning the values in priority over the returned channel,
// and reading more values from inCh when required. When inCh is closed, all
// remaining values are written to the returned channel, and then the returned
// channel is closed.
//
// If ctx is canceled then the returned channel is closed immediately.
func BufferedPriorityChannel[T any](ctx context.Context, inCh <-chan T, size int, lessFunc func(T, T) bool) <-chan T {
	if size <= 0 {
		panic("size out of range")
	}

	outCh := make(chan T)

	go func() {
		defer close(outCh)
		heap := NewHeap(lessFunc)
		var leastValue T

		// Pre-fill the heap with up to size values.
		for range size {
			select {
			case <-ctx.Done():
				return
			case value, ok := <-inCh:
				if !ok {
					goto DRAIN
				}
				heap.Push(value)
			}
		}

		// Prepare the least value to send.
		leastValue, _ = heap.Pop()

		// Main loop.
		for {
			// Send the least value.
			select {
			case <-ctx.Done():
				return
			case outCh <- leastValue:
			}

			// Read the next value from inCh and update the heap and least
			// value.
			select {
			case <-ctx.Done():
				return
			case value, ok := <-inCh:
				if !ok {
					goto DRAIN
				}
				leastValue = heap.PushPop(value)
			}
		}

	DRAIN:
		// inCh was closed so we are done. Write all remaining values and
		// return.
		for value := range heap.PopAll() {
			select {
			case <-ctx.Done():
				return
			case outCh <- value:
			}
		}
	}()

	return outCh
}
