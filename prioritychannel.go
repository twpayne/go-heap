package heap

// PriorityChannel greedily reads values from inCh and returns a channel that
// returns the same values prioritized according to lessFunc, with lesser values
// returned first. When inCh is closed, all remaining values are written to the
// returned channel, and then the returned channel is closed.
func PriorityChannel[T any](inCh <-chan T, lessFunc func(T, T) bool) <-chan T {
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
					valueToSend, ok = <-inCh
					if !ok {
						// inCh was closed so we are done.
						return
					}
				}
				valueToSendValid = true //nolint:wastedassign
			}

			// Either write valueToSend to outCh or read a new value from inCh
			// and update valueToSend.
			select {
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
						outCh <- value
					}
					return
				}

				// Otherwise, add value to the heap and get the new value to send.
				valueToSend = heap.PushPop(value)
				valueToSendValid = true
			}
		}
	}()

	return outCh
}

// BufferedPriorityChannel reads values from inCh and returns a channel that
// returns the same values prioritized according to lessFunc, with lesser values
// returned first. It maintains a buffer of size size, reading from inCh until
// the buffer is full, and then returning the values in priority over the
// returned channel, and reading more values from inCh when required. When inCh
// is closed, all remaining values are written to the returned channel, and then
// the returned channel is closed.
func BufferedPriorityChannel[T any](inCh <-chan T, size int, lessFunc func(T, T) bool) <-chan T {
	if size <= 0 {
		panic("size out of range")
	}

	outCh := make(chan T)

	go func() {
		defer close(outCh)
		heap := NewHeap(lessFunc)

		// Pre-fill the heap with up to size values.
		for range size - heap.Len() {
			value, ok := <-inCh
			if !ok {
				// inCh was closed so we are done.
				goto DONE
			}
			heap.Push(value)
		}

		// FIXME use heap.PushPop

		for {
			// Send the least value.
			value, _ := heap.Pop()
			outCh <- value

			// Read another value.
			if value, ok := <-inCh; ok {
				heap.Push(value)
			} else {
				// inCh was closed so we are done.
				goto DONE
			}
		}

	DONE:
		// Write all remaining values.
		for value := range heap.PopAll() {
			outCh <- value
		}
	}()

	return outCh
}
