package heap

// PriorityChannel greedily reads values from inCh and returns a channel that
// returns the same values prioritized according to lessFunc, with lesser values
// returned first. When inCh is closed, all remaining values are written to the
// returned channel, and then the returned channel is closed.
func PriorityChannel[T any](inCh <-chan T, lessFunc func(T, T) bool) <-chan T {
	// FIXME add channel sizes (i.e. min and max heap size)

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
					for value := range heap.All() {
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
